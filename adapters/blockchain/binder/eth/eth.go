package binder

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	bindings "github.com/republicprotocol/atom-go/adapters/blockchain/bindings/eth"
	ethclient "github.com/republicprotocol/atom-go/adapters/blockchain/clients/eth"
	"github.com/republicprotocol/atom-go/domains/match"
	"github.com/republicprotocol/atom-go/domains/order"
	"github.com/republicprotocol/atom-go/domains/swap"
)

// Binder implements all methods that will communicate with the smart contracts
type Binder struct {
	mu           *sync.RWMutex
	conn         ethclient.Conn
	network      string
	transactOpts *bind.TransactOpts
	callOpts     *bind.CallOpts

	*bindings.AtomicInfo
	*bindings.Orderbook
	*bindings.RenExSettlement
	*bindings.AtomicSwap
}

// NewBinder returns a Binder to communicate with contracts
func NewBinder(auth *bind.TransactOpts, conn ethclient.Conn) (Binder, error) {
	auth.GasPrice = big.NewInt(20000000000)

	atomicInfo, err := bindings.NewAtomicInfo(conn.InfoAddress(), bind.ContractBackend(conn.Client()))
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to atom info: %v", err))
		return Binder{}, err
	}

	atomicSwap, err := bindings.NewAtomicSwap(conn.AtomAddress(), bind.ContractBackend(conn.Client()))
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to atomic swap: %v", err))
		return Binder{}, err
	}

	orderbook, err := bindings.NewOrderbook(conn.OrderBookAddress(), bind.ContractBackend(conn.Client()))
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to Orderbook: %v", err))
		return Binder{}, err
	}

	renExSettlement, err := bindings.NewRenExSettlement(conn.WalletAddress(), bind.ContractBackend(conn.Client()))
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to RenEx accounts: %v", err))
		return Binder{}, err
	}

	return Binder{
		mu:           new(sync.RWMutex),
		network:      conn.Network(),
		conn:         conn,
		transactOpts: auth,
		callOpts:     &bind.CallOpts{},

		AtomicInfo:      atomicInfo,
		AtomicSwap:      atomicSwap,
		Orderbook:       orderbook,
		RenExSettlement: renExSettlement,
	}, nil
}

// SendOwnerAddress set's the owner address for atomic swap
func (binder *Binder) SendOwnerAddress(orderID order.ID, address []byte) error {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.sendOwnerAddress(orderID, address)
}

func (binder *Binder) sendOwnerAddress(orderID order.ID, address []byte) error {
	tx, err := binder.SetOwnerAddress(binder.transactOpts, orderID, address)
	if err != nil {
		return err
	}
	_, err = binder.conn.PatchedWaitMined(context.Background(), tx)
	return err
}

// RecieveOwnerAddress recieves the owner address for atomic swap
func (binder *Binder) RecieveOwnerAddress(orderID order.ID) ([]byte, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.recieveOwnerAddress(orderID)
}

func (binder *Binder) recieveOwnerAddress(orderID order.ID) ([]byte, error) {
	return binder.GetOwnerAddress(binder.callOpts, orderID)
}

// WaitForMatch waits for the match to be found and returns the match object
func (binder *Binder) WaitForMatch(orderID order.ID) (match.Match, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.waitForMatch(orderID)
}

func (binder *Binder) waitForMatch(orderID order.ID) (match.Match, error) {
	for {
		expired, err := binder.expired(orderID)
		if err != nil {
			return nil, err
		}

		if expired {
			break
		}

		status, err := binder.OrderStatus(binder.callOpts, orderID)
		if err != nil {
			return nil, err
		}

		if status == 2 {
			PersonalOrder, ForeignOrder, ReceiveValue, SendValue, ReceiveCurrency, SendCurrency, err := binder.GetMatchDetails(&bind.CallOpts{}, orderID)
			if err != nil {
				return nil, err
			}
			return match.NewMatch(PersonalOrder, ForeignOrder, SendValue, ReceiveValue, SendCurrency, ReceiveCurrency), nil
		}

		time.Sleep(15 * time.Second)
	}
	return nil, fmt.Errorf("Order expired")
}

func (binder *Binder) expired(orderID order.ID) (bool, error) {
	details, err := binder.OrderDetails(binder.callOpts, orderID)
	if err != nil {
		return false, err
	}
	if time.Now().Unix() > int64(details.Expiry) {
		return true, nil
	}
	return false, nil
}

// SendSwapDetails stores the swap details on the ethereum blockchain
func (binder *Binder) SendSwapDetails(orderID order.ID, swapDetails []byte) error {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.sendSwapDetails(orderID, swapDetails)
}

func (binder *Binder) sendSwapDetails(orderID order.ID, swapDetails []byte) error {
	tx, err := binder.SubmitDetails(binder.transactOpts, orderID, swapDetails)
	if err != nil {
		return err
	}
	_, err = binder.conn.PatchedWaitMined(context.Background(), tx)
	return err
}

