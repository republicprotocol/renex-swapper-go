package network

import "github.com/republicprotocol/renex-swapper-go/domains/order"

type Network interface {
	SendOwnerAddress(order.ID, []byte) error
	SendSwapDetails(order.ID, []byte) error
	RecieveOwnerAddress(order.ID, int64) ([]byte, error)
	RecieveSwapDetails(order.ID, int64) ([]byte, error)
}
