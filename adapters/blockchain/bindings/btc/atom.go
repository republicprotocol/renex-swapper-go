package btc

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/republicprotocol/renex-swapper-go/adapters/blockchain/clients/btc"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/keystore"
	"golang.org/x/crypto/ripemd160"
)

const (

	/*
		Bitcoin Refund Script: Alice is trying to get refunded

		OP_DATA_73  (1)
		<Signature> (73)
		OP_DATA_33  (1)
		<PublicKey> (33)
		OP_DATA_32  (1)
		<Secret>    (32)
		<True>     (1)

	*/

	redeemAtomicSwapSigScriptSize = 1 + 73 + 1 + 33 + 1 + 32 + 1

	/*
	   Bitcoin Refund Script: Alice is trying to get refunded

	   OP_DATA_73  (1)
	   <Signature> (73)
	   OP_DATA_33  (1)
	   <PublicKey> (33)
	   <False>     (1)

	*/
	refundAtomicSwapSigScriptSize = 1 + 73 + 1 + 33 + 1

	txVersion = 2

	secretSize = 32

	verify = true
)

type builtContract struct {
	contract       []byte
	contractP2SH   btcutil.Address
	contractTxHash *chainhash.Hash
	contractTx     *wire.MsgTx
	refundTx       *wire.MsgTx
}

type contractArgs struct {
	me         *btcutil.AddressPubKeyHash
	them       *btcutil.AddressPubKeyHash
	amount     int64
	locktime   int64
	secretHash []byte
}

type redeemResult struct {
	RedeemTx     []byte
	RedeemTxHash [32]byte
}

type readResult struct {
	ContractAddress  []byte
	Amount           int64
	RecipientAddress []byte
	RefundAddress    []byte
	SecretHash       [32]byte
	LockTime         int64
}

type bitcoinData struct {
	ContractHash   string
	Contract       []byte
	ContractTxHash []byte
	ContractTx     []byte
	RefundTxHash   [32]byte
	RefundTx       []byte
}

/*
Bitcoin AtomicSwap Script: Alice is trying to do an atomic swap with bob.

OP_IF
	OP_SHA256
	<secret_hash>
	OP_EQUALVERIFY
	OP_DUP
	OP_HASH160
	<pubkey_hash_bob>
OP_ELSE
	<lock_time>
	OP_CHECKLOCKTIMEVERIFY
	OP_DROP
	OP_DUP
	OP_HASH160
	<pubKey_hash_alice>
OP_ENDIF
OP_EQUALVERIFY
OP_CHECKSIG

*/

