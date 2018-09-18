package swap

import "github.com/republicprotocol/renex-swapper-go/domain/order"

// Swapper is the interface for an atomic swapper object
type Swapper interface {
	NewSwap(orderID order.ID) (Swap, error)
}

type swapper struct {
	adapter SwapperAdapter
}

// NewSwapper returns a new Swapper instance
func NewSwapper(adapter SwapperAdapter) Swapper {
	return &swapper{
		adapter: adapter,
	}
}

func (swapper *swapper) NewSwap(orderID order.ID) (Swap, error) {
	personalAtom, foreignAtom, match, adapter, err := swapper.adapter.NewSwap(orderID)
	if err != nil {
		return nil, err
	}
	return &swap{
		personalAtom: personalAtom,
		foreignAtom:  foreignAtom,
		order:        match,
		Adapter:      adapter,
	}, nil
}
