package btc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	rpc "github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/keystore"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/network"
)

type Conn struct {
	Client      *rpc.Client
	ChainParams *chaincfg.Params
	Network     string
}

func Connect(networkConfig network.Config) (Conn, error) {
	connParams := networkConfig.GetBitcoinNetwork()
	return ConnectWithParams(connParams.Network, connParams.URL, connParams.User, connParams.Password)
}

func ConnectWithParams(chain, url, user, password string) (Conn, error) {
	var chainParams *chaincfg.Params
	var connect string
	var err error

	switch chain {
	case "regtest":
		chainParams = &chaincfg.RegressionNetParams
	case "testnet":
		chainParams = &chaincfg.TestNet3Params
	default:
		chainParams = &chaincfg.MainNetParams
	}

	if url == "" {
		connect, err = normalizeAddress("localhost", walletPort(chainParams))
		if err != nil {
			return Conn{}, fmt.Errorf("wallet server address: %v", err)
		}
	} else {
		connect = url
	}

	connConfig := &rpc.ConnConfig{
		Host:         connect,
		User:         user,
		Pass:         password,
		DisableTLS:   true,
		HTTPPostMode: true,
	}

	rpcClient, err := rpc.New(connConfig, nil)
	if err != nil {
		return Conn{}, fmt.Errorf("rpc connect: %v", err)
	}

	// Should call the following after this function:
	/*
		defer func() {
			rpcClient.Shutdown()
			pcClient.WaitForShutdown()
		}()
	*/

	return Conn{
		Client:      rpcClient,
		ChainParams: chainParams,
		Network:     chain,
	}, nil
}

func (conn *Conn) FundRawTransaction(tx *wire.MsgTx) (fundedTx *wire.MsgTx, err error) {
	var buf bytes.Buffer
	buf.Grow(tx.SerializeSize())
	tx.Serialize(&buf)
	param0, err := json.Marshal(hex.EncodeToString(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	params := []json.RawMessage{param0}
	rawResp, err := conn.Client.RawRequest("fundrawtransaction", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Hex       string  `json:"hex"`
		Fee       float64 `json:"fee"`
		ChangePos float64 `json:"changepos"`
	}
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		return nil, err
	}
	fundedTxBytes, err := hex.DecodeString(resp.Hex)
	if err != nil {
		return nil, err
	}
	fundedTx = &wire.MsgTx{}
	err = fundedTx.Deserialize(bytes.NewReader(fundedTxBytes))
	if err != nil {
		return nil, err
	}
	return fundedTx, nil
}

func (conn *Conn) PromptPublishTx(tx *wire.MsgTx, name string) (*chainhash.Hash, error) {
	// FIXME: Transaction fees are set to high, change it before deploying to mainnet. By changing the booleon to false.
	txHash, err := conn.Client.SendRawTransaction(tx, true)
	if err != nil {
		return nil, fmt.Errorf("sendrawtransaction: %v", err)
	}
	return txHash, nil
}

func (conn *Conn) WaitForConfirmations(txHash *chainhash.Hash, requiredConfirmations int64) error {
	confirmations := int64(0)
	for confirmations < requiredConfirmations {
		txDetails, err := conn.Client.GetTransaction(txHash)
		if err != nil {
			return err
		}
		confirmations = txDetails.Confirmations

		// TODO: Base delay on chain config
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (conn *Conn) Shutdown() {
	conn.Client.Shutdown()
	conn.Client.WaitForShutdown()
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

type UTXO struct {
	Amount       float64
	ScriptPubKey string
	RedeemScript string
	TxID         string
	Vout         uint32
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

	var value, unspentValue float64
	for _, j := range tx.TxOut {
		value = value + float64(j.Value)
	}
	utxos, err := conn.Client.ListUnspentMinMaxAddresses(0, 99999, []btcutil.Address{myAddr})
	if err != nil {
		return nil, false, err
	}

	value = value / 100000000
	for _, utxo := range utxos {
		unspentValue = unspentValue + utxo.Amount
	}
	if value > unspentValue {
		return nil, false, fmt.Errorf("Not enough balance required:%f current:%f", value, unspentValue)
	}
	selectedTxIns := []btcjson.RawTxInput{}

	for _, j := range utxos {
		if value <= 0 {
			break
		}
		hashBytes, err := hex.DecodeString(j.TxID)
		if err != nil {
			return nil, false, err
		}
		hash, err := chainhash.NewHash(reverse(hashBytes))
		if err != nil {
			return nil, false, err
		}
		ScriptPubKey, err := hex.DecodeString(j.ScriptPubKey)
		if err != nil {
			return nil, false, err
		}
		tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(hash, j.Vout), ScriptPubKey, [][]byte{}))
		selectedTxIns = append(selectedTxIns, btcjson.RawTxInput{
			Txid:         j.TxID,
			Vout:         j.Vout,
			ScriptPubKey: j.ScriptPubKey,
			RedeemScript: j.RedeemScript,
		})
		value = value - j.Amount
	}

	P2PKHScript, err := txscript.PayToAddrScript(myAddr)
	if err != nil {
		return nil, false, err
	}

	if value <= 0 {
		tx.AddTxOut(wire.NewTxOut(int64(-value*100000000)-10000, P2PKHScript))
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

// func ListUnspent(address btcutil.Address) ([]UTXO) {

// 	{
// 		"tx_hash":"56e267a8ee056c88c26a57313c2689eeaadb5dae312a1f02c3a31035b90f420d",
// 		"tx_hash_big_endian":"0d420fb93510a3c3021f2a31ae5ddbaaee89263c31576ac2886c05eea867e256",
// 		"tx_index":241648797,
// 		"tx_output_n": 1,
// 		"script":"76a91448b5c6986b7bc6390bd1cc416154d1874fe116fd88ac",
// 		"value": 2385354313,
// 		"value_hex": "008e2d9e49",
// 		"confirmations":25612
// 	}

// https: //testnet.blockchain.info/unspent?active=
// }

func reverse(arr []byte) []byte {
	for i, j := 0, len(arr)-1; i < len(arr)/2; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
