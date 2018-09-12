package config

import (
	"fmt"

	"github.com/republicprotocol/renex-swapper-go/adapter/config"
	"github.com/republicprotocol/renex-swapper-go/utils"
)

// Global is the default global config
var Global = config.Config{
	Version:             "0.1.0",
	SupportedCurrencies: []string{"ETH", "BTC"},
	StoreLocation:       utils.GetHome() + "/.swapper/db",
	Ethereum:            EthereumKovan,
	Bitcoin:             BitcoinTestnet,
}

// EthereumKovan is the ethereum config object on the kovan testnet
var EthereumKovan = config.EthereumNetwork{
	Network: "kovan",
	URL:     "https://kovan.infura.io",
}

// BitcoinTestnet is the bitcoin config object on the bitcoin testnet
var BitcoinTestnet = config.BitcoinNetwork{
	Network: "testnet",
	URL:     "https://testnet.blockchain.info",
}

// NewRenExNetwork creates a RenEx config object for the given RenEx network
func NewRenExNetwork(net string) config.RenExNetwork {
	return config.RenExNetwork{
		Network:  net,
		Watchdog: fmt.Sprintf("renex-watchdog-%s.herokuapp.com", net),
		Ingress:  fmt.Sprintf("renex-ingress-%s.herokuapp.com", net),
	}
}

// New creates a new config object from the config data object
func New(net string) config.Config {
	conf := Global
	conf.RenEx = NewRenExNetwork(net)
	switch net {
	case "nightly":
		conf.RenEx.Orderbook = ""
		conf.RenEx.Settlement = ""
		conf.RenEx.Swapper = ""
		return conf
	case "falcon":
		conf.RenEx.Orderbook = ""
		conf.RenEx.Settlement = ""
		conf.RenEx.Swapper = ""
		return conf
	case "testnet":
		conf.RenEx.Orderbook = ""
		conf.RenEx.Settlement = ""
		conf.RenEx.Swapper = ""
		return conf
	default:
		panic("unimplemented")
	}
}