func atomicSwapContract(pkhMe, pkhThem *[ripemd160.Size]byte, locktime int64, secretHash []byte) ([]byte, error) {
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

/*
Bitcoin Refund Script: Alice is trying to get refunded

<Signature>
<PublicKey>
<False>(Int 0)
<Contract>
*/
func refundP2SHContract(contract, sig, pubkey []byte) ([]byte, error) {
	b := txscript.NewScriptBuilder()
	b.AddData(sig)
	b.AddData(pubkey)
	b.AddInt64(0)
	b.AddData(contract)
	return b.Script()
}

/*
Bitcoin Refund Script: Bob is trying to redeem and get his bitcoins.

<Signature>
<PublicKey>
<Secret>
<True>(Int 1)
<Contract>
*/

func redeemP2SHContract(contract, sig, pubkey []byte, secret [32]byte) ([]byte, error) {
	b := txscript.NewScriptBuilder()
	b.AddData(sig)
	b.AddData(pubkey)
	b.AddData(secret[:])
	b.AddInt64(1)
	b.AddData(contract)
	return b.Script()
}

func Initiate(conn btc.Conn, myKey keystore.Key, participantAddress string, value int64, hash []byte, lockTime int64) (bitcoinData, error) {
	addr, err := myKey.GetAddress()
	if err != nil {
		return bitcoinData{}, fmt.Errorf("failed to decode your address: %v", err)
	}
	myAddr, err := btcutil.DecodeAddress(string(addr), conn.ChainParams)
	if err != nil {
		return bitcoinData{}, fmt.Errorf("your address is not "+
			"intended for use on %v", conn.ChainParams.Name)
	}

	myAddrP2PKH, ok := myAddr.(*btcutil.AddressPubKeyHash)
	if !ok {
		return bitcoinData{}, errors.New("your address is not P2PKH")
	}

	cp2Addr, err := btcutil.DecodeAddress(participantAddress, conn.ChainParams)
	if err != nil {
		return bitcoinData{}, fmt.Errorf("failed to decode participant address: %v", err)
	}
	if !cp2Addr.IsForNet(conn.ChainParams) {
		return bitcoinData{}, fmt.Errorf("participant address is not "+
			"intended for use on %v", conn.ChainParams.Name)
	}

	cp2AddrP2PKH, ok := cp2Addr.(*btcutil.AddressPubKeyHash)
	if !ok {
		return bitcoinData{}, errors.New("participant address is not P2PKH")
	}

	b, err := buildContract(conn, &contractArgs{
		me:         myAddrP2PKH,
		them:       cp2AddrP2PKH,
		amount:     value,
		locktime:   lockTime,
		secretHash: hash,
	}, myKey)

	if err != nil {
		return bitcoinData{}, err
	}

	// redeemSig, redeemPubKey, err := createSig(conn, redeemTx, 0, contract, myAddr)

	var contractBuf bytes.Buffer
	contractBuf.Grow(b.contractTx.SerializeSize())
	b.contractTx.Serialize(&contractBuf)

	var refundBuf bytes.Buffer
	refundBuf.Grow(b.refundTx.SerializeSize())
	b.refundTx.Serialize(&refundBuf)

	txHash, err := conn.PromptPublishTx(b.contractTx, "contract")
	if err != nil {
		return bitcoinData{}, err
	}

	refundTx := *b.refundTx
	return bitcoinData{
		Contract:       b.contract,
		ContractHash:   b.contractP2SH.EncodeAddress(),
		ContractTx:     contractBuf.Bytes(),
		ContractTxHash: txHash.CloneBytes(),
		RefundTx:       refundBuf.Bytes(),
		RefundTxHash:   refundTx.TxHash(),
	}, conn.WaitTillMined(txHash, 0)
}

func Redeem(conn btc.Conn, myKey keystore.Key, contract, contractTxBytes []byte, secret [32]byte) (redeemResult, error) {
	var contractTx wire.MsgTx
	err := contractTx.Deserialize(bytes.NewReader(contractTxBytes))
	if err != nil {
		return redeemResult{}, fmt.Errorf("failed to decode contract transaction: %v", err)
	}

	pushes, err := txscript.ExtractAtomicSwapDataPushes(0, contract)
	if err != nil {
		return redeemResult{}, err
	}
	if pushes == nil {
		return redeemResult{}, errors.New("contract is not an atomic swap script recognized by this tool")
	}
	recipientAddr, err := btcutil.NewAddressPubKeyHash(pushes.RecipientHash160[:], conn.ChainParams)
	if err != nil {
		return redeemResult{}, err
	}
	contractHash := btcutil.Hash160(contract)
	contractOut := -1
	for i, out := range contractTx.TxOut {
		sc, addrs, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, conn.ChainParams)
		if sc == txscript.ScriptHashTy &&
			bytes.Equal(addrs[0].(*btcutil.AddressScriptHash).Hash160()[:], contractHash) {
			contractOut = i
			break
		}
	}
	if contractOut == -1 {
		return redeemResult{}, errors.New("transaction does not contain a contract output")
	}

	myAddress, err := myKey.GetAddress()
	if err != nil {
		return redeemResult{}, fmt.Errorf("Decoded Address: %v", err)
	}

	if bytes.Compare([]byte(recipientAddr.EncodeAddress()), myAddress) != 0 {
		return redeemResult{}, fmt.Errorf("Reciepient Address Mismatch")
	}

	addr, err := btcutil.DecodeAddress(string(myAddress), conn.ChainParams)
	if err != nil {
		return redeemResult{}, fmt.Errorf("Decoded Address: %v", err)
	}

	outScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return redeemResult{}, err
	}

	contractTxHash := contractTx.TxHash()

	contractOutPoint := wire.OutPoint{
		Hash:  contractTxHash,
		Index: uint32(contractOut),
	}

	redeemTx := wire.NewMsgTx(txVersion)
	redeemTx.LockTime = uint32(pushes.LockTime)
	redeemTx.AddTxIn(wire.NewTxIn(&contractOutPoint, nil, nil))
	redeemTx.AddTxOut(wire.NewTxOut(contractTx.TxOut[contractOut].Value-10000, outScript))

	redeemSig, redeemPubKey, err := createSig(conn, redeemTx, 0, contract, myKey)
	if err != nil {
		return redeemResult{}, err
	}
	redeemSigScript, err := redeemP2SHContract(contract, redeemSig, redeemPubKey, secret)
	if err != nil {
		return redeemResult{}, err
	}
	redeemTx.TxIn[0].SignatureScript = redeemSigScript

	redeemTxHash := redeemTx.TxHash()

	var buf bytes.Buffer
	buf.Grow(redeemTx.SerializeSize())
	redeemTx.Serialize(&buf)

	if verify {
		e, err := txscript.NewEngine(contractTx.TxOut[contractOutPoint.Index].PkScript,
			redeemTx, 0, txscript.StandardVerifyFlags, txscript.NewSigCache(10),
			txscript.NewTxSigHashes(redeemTx), contractTx.TxOut[contractOut].Value)
		if err != nil {
			return redeemResult{}, err
		}
		err = e.Execute()
		if err != nil {
			return redeemResult{}, err
		}
	}

	txHash, err := conn.PromptPublishTx(redeemTx, "redeem")
	if err != nil {
		return redeemResult{}, err
	}

	return redeemResult{
		RedeemTx:     buf.Bytes(),
		RedeemTxHash: redeemTxHash,
	}, conn.WaitTillMined(txHash, 0)

}

