package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/republicprotocol/renex-swapper-go/adapter/config"
	"github.com/republicprotocol/renex-swapper-go/adapter/keystore"
)

type Conn struct {
	URL         string
	ChainParams *chaincfg.Params
	Network     string
}

func Connect(conf config.Config) (Conn, error) {
	return ConnectWithParams(conf.Bitcoin.Network, conf.Bitcoin.URL)
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
	// TODO: Fees are set high, for faster testnet transactions, decrease them
	// before mainnet
	var fee = int64(500000)

	btcKey := key.(keystore.BitcoinKey)

	addrStr := btcKey.AddressString
	addr := btcKey.Address

	var value int64
	for _, j := range tx.TxOut {
		value = value + j.Value
	}
	value = value + fee

	unspentValue, err := conn.Balance(addrStr)
	if err != nil {
		return nil, false, err
	}

	if value > unspentValue {
		return nil, false, fmt.Errorf("Not enough balance"+
			"required:%d current:%d", value, unspentValue)
	}

	utxos, err := conn.GetUnspentOutputs(addrStr)
	if err != nil {
		return nil, false, err
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

	if value < 0 {
		P2PKHScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, false, err
		}
		tx.AddTxOut(wire.NewTxOut(int64(-value), P2PKHScript))
	}

	for i, txin := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, txin.SignatureScript, txscript.SigHashAll, btcKey.WIF.PrivKey, false)
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
