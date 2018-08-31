package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/republicprotocol/renex-swapper-go/utils"
)

var DefaultConfig = Config{
	Version:             "0.1.0",
	SupportedCurrencies: []string{"ETH", "BTC"},
	path:                utils.GetHome() + "/.swapper/config.json",
	mu:                  new(sync.RWMutex),
}

type Config struct {
	Version             string   `json:"version"`
	SupportedCurrencies []string `json:"supportedCurrencies"`
	AuthorizedAddresses []string `json:"authorizedAddresses"`

	mu   *sync.RWMutex
	path string
}

var ErrUnSupportedPriorityCode = errors.New("Unsupported Priority Code")

func NewConfig() Config {
	var config Config
	config.path = utils.GetHome() + "/.swapper/config.json"
	config.mu = new(sync.RWMutex)
	raw, err := ioutil.ReadFile(config.path)
	if err != nil {
		return DefaultConfig
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return DefaultConfig
	}
	return config
}

func (config *Config) Update() error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(config.path, data, 0644)
}

func (config *Config) GetVersion() string {
	return config.Version
}

func (config *Config) GetSupportedCurrencies() []string {
	return config.SupportedCurrencies
}

func (config *Config) GetAuthorizedAddresses() []common.Address {
	addrs := []common.Address{}
	for _, j := range config.AuthorizedAddresses {
		addrs = append(addrs, common.HexToAddress(j))
	}
	return addrs
}

func (config *Config) StoreLocation() string {
	return utils.GetHome() + "/.swapper/db"
}

func (config *Config) AuthorizeAddress(address string) error {
	config.AuthorizedAddresses = append(config.AuthorizedAddresses, address)
	return config.Update()
}
