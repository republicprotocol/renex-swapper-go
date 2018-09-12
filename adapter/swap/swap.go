package swap

import (
	"github.com/republicprotocol/renex-swapper-go/adapters/network"
	"github.com/republicprotocol/renex-swapper-go/services/logger"
	"github.com/republicprotocol/renex-swapper-go/services/swap"
	"github.com/republicprotocol/renex-swapper-go/services/watchdog"
)

func NewSwapAdapter(network network.Network, watchdog watchdog.WatchdogClient, logger logger.Logger) (swap.Adapter, error) {
	return nil, nil
}
