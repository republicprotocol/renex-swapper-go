package btc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/republicprotocol/renex-swapper-go/adapters/blockchain/clients/btc"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/keystore"
)

type AtomicSwapRequest struct {
	Conn         btc.Conn
	Key          keystore.Key
	PersonalAddr []byte
	ForeignAddr  []byte
	Fee          int64
	TxVersion    int32
}

func sign(tx *wire.MsgTx, idx int, pkScript []byte, key keystore.Key) (sig, pubkey []byte, err error) {
	privKey, err := key.GetKey()
	if err != nil {
		return nil, nil, err
	}
	btcPrivKey := btcec.PrivateKey(*privKey)
	sig, err = txscript.RawTxInSignature(tx, idx, pkScript, txscript.SigHashAll, &btcPrivKey)
	if err != nil {
		return nil, nil, err
	}
	return sig, btcPrivKey.PubKey().SerializeUncompressed(), nil
}

func addressToPubKeyHash(addr string, chainParams *chaincfg.Params) (*btcutil.AddressPubKeyHash, error) {
	btcAddr, err := btcutil.DecodeAddress(addr, chainParams)
	if err != nil {
		return nil, fmt.Errorf("address %s is not "+
			"intended for use on %v", addr, chainParams.Name)
	}
	Addr, ok := btcAddr.(*btcutil.AddressPubKeyHash)
	if !ok {
		return nil, fmt.Errorf("address %s is not Pay to Public Key Hash")
	}
	return Addr, nil
}
