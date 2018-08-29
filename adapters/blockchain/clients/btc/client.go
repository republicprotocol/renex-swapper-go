package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/keystore"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/network"
)

type Conn struct {
	URL         string
	ChainParams *chaincfg.Params
	Network     string
}

func Connect(networkConfig network.Config) (Conn, error) {
	connParams := networkConfig.GetBitcoinNetwork()
	return ConnectWithParams(connParams.Network, connParams.URL)
}

func ConnectWithParams(chain, url string) (Conn, error) {
	var chainParams *chaincfg.Params

	switch chain {
	case "regtest":
		chainParams = &chaincfg.RegressionNetParams
	case "testnet":
		chainParams = &chaincfg.TestNet3Params
	default:
		chainParams = &chaincfg.MainNetParams
	}

	return Conn{
		URL:         url,
		ChainParams: chainParams,
		Network:     chain,
	}, nil
}

func (conn *Conn) PromptPublishTx(tx *wire.MsgTx, name string) (*chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := tx.Serialize(buf); err != nil {
		return nil, err
	}
	stx := hex.EncodeToString(buf.Bytes())
	if err := conn.SubmitSignedTransaction(stx); err != nil {
		return nil, err
	}
	hash := tx.TxHash()
	return &hash, nil
}

func walletPort(params *chaincfg.Params) string {
	switch params {
	case &chaincfg.MainNetParams:
		return "8332"
	case &chaincfg.TestNet3Params:
		return "18332"
	case &chaincfg.RegressionNetParams:
		return "18443"
	default:
		return ""
	}
}

func (conn *Conn) SignTransaction(tx *wire.MsgTx, key keystore.Key) (*wire.MsgTx, bool, error) {
	addr, err := key.GetAddress()
	if err != nil {
		return nil, false, fmt.Errorf("your keystore is "+
			"malformed %v", conn.ChainParams.Name)
	}

	myAddr, err := btcutil.DecodeAddress(string(addr), conn.ChainParams)
	if err != nil {
		return nil, false, fmt.Errorf("your address is not "+
			"intended for use on %v", conn.ChainParams.Name)
	}

	var value int64

	for _, j := range tx.TxOut {
		value = value + j.Value
	}

	unspentValue, err := conn.Balance(string(addr))
	if err != nil {
		return nil, false, err
	}

	utxos, err := conn.GetUnspentOutputs(string(addr))
	if err != nil {
		return nil, false, err
	}

	if value > unspentValue {
		return nil, false, fmt.Errorf("Not enough balance"+
			"required:%d current:%d", value, unspentValue)
	}

	for _, j := range utxos.Outputs {
		if value <= 0 {
			break
		}
		hashBytes, err := hex.DecodeString(j.TxID)
		if err != nil {
			return nil, false, err
		}
		hash, err := chainhash.NewHash(hashBytes)
		if err != nil {
			return nil, false, err
		}
		ScriptPubKey, err := hex.DecodeString(j.ScriptPubKey)
		if err != nil {
			return nil, false, err
		}
		tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(hash, j.Vout), ScriptPubKey, [][]byte{}))
		value = value - j.Amount
	}
	P2PKHScript, err := txscript.PayToAddrScript(myAddr)
	if err != nil {
		return nil, false, err
	}

	if value <= 0 {
		tx.AddTxOut(wire.NewTxOut(int64(-value)-10000, P2PKHScript))
	}

	privKey, err := key.GetKey()
	if err != nil {
		return nil, false, err
	}

	priKey := btcec.PrivateKey(*privKey)
	for i, txin := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, txin.SignatureScript, txscript.SigHashAll, &priKey, false)
		if err != nil {
			return nil, false, err
		}
		tx.TxIn[i].SignatureScript = sigScript
	}

	return tx, true, nil
}

func (conn *Conn) Balance(address string) (int64, error) {
	utxos, err := conn.GetUnspentOutputs(address)
	if err != nil {
		return -1, err
	}
	var balance int64
	for _, utxo := range utxos.Outputs {
		balance = balance + utxo.Amount
	}
	return balance, nil
}

// WaitTillMined waits for the transactions to be mined, and gets the given
// number of confirmations.
func (conn *Conn) WaitTillMined(txHash *chainhash.Hash, confirmations int64) error {
	for {
		mined, err := conn.Mined(txHash.String(), confirmations)
		if err != nil {
			return err
		}

		if mined {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}
