package storage

import (
	"github.com/republicprotocol/co-go"
	"github.com/republicprotocol/renex-swapper-go/domain/swap"
)

type storage struct {
	Adapter
}

type Adapter interface {
	ActiveSwaps() []swap.Request
}

func (storage *storage) LoadSwaps(swapRequestCh chan<- swap.Request, errCh chan<- error) {
	swaps := storage.ActiveSwaps()
	co.ParForAll(swaps, func(i int) {
		swapRequestCh <- swaps[i]
	})
}
