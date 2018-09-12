package btc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type RawTransaction struct {
	BlockHeight      int64    `json:"block_height"`
	VinSize          uint32   `json:"vin_sz"`
	VoutSize         uint32   `json:"vout_sz"`
	Version          uint8    `json:"ver"`
	TransactionHash  string   `json:"hash"`
	TransactionIndex uint64   `json:"tx_index"`
	Inputs           []Input  `json:"inputs"`
	Outputs          []Output `json:"out"`
}

type Input struct {
	PrevOut PreviousOut `json:"prev_out"`
	Script  string      `json:"script"`
}

type PreviousOut struct {
	TransactionHash  string `json:"hash"`
	Value            int64  `json:"value"`
	TransactionIndex uint64 `json:"tx_index"`
	VoutNumber       uint8  `json:"n"`
}

type Output struct {
	TransactionHash string `json:"hash"`
	Value           int64  `json:"value"`
	Script          string `json:"script"`
}

func (conn *Conn) GetRawTransaction(txhash string) (RawTransaction, error) {
	resp, err := http.Get(fmt.Sprintf(conn.URL + "/rawtx/" + txhash))
	if err != nil {
		return RawTransaction{}, err
	}
	defer resp.Body.Close()
	txBytes, err := ioutil.ReadAll(resp.Body)
	transaction := RawTransaction{}
	if err := json.Unmarshal(txBytes, &transaction); err != nil {
		return RawTransaction{}, err
	}
	return transaction, nil
}

type LatestBlock struct {
	BlockHash          string  `json:"hash"`
	Time               int64   `json:"time"`
	BlockIndex         int64   `json:"block_index"`
	Height             int64   `json:"height"`
	TransactionIndexes []int64 `json:"txIndexes"`
}

func (conn *Conn) Mined(txhash string, confirmations int64) (bool, error) {
	if confirmations <= 0 {
		return true, nil
	}

	confirmations = confirmations - 1
	latestBlock := LatestBlock{}

	resp, err := http.Get(fmt.Sprintf(conn.URL + "/latestblock"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	blockBytes, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(blockBytes, &latestBlock); err != nil {
		return false, err
	}

	tx, err := conn.GetRawTransaction(txhash)
	if err != nil {
		return false, err
	}

	if tx.BlockHeight != 0 {
		return true, nil
	}

	return false, nil
}

type UnspentOutput struct {
	TxID         string `json:"tx_hash"`
	Vout         uint32 `json:"tx_output_n"`
	ScriptPubKey string `json:"script"`
	Amount       int64  `json:"value"`
}

type UnspentOutputs struct {
	Outputs []UnspentOutput `json:"unspent_outputs"`
}

func (conn *Conn) GetUnspentOutputs(address string) (UnspentOutputs, error) {
	resp, err := http.Get(fmt.Sprintf(conn.URL + "/unspent?active=" + address + "&confirmations=6"))
	if err != nil {
		return UnspentOutputs{}, err
	}
	defer resp.Body.Close()
	utxoBytes, err := ioutil.ReadAll(resp.Body)
	utxos := UnspentOutputs{}
	json.Unmarshal(utxoBytes, &utxos)
	return utxos, nil
}

func (conn *Conn) SubmitSignedTransaction(stx string) error {
	fmt.Println("STX: ", stx)
	data := url.Values{}
	data.Set("tx", stx)
	client := &http.Client{}
	r, err := http.NewRequest("POST", conn.URL+"/pushtx", strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if _, err = client.Do(r); err != nil {
		return err
	}
	return nil
}

func (conn *Conn) GetRawTransactionsByAddress(addr string, offset, limit int) ([]RawTransaction, error) {
	return []RawTransaction{}, nil
}
