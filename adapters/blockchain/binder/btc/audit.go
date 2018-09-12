package btc

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type auditResponse struct {
	Amount        int64
	RecipientAddr []byte
	RefundAddr    []byte
	SecretHash    [32]byte
	LockTime      int64
}

func audit(req AtomicSwapRequest) (auditResponse, error) {
	txOut, err := waitForCounterInitiation(req)

	initiateScript, err := txscript.ExtractAtomicSwapDataPushes(0, txOut.PkScript)
	if err != nil {
		return auditResponse{}, err
	}
	if initiateScript == nil {
		return auditResponse{}, errors.New("contract is not an atomic swap script recognized by this tool")
	}

	recipientAddr, err := btcutil.NewAddressPubKeyHash(
		initiateScript.RecipientHash160[:],
		req.Conn.ChainParams,
	)
	if err != nil {
		return auditResponse{}, err
	}
	refundAddr, err := btcutil.NewAddressPubKeyHash(
		initiateScript.RefundHash160[:],
		req.Conn.ChainParams,
	)
	if err != nil {
		return auditResponse{}, err
	}

	return auditResponse{
		Amount:        int64(btcutil.Amount(txOut.Value)),
		RecipientAddr: []byte(recipientAddr.EncodeAddress()),
		RefundAddr:    []byte(refundAddr.EncodeAddress()),
		SecretHash:    initiateScript.SecretHash,
		LockTime:      initiateScript.LockTime,
	}, nil

}

func waitForCounterInitiation(req AtomicSwapRequest) (*wire.TxOut, error) {
	// decoding bitcoin addresses
	PersonalAddr, err := addressToPubKeyHash(string(req.PersonalAddr), req.Conn.ChainParams)
	if err != nil {
		return nil, NewErrDecodeAddress(string(req.PersonalAddr), err)
	}
	ForeignAddr, err := addressToPubKeyHash(string(req.ForeignAddr), req.Conn.ChainParams)
	if err != nil {
		return nil, NewErrDecodeAddress(string(req.PersonalAddr), err)
	}

	// wait for the initiating trader to submit the atomic swap transaction
	// to the bitcoin blockchain
	txs, err := req.Conn.GetRawTransactionsByAddress(ForeignAddr.String(), -1, -1)
	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if strings.Contains(output.Script, PersonalAddr.String()[2:]) &&
				strings.Contains(output.Script, PersonalAddr.String()[2:]) {
				pkScript, err := hex.DecodeString(output.Script)
				if err != nil {
					return nil, err
				}
				return wire.NewTxOut(output.Value, pkScript), nil
			}
		}
	}
	return nil, err
}
