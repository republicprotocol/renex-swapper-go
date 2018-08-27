package btc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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
	data := url.Values{}
	data.Set("tx", stx)

	client := &http.Client{}
	r, err := http.NewRequest("POST", conn.URL+"/pushtx", strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = client.Do(r)
	if err != nil {
		return nil, err
	}

	hash := tx.TxHash()
	return &hash, nil
}

func normalizeAddress(addr string, defaultPort string) (hostport string, err error) {
	host, port, origErr := net.SplitHostPort(addr)
	if origErr == nil {
		return net.JoinHostPort(host, port), nil
	}
	addr = net.JoinHostPort(addr, defaultPort)
	_, _, err = net.SplitHostPort(addr)
	if err != nil {
		return "", origErr
	}
	return addr, nil
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
		return nil, false, fmt.Errorf("failed to decode your address: %v", err)
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

	unspentValue, err := conn.Balance(myAddr)
	if err != nil {
		return nil, false, err
	}

	utxos, err := conn.ListUnspent(myAddr)
	if err != nil {
		return nil, false, err
	}

	if value > unspentValue {
		return nil, false, fmt.Errorf("Not enough balance required:%d current:%d", value, unspentValue)
	}

	for _, j := range utxos.UnspentOutputs {
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

type Unspent struct {
	TxID         string `json:"tx_hash"`
	Vout         uint32 `json:"tx_output_n"`
	ScriptPubKey string `json:"script"`
	Amount       int64  `json:"value"`
}

type Unspents struct {
	UnspentOutputs []Unspent `json:"unspent_outputs"`
}

func (conn *Conn) ListUnspent(address btcutil.Address) (Unspents, error) {
	resp, err := http.Get(fmt.Sprintf(conn.URL + "/unspent?active=" + address.EncodeAddress()))
	if err != nil {
		return Unspents{}, err
	}
	defer resp.Body.Close()
	utxoBytes, err := ioutil.ReadAll(resp.Body)
	utxos := Unspents{}
	json.Unmarshal(utxoBytes, &utxos)
	return utxos, nil
}

func (conn *Conn) Balance(address btcutil.Address) (int64, error) {
	utxos, err := conn.ListUnspent(address)
	if err != nil {
		return -1, err
	}
	var balance int64
	for _, utxo := range utxos.UnspentOutputs {
		balance = balance + utxo.Amount
	}
	return balance, nil
}

// WaitTillMined doesnot wait for the transactions to be mined, can be updated
// later
func (conn *Conn) WaitTillMined(txHash *chainhash.Hash, confirmations int64) error {
	type tx struct {
		BlockHeight int64 `json:"block_height"`
	}

	type block struct {
		Height int64 `json:"height"`
	}

	if confirmations <= 0 {
		return nil
	}
	confirmations = confirmations - 1
	for {
		resp, err := http.Get(fmt.Sprintf(conn.URL + "/rawtx/" + txHash.String()))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		txBytes, err := ioutil.ReadAll(resp.Body)
		transaction := tx{}
		json.Unmarshal(txBytes, &transaction)

		resp2, err := http.Get(fmt.Sprintf(conn.URL + "/latestblock"))
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		blockBytes, err := ioutil.ReadAll(resp2.Body)
		blockDetails := block{}
		json.Unmarshal(blockBytes, &blockDetails)

		if transaction.BlockHeight != 0 && blockDetails.Height-transaction.BlockHeight >= confirmations {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
