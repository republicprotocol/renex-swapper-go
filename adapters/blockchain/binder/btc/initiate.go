package btc

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/republicprotocol/renex-swapper-go/adapters/blockchain/clients/btc"
)

// InitiateRequest contains all the information to successfully initiate an
// atomic swap on Bitcoin
type InitiateRequest struct {
	Value    int64
	HashLock []byte
	LockTime int64
	AtomicSwapRequest
}

// Initiate initiates and funds an Atomic Swap on the bitcoin blockchain.
func Initiate(req InitiateRequest) error {
	// decoding bitcoin addresses
	PersonalAddr, err := addressToPubKeyHash(string(req.PersonalAddr), req.Conn.ChainParams)
	if err != nil {
		return NewErrDecodeAddress(string(req.PersonalAddr), err)
	}
	ForeignAddr, err := addressToPubKeyHash(string(req.ForeignAddr), req.Conn.ChainParams)
	if err != nil {
		return NewErrDecodeAddress(string(req.PersonalAddr), err)
	}

	// creating atomic swap initiate script, addressScriptHash and script to
	// deposit bitcoin tokens.
	initiateScript, err := atomicSwapInitiateScript(
		PersonalAddr.Hash160(),
		ForeignAddr.Hash160(),
		req.LockTime,
		req.HashLock,
	)
	if err != nil {
		return NewErrInitiate(NewErrBuildScript(err))
	}
	initiateScriptP2SH, err := btcutil.NewAddressScriptHash(initiateScript, req.Conn.ChainParams)
	if err != nil {
		return NewErrInitiate(NewErrBuildScript(err))
	}

	// check whether the transaction is initiated
	if err := initiated(req.Conn, req.PersonalAddr, initiateScriptP2SH.String()); err != nil {
		return NewErrInitiate(err)
	}

	initiateScriptP2SHPkScript, err := txscript.PayToAddrScript(initiateScriptP2SH)
	if err != nil {
		return NewErrInitiate(NewErrBuildScript(err))
	}

	// creating unsigned transaction and adding transaction outputs
	unsignedTx := wire.NewMsgTx(req.TxVersion)
	unsignedTx.AddTxOut(wire.NewTxOut(int64(req.Value), nil))

	// signing a transaction with the given private key
	stx, complete, err := req.Conn.SignTransaction(unsignedTx, req.Key, req.Fee)
	if err != nil {
		return NewErrInitiate(NewErrSignTransaction(err))
	}
	if !complete {
		return NewErrInitiate(ErrCompleteSignTransaction)
	}

	// publish transaction on the bitcoin blockchain and wait for given number
	// of confirmations
	if err := req.Conn.PublishTransaction(stx, 1); err != nil {
		return NewErrInitiate(NewErrPublishTransaction(err))
	}

	return nil
}

// initiated checks whether the same atomic swap is initiated before, if it is
// it returns an error
func initiated(conn btc.Conn, personalAddr []byte, scriptHash string) error {
	txs, err := conn.GetRawTransactionsByAddress(string(personalAddr), -1, -1)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if output.TransactionHash == scriptHash {
				return ErrInitiated
			}
		}
	}
	return nil
}
