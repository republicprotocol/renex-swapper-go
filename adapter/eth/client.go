package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/republicprotocol/renex-swapper-go/adapter/config"
	"github.com/republicprotocol/renex-swapper-go/adapter/keystore"
)

type Conn struct {
	Network            string
	Client             *ethclient.Client
	RenExAtomicSwapper common.Address
}

// NewConnWithConfig creates a new ethereum connection with the given config
// file.
func NewConnWithConfig(config config.EthereumNetwork) (Conn, error) {
	return NewConn(config.URL, config.Network, config.Swapper)
}

// NewConn creates a new ethereum connection with the given config parameters.
func NewConn(url, network, swapperAddress string) (Conn, error) {
	ethclient, err := ethclient.Dial(url)
	if err != nil {
		return Conn{}, err
	}
	return Conn{
		Client:             ethclient,
		Network:            network,
		RenExAtomicSwapper: common.HexToAddress(swapperAddress),
	}, nil
}

// Balance of the given address
func (b *Conn) Balance(address common.Address) (*big.Int, error) {
	return b.Client.PendingBalanceAt(context.Background(), address)
}

// Transfer is a helper function for sending ETH to an address
func (b *Conn) Transfer(to common.Address, key keystore.EthereumKey, value, fee int64) error {

	// Why is there no ethclient.Transfer?
	bound := bind.NewBoundContract(to, abi.ABI{}, nil, b.Client, nil)

	key.SubmitTx(
		func(tops *bind.TransactOpts) error {
			_, err := bound.Transfer(tops)
			return err
		},
		func() bool {
			return true
		},
	)

	return nil
}
