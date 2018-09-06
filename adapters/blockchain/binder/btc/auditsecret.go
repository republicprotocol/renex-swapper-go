package btc

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type AuditSecretRequest struct {
	TxHash     string
	SecretHash [32]byte
	AtomicSwapRequest
}

type AuditSecretResponse struct {
	Secret [32]byte
}

func AuditSecret(req AuditSecretRequest) (AuditSecretResponse, error) {
	sigScript, err := WaitForCounterRedemption(req.AtomicSwapRequest, req.TxHash)
	pushes, err := txscript.PushedData(sigScript)
	if err != nil {
		return AuditSecretResponse{}, err
	}
	for _, push := range pushes {
		if sha256.Sum256(push) == req.SecretHash {
			var secret [32]byte
			for i := 0; i < 32; i++ {
				secret[i] = push[i]
			}
			return AuditSecretResponse{
				Secret: secret,
			}, nil
		}
	}

}

func WaitForCounterRedemption(req AtomicSwapRequest, txHash string) ([]byte, error) {
	tx, err := req.Conn.GetRawTransaction(txHash)
	txs, err := req.Conn.GetRawTransactionsByAddress(tx.Outputs[0].TransactionHash, 0, 0)

	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if strings.Contains(output.Script, req.PersonalAddr.String()[2:]) &&
				strings.Contains(output.Script, req.PersonalAddr.String()[2:]) {
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