func Refund(conn btc.Conn, myKey keystore.Key, contract, contractTxBytes []byte) error {

	var contractTx wire.MsgTx
	err := contractTx.Deserialize(bytes.NewReader(contractTxBytes))
	if err != nil {
		return fmt.Errorf("failed to decode contract transaction: %v", err)
	}

	pushes, err := txscript.ExtractAtomicSwapDataPushes(0, contract)
	if err != nil {
		return err
	}
	if pushes == nil {
		return errors.New("contract is not an atomic swap script recognized by this tool")
	}

	refundTx, err := buildRefund(conn, myKey, contract, &contractTx)
	if err != nil {
		return err
	}

	txHash, err := conn.PromptPublishTx(refundTx, "refund")
	if err != nil {
		return err
	}

	return conn.WaitTillMined(txHash, 0)
}

func Audit(conn btc.Conn, contract, contractTxBytes []byte) (readResult, error) {

	var contractTx wire.MsgTx
	err := contractTx.Deserialize(bytes.NewReader(contractTxBytes))
	if err != nil {
		return readResult{}, fmt.Errorf("failed to decode contract transaction: %v", err)
	}

	contractHash160 := btcutil.Hash160(contract)
	contractOut := -1

	for i, out := range contractTx.TxOut {
		sc, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, conn.ChainParams)
		if err != nil || sc != txscript.ScriptHashTy {
			continue
		}
		if bytes.Equal(addrs[0].(*btcutil.AddressScriptHash).Hash160()[:], contractHash160) {
			contractOut = i
			break
		}
	}
	if contractOut == -1 {
		return readResult{}, errors.New("transaction does not contain the contract output")
	}

	pushes, err := txscript.ExtractAtomicSwapDataPushes(0, contract)
	if err != nil {
		return readResult{}, err
	}
	if pushes == nil {
		return readResult{}, errors.New("contract is not an atomic swap script recognized by this tool")
	}

	contractAddr, err := btcutil.NewAddressScriptHash(contract, conn.ChainParams)
	if err != nil {
		return readResult{}, err
	}
	recipientAddr, err := btcutil.NewAddressPubKeyHash(pushes.RecipientHash160[:],
		conn.ChainParams)
	if err != nil {
		return readResult{}, err
	}
	refundAddr, err := btcutil.NewAddressPubKeyHash(pushes.RefundHash160[:],
		conn.ChainParams)
	if err != nil {
		return readResult{}, err
	}

	return readResult{
		ContractAddress:  contractAddr.ScriptAddress(),
		Amount:           int64(btcutil.Amount(contractTx.TxOut[contractOut].Value)),
		RecipientAddress: []byte(recipientAddr.EncodeAddress()),
		RefundAddress:    []byte(refundAddr.EncodeAddress()),
		SecretHash:       pushes.SecretHash,
		LockTime:         pushes.LockTime,
	}, nil
}

