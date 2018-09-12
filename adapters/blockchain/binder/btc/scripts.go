package btc

import (
	"github.com/btcsuite/btcd/txscript"
	"golang.org/x/crypto/ripemd160"
)

// AtomicSwapRefundScriptSize is the size of the Bitcoin Atomic Swap's
// RefundScript
const AtomicSwapRefundScriptSize = 1 + 73 + 1 + 33 + 1

// AtomicSwapRedeemScriptSize is the size of the Bitcoin Atomic Swap's
// RedeemScript
const AtomicSwapRedeemScriptSize = 1 + 73 + 1 + 33 + 1 + 32 + 1

// atomicSwapInitiateScript creates a Bitcoin Atomic Swap initiate script.
//
//			OP_IF
//				OP_SHA256
//				<secret_hash>
//				OP_EQUALVERIFY
//				OP_DUP
//				OP_HASH160
//				<foreign_address>
//			OP_ELSE
//				<lock_time>
//				OP_CHECKLOCKTIMEVERIFY
//				OP_DROP
//				OP_DUP
//				OP_HASH160
//				<personal_address>
//			OP_ENDIF
//			OP_EQUALVERIFY
//			OP_CHECKSIG
//
func atomicSwapInitiateScript(pkhMe, pkhThem *[ripemd160.Size]byte, locktime int64, secretHash []byte) ([]byte, error) {
	b := txscript.NewScriptBuilder()

	b.AddOp(txscript.OP_IF)
	{
		b.AddOp(txscript.OP_SIZE)
		b.AddData([]byte{32})
		b.AddOp(txscript.OP_EQUALVERIFY)
		b.AddOp(txscript.OP_SHA256)
		b.AddData(secretHash)
		b.AddOp(txscript.OP_EQUALVERIFY)
		b.AddOp(txscript.OP_DUP)
		b.AddOp(txscript.OP_HASH160)
		b.AddData(pkhThem[:])
	}
	b.AddOp(txscript.OP_ELSE)
	{
		b.AddInt64(locktime)
		b.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
		b.AddOp(txscript.OP_DROP)
		b.AddOp(txscript.OP_DUP)
		b.AddOp(txscript.OP_HASH160)
		b.AddData(pkhMe[:])
	}
	b.AddOp(txscript.OP_ENDIF)
	b.AddOp(txscript.OP_EQUALVERIFY)
	b.AddOp(txscript.OP_CHECKSIG)

	return b.Script()
}

// atomicSwapRedeemScript creates a Redeem Script for the Bitcoin Atomic Swap.
//
//			<Signature>
//			<PublicKey>
//			<Secret>
//			<True>(Int 1)
//			<InitiateScript>
//
func atomicSwapRedeemScript(initiateScript, sig, pubkey []byte, secret [32]byte) ([]byte, error) {
	b := txscript.NewScriptBuilder()
	b.AddData(sig)
	b.AddData(pubkey)
	b.AddData(secret[:])
	b.AddInt64(1)
	b.AddData(initiateScript)
	return b.Script()
}

// AtomicSwapRefundScript creates a Bitcoin Refund Atomic Swap.
//
//			<Signature>
//			<PublicKey>
//			<False>(Int 0)
//			<InitiateScript>
//
func AtomicSwapRefundScript(initiateScript, sig, pubkey []byte) ([]byte, error) {
	b := txscript.NewScriptBuilder()
	b.AddData(sig)
	b.AddData(pubkey)
	b.AddInt64(0)
	b.AddData(initiateScript)
	return b.Script()
}