// RecieveSwapDetails recieves the swap details from the ethereum blockchain
func (binder *Binder) RecieveSwapDetails(orderID order.ID) ([]byte, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.recieveSwapDetails(orderID)
}

func (binder *Binder) recieveSwapDetails(orderID order.ID) ([]byte, error) {
	return binder.SwapDetails(binder.callOpts, orderID)
}

// InfoTimeStamp returns the time at which the address for the atomic swap is submitted.
func (binder *Binder) InfoTimeStamp(orderID order.ID) (int64, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.infoTimeStamp(orderID)
}

func (binder *Binder) infoTimeStamp(orderID order.ID) (int64, error) {
	ts, err := binder.OwnerAddressTimestamp(binder.callOpts, orderID)
	if err != nil {
		return 0, err
	}
	return ts.Int64(), nil
}

// InitiateTimeStamp returns the time at which the atomic swap is intiated.
func (binder *Binder) InitiateTimeStamp(orderID order.ID) (int64, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.initiateTimeStamp(orderID)
}

func (binder *Binder) initiateTimeStamp(orderID order.ID) (int64, error) {
	ts, err := binder.SwapDetailsTimestamp(binder.callOpts, orderID)
	if err != nil {
		return 0, err
	}
	return ts.Int64(), nil
}

// RedeemTimeStamp returns the time at which the atomic swap is redeemed.
func (binder *Binder) RedeemTimeStamp(orderID swap.ID) (int64, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.redeemTimeStamp(orderID)
}

func (binder *Binder) redeemTimeStamp(orderID swap.ID) (int64, error) {
	ts, err := binder.RedeemedAt(binder.callOpts, orderID)
	if err != nil {
		return 0, err
	}
	return ts.Int64(), nil
}

// InitiateAtomicSwap initiates a new Ethereum Atomic swap
func (binder *Binder) InitiateAtomicSwap(swapID swap.ID, to []byte, hash [32]byte, value *big.Int, expiry int64) error {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.initiateAtomicSwap(swapID, to, hash, value, expiry)
}

func (binder *Binder) initiateAtomicSwap(swapID swap.ID, to []byte, hash [32]byte, value *big.Int, expiry int64) error {
	transactOpts := *binder.transactOpts
	auth := &transactOpts
	auth.Value = value
	auth.GasLimit = 3000000

	tx, err := binder.Initiate(auth, swapID, common.BytesToAddress(to), hash, big.NewInt(expiry))
	if err != nil {
		return err
	}
	_, err = binder.conn.PatchedWaitMined(context.Background(), tx)
	return err
}

// RedeemAtomicSwap initiates a new Ethereum Atomic swap
func (binder *Binder) RedeemAtomicSwap(swapID [32]byte, secret [32]byte) error {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.redeemAtomicSwap(swapID, secret)
}

func (binder *Binder) redeemAtomicSwap(swapID [32]byte, secret [32]byte) error {
	transactOpts := *binder.transactOpts
	auth := &transactOpts
	auth.GasLimit = 3000000
	tx, err := binder.Redeem(auth, swapID, secret)
	if err == nil {
		_, err = binder.conn.PatchedWaitMined(context.Background(), tx)
	}
	return err
}

// RefundAtomicSwap refunds an Ethereum Atomic swap
func (binder *Binder) RefundAtomicSwap(swapID [32]byte) error {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.refundAtomicSwap(swapID)
}

func (binder *Binder) refundAtomicSwap(swapID [32]byte) error {
	transactOpts := *binder.transactOpts
	auth := &transactOpts
	auth.GasLimit = 3000000
	tx, err := binder.Refund(auth, swapID)
	if err == nil {
		_, err = binder.conn.PatchedWaitMined(context.Background(), tx)
	}
	return err
}

// AuditAtomicSwap Audits an Atomic swap
func (binder *Binder) AuditAtomicSwap(swapID [32]byte) ([32]byte, []byte, *big.Int, int64, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.auditAtomicSwap(swapID)
}

func (binder *Binder) auditAtomicSwap(swapID [32]byte) ([32]byte, []byte, *big.Int, int64, error) {
	auditReport, err := binder.Audit(&bind.CallOpts{}, swapID)
	if err != nil {
		return [32]byte{}, nil, nil, 0, err
	}
	return auditReport.SecretLock, auditReport.From.Bytes(), auditReport.Value, auditReport.Timelock.Int64(), nil
}

// AuditSecretAtomicSwap audits the secret of an Atom swap
func (binder *Binder) AuditSecretAtomicSwap(swapID [32]byte) ([32]byte, error) {
	binder.mu.Lock()
	defer binder.mu.Unlock()
	return binder.auditSecretAtomicSwap(swapID)
}

func (binder *Binder) auditSecretAtomicSwap(swapID swap.ID) ([32]byte, error) {
	return binder.AuditSecret(&bind.CallOpts{}, swapID)
}