func AuditSecret(conn btc.Conn, redemptionTxBytes, secretHash []byte) ([32]byte, error) {
	var redemptionTx wire.MsgTx
	err := redemptionTx.Deserialize(bytes.NewReader(redemptionTxBytes))
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to decode redemption transaction: %v", err)
	}

	if len(secretHash) != sha256.Size {
		return [32]byte{}, errors.New("secret hash has wrong size")
	}

	for _, in := range redemptionTx.TxIn {
		pushes, err := txscript.PushedData(in.SignatureScript)
		if err != nil {
			return [32]byte{}, err
		}
		for _, push := range pushes {
			if bytes.Equal(sha256Hash(push), secretHash) {
				var secret [32]byte
				for i := 0; i < 32; i++ {
					secret[i] = push[i]
				}
				return secret, nil
			}
		}
	}
	return [32]byte{}, errors.New("transaction does not contain the secret")
}

func sumOutputSerializeSizes(outputs []*wire.TxOut) (serializeSize int) {
	for _, txOut := range outputs {
		serializeSize += txOut.SerializeSize()
	}
	return serializeSize
}

func inputSize(sigScriptSize int) int {
	return 32 + 4 + wire.VarIntSerializeSize(uint64(sigScriptSize)) + sigScriptSize + 4
}

func estimateRedeemSerializeSize(contract []byte, txOuts []*wire.TxOut) int {
	contractPush, err := txscript.NewScriptBuilder().AddData(contract).Script()
	if err != nil {
		panic(err)
	}
	contractPushSize := len(contractPush)

	return 12 + wire.VarIntSerializeSize(1) +
		wire.VarIntSerializeSize(uint64(len(txOuts))) +
		inputSize(redeemAtomicSwapSigScriptSize+contractPushSize) +
		sumOutputSerializeSizes(txOuts)
}

func sha256Hash(x []byte) []byte {
	h := sha256.Sum256(x)
	return h[:]
}

func buildContract(conn btc.Conn, args *contractArgs, key keystore.Key) (*builtContract, error) {
	contract, err := atomicSwapContract(args.me.Hash160(), args.them.Hash160(),
		args.locktime, args.secretHash)
	if err != nil {
		return nil, err
	}
	contractP2SH, err := btcutil.NewAddressScriptHash(contract, conn.ChainParams)
	if err != nil {
		return nil, err
	}
	contractP2SHPkScript, err := txscript.PayToAddrScript(contractP2SH)
	if err != nil {
		return nil, err
	}

	unsignedContract := wire.NewMsgTx(txVersion)
	unsignedContract.AddTxOut(wire.NewTxOut(int64(args.amount), contractP2SHPkScript))

	contractTx, complete, err := conn.SignTransaction(unsignedContract, key)
	if err != nil {
		return nil, fmt.Errorf("signrawtransaction: %v", err)
	}
	if !complete {
		return nil, errors.New("signrawtransaction: failed to completely sign contract transaction")
	}

	contractTxHash := contractTx.TxHash()

	refundTx, err := buildRefund(conn, key, contract, contractTx)
	if err != nil {
		return nil, err
	}

	return &builtContract{
		contract,
		contractP2SH,
		&contractTxHash,
		contractTx,
		refundTx,
	}, nil
}

