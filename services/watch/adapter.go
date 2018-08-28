package watch

import (
	"github.com/republicprotocol/renex-swapper-go/domains/match"
	"github.com/republicprotocol/renex-swapper-go/domains/order"
	"github.com/republicprotocol/renex-swapper-go/services/store"
	"github.com/republicprotocol/renex-swapper-go/services/swap"
)

// Wallet is an interface for the Atom Wallet Contract
type Adapter interface {
	// TODO: Idiomatic Go requires this method to be called "Match" instead of
	// "GetMatch"
	swap.SwapAdapter
	BuildAtoms(store.State, match.Match) (swap.Atom, swap.Atom, error)
	CheckForMatch(order.ID, int64) (match.Match, error)
}
