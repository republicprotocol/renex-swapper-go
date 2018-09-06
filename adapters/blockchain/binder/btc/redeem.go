package btc

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type RedeemRequest struct {
	Secret     [32]byte
	TxHash     chainhash.Hash
	TxOut      *wire.TxOut
	TxOutIndex uint32
	AtomicSwapRequest
}

// Redeem redeems and funds an Atomic Swap on the bitcoin blockchain.
func Redeem(req RedeemRequest) error {
	// decoding bitcoin addresses
	PersonalAddr, err := addressToPubKeyHash(string(req.PersonalAddr), req.Conn.ChainParams)
	if err != nil {
		return NewErrDecodeAddress(string(req.PersonalAddr), err)
	}

	// create bitcoin script to pay to the user's personal address
	payToAddrScript, err := txscript.PayToAddrScript(PersonalAddr)
	if err != nil {
		return err
	}

	// build transaction
	redeemTx := wire.NewMsgTx(req.TxVersion)
	redeemTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&req.TxHash, req.TxOutIndex), nil, nil))
	redeemTx.AddTxOut(wire.NewTxOut(req.TxOut.Value-req.Fee, payToAddrScript))

	// sign transaction
	redeemSig, redeemPubKey, err := sign(redeemTx, 0, req.TxOut.PkScript, req.Key)
	if err != nil {
		return err
	}

	// build signature script
	redeemSigScript, err := atomicSwapRedeemScript(req.TxOut.PkScript, redeemSig, redeemPubKey, req.Secret)
	if err != nil {
		return err
	}
	redeemTx.TxIn[0].SignatureScript = redeemSigScript

	// verifing the redeem script
	e, err := txscript.NewEngine(req.TxOut.PkScript, redeemTx, 0,
		txscript.StandardVerifyFlags, txscript.NewSigCache(10),
		txscript.NewTxSigHashes(redeemTx), req.TxOut.Value)
	if err != nil {
		return err
	}
	err = e.Execute()
	if err != nil {
		return err
	}

	// publishing the transaction
	return req.Conn.PublishTransaction(redeemTx, 1)
}