func buildRefund(conn btc.Conn, key keystore.Key, contract []byte, contractTx *wire.MsgTx) (
	refundTx *wire.MsgTx, err error) {

	contractP2SH, err := btcutil.NewAddressScriptHash(contract, conn.ChainParams)
	if err != nil {
		return nil, err
	}
	contractP2SHPkScript, err := txscript.PayToAddrScript(contractP2SH)
	if err != nil {
		return nil, err
	}

	contractTxHash := contractTx.TxHash()
	contractOutPoint := wire.OutPoint{Hash: contractTxHash, Index: ^uint32(0)}
	for i, o := range contractTx.TxOut {
		if bytes.Equal(o.PkScript, contractP2SHPkScript) {
			contractOutPoint.Index = uint32(i)
			break
		}
	}
	if contractOutPoint.Index == ^uint32(0) {
		return nil, errors.New("contract tx does not contain a P2SH contract payment")
	}

	addr, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	refundAddress, err := btcutil.DecodeAddress(string(addr), conn.ChainParams)
	if err != nil {
		return nil, err
	}

	refundOutScript, err := txscript.PayToAddrScript(refundAddress)
	if err != nil {
		return nil, err
	}

	pushes, err := txscript.ExtractAtomicSwapDataPushes(0, contract)
	if err != nil {
		return nil, err
	}
	if pushes == nil {
		return nil, fmt.Errorf("failed to extract atomic swap data")
	}

	refundAddr, err := btcutil.NewAddressPubKeyHash(pushes.RefundHash160[:], conn.ChainParams)
	if err != nil {
		return nil, err
	}

	if bytes.Compare([]byte(refundAddr.EncodeAddress()), addr) != 0 {
		return nil, fmt.Errorf("Refund Address Mismatch")
	}

	refundTx = wire.NewMsgTx(txVersion)
	refundTx.LockTime = uint32(pushes.LockTime)
	refundTx.AddTxOut(wire.NewTxOut(contractTx.TxOut[contractOutPoint.Index].Value-10000, refundOutScript))

	txIn := wire.NewTxIn(&contractOutPoint, nil, nil)
	txIn.Sequence = 0
	refundTx.AddTxIn(txIn)

	refundSig, refundPubKey, err := createSig(conn, refundTx, 0, contract, key)
	if err != nil {
		return nil, err
	}
	refundSigScript, err := refundP2SHContract(contract, refundSig, refundPubKey)
	if err != nil {
		return nil, err
	}
	refundTx.TxIn[0].SignatureScript = refundSigScript

	if verify {
		e, err := txscript.NewEngine(contractTx.TxOut[contractOutPoint.Index].PkScript,
			refundTx, 0, txscript.StandardVerifyFlags, txscript.NewSigCache(10),
			txscript.NewTxSigHashes(refundTx), contractTx.TxOut[contractOutPoint.Index].Value)

		if err != nil {
			return nil, err
		}
		err = e.Execute()
		if err != nil {
			return nil, err
		}
	}

	return refundTx, nil
}

func estimateRefundSerializeSize(contract []byte, txOuts []*wire.TxOut) int {
	contractPush, err := txscript.NewScriptBuilder().AddData(contract).Script()
	if err != nil {
		// Should never be hit since this script does exceed the limits.
		panic(err)
	}
	contractPushSize := len(contractPush)

	// 12 additional bytes are for version, locktime and expiry.
	return 12 + wire.VarIntSerializeSize(1) +
		wire.VarIntSerializeSize(uint64(len(txOuts))) +
		inputSize(refundAtomicSwapSigScriptSize+contractPushSize) +
		sumOutputSerializeSizes(txOuts)
}

func createSig(conn btc.Conn, tx *wire.MsgTx, idx int, pkScript []byte, key keystore.Key) (sig, pubkey []byte, err error) {
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
