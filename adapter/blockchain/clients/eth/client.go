package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/republicprotocol/renex-swapper-go/adapter/config"
)

type Conn struct {
	network            string
	client             *ethclient.Client
	renExAtomicSwapper common.Address
	renExAtomicInfo    common.Address
	renExSettlement    common.Address
	orderbook          common.Address
}

// Connect to an ethereum network.
func Connect(config config.Config) (Conn, error) {
	ethclient, err := ethclient.Dial(config.Ethereum.URL)
	if err != nil {
		return Conn{}, err
	}

	return Conn{
		client:             ethclient,
		network:            config.Ethereum.Network,
		renExAtomicSwapper: common.HexToAddress(config.RenEx.Swapper),
		renExSettlement:    common.HexToAddress(config.RenEx.Settlement),
		orderbook:          common.HexToAddress(config.RenEx.Orderbook),
	}, nil
}

// NewAccount creates a new account and funds it with ether
func (conn *Conn) NewAccount(value int64, from *bind.TransactOpts) (common.Address, *bind.TransactOpts, error) {
	account, err := crypto.GenerateKey()
	if err != nil {
		return common.Address{}, &bind.TransactOpts{}, err
	}

	accountAddress := crypto.PubkeyToAddress(account.PublicKey)
	accountAuth := bind.NewKeyedTransactor(account)

	return accountAddress, accountAuth, conn.Transfer(accountAddress, from, value)
}

// Transfer is a helper function for sending ETH to an address
func (conn *Conn) Transfer(to common.Address, from *bind.TransactOpts, value int64) error {
	transactor := &bind.TransactOpts{
		From:     from.From,
		Nonce:    from.Nonce,
		Signer:   from.Signer,
		Value:    big.NewInt(value),
		GasPrice: from.GasPrice,
		GasLimit: 30000,
		Context:  from.Context,
	}

	// Why is there no ethclient.Transfer?
	bound := bind.NewBoundContract(to, abi.ABI{}, nil, conn.client, nil)
	tx, err := bound.Transfer(transactor)
	if err != nil {
		return err
	}
	_, err = conn.PatchedWaitMined(context.Background(), tx)
	return err
}

// PatchedWaitMined waits for tx to be mined on the blockchain.
// It stops waiting when the context is canceled.
//
// TODO: THIS DOES NOT WORK WITH PARITY, WHICH SENDS A TRANSACTION RECEIPT UPON
// RECEIVING A TX, NOT AFTER IT'S MINED
func (conn *Conn) PatchedWaitMined(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	switch conn.network {
	case "ganache":
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	default:
		receipt, err := bind.WaitMined(ctx, conn.client, tx)
		if err != nil {
			return nil, err
		}
		if receipt.Status != 1 {
			return nil, fmt.Errorf("Transaction reverted")
		}
		return receipt, nil
	}
}

// PatchedWaitDeployed waits for a contract deployment transaction and returns the on-chain
// contract address when it is mined. It stops waiting when ctx is canceled.
//
// TODO: THIS DOES NOT WORK WITH PARITY, WHICH SENDS A TRANSACTION RECEIPT UPON
// RECEIVING A TX, NOT AFTER IT'S MINED
func (conn *Conn) PatchedWaitDeployed(ctx context.Context, tx *types.Transaction) (common.Address, error) {
	switch conn.network {
	case "ganache":
		time.Sleep(100 * time.Millisecond)
		return common.Address{}, nil
	default:
		return bind.WaitDeployed(ctx, conn.client, tx)
	}
}

// Balance returns the balance of the given ethereum address
func (conn *Conn) Balance(addr common.Address) (*big.Int, error) {
	return conn.client.PendingBalanceAt(context.Background(), addr)
}

func (conn *Conn) RenExAtomicSwapperAddress() common.Address {
	return conn.renExAtomicSwapper
}

func (conn *Conn) RenExAtomicInfoAddress() common.Address {
	return conn.renExAtomicInfo
}

func (conn *Conn) RenExSettlementAddress() common.Address {
	return conn.renExSettlement
}

func (conn *Conn) OrderbookAddress() common.Address {
	return conn.orderbook
}

func (conn *Conn) Network() string {
	return conn.network
}

func (conn *Conn) Client() *ethclient.Client {
	return conn.client
}
