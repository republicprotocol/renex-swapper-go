package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/republicprotocol/renex-swapper-go/utils"
)

type RepublicNetwork string

var RepublicNetworkNightly = RepublicNetwork("nightly")
var RepublicNetworkFalcon = RepublicNetwork("falcon")
var RepublicNetworkTestnet = RepublicNetwork("testnet")

var ErrUnknownRepublicNetwork = fmt.Errorf("Unknown Republic Network")

type Config struct {
	Network  RepublicNetwork `json:"network"`
	Ethereum EthereumNetwork `json:"ethereum"`
	Bitcoin  BitcoinNetwork  `json:"bitcoin"`
	Watchdog string          `json:"watchdog"`
	Ingress  string          `json:"ingress"`

	mu   *sync.RWMutex
	path string
}

var ErrUnSupportedPriorityCode = errors.New("Un Supported Priority Code")

func buildNetwork(net RepublicNetwork) Config {
	eth := NewEthNetwork(net)
	btc := NewBtcNetwork(net)
	return Config{
		Network:  net,
		Ethereum: eth,
		Bitcoin:  btc,
		path:     utils.GetHome() + "/.swapper/network.json",
		mu:       new(sync.RWMutex),
		Watchdog: fmt.Sprintf("renex-watchdog-%s.herokuapp.com", net),
	}
}

func NewNetwork(net RepublicNetwork) Config {
	var network Config
	network.mu = new(sync.RWMutex)
	network.path = utils.GetHome() + "/.swapper/network.json"
	raw, err := ioutil.ReadFile(network.path)
	if err != nil {
		return buildNetwork(net)
	}
	if err := json.Unmarshal(raw, &network); err != nil {
		return buildNetwork(net)
	}
	return network
}

func (network *Config) Update() error {
	data, err := json.Marshal(network)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(network.path, data, 0644)
}
