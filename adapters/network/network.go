package network

import "github.com/republicprotocol/renex-swapper-go/domains/order"

type Network interface {
	SendOwnerAddress(order.ID, []byte) error
	ReceiveOwnerAddress(order.ID, int64) ([]byte, error)
	ReceiveSwapDetails(order.ID, int64) ([]byte, error)
	SendSwapDetails(order.ID, []byte) error
}
