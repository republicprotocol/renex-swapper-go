package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	Version             string   `json:"version"`
	SupportedCurrencies []string `json:"supportedCurrencies"`
	AuthorizedAddresses []string `json:"authorizedAddresses"`
	Watchdog            string   `json:"watchdogURL"`

	mu   *sync.RWMutex
	path string
}

var ErrUnSupportedPriorityCode = errors.New("Unsupported Priority Code")

func LoadConfig(path string) (Config, error) {
	var config Config
	config.path = path
	config.mu = new(sync.RWMutex)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	json.Unmarshal(raw, &config)
	return config, nil
}

func (config *Config) Update() error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(config.path, data, 700)
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

func (config *Config) StoreLocation() (string, error) {
	winHome := os.Getenv("userprofile")
	unixHome := os.Getenv("HOME")

	if unixHome != "" {
		return unixHome + "/.swapper/db", nil
	}

	if winHome != "" {
		return winHome + "/.swapper/db", nil
	}

	return "", fmt.Errorf("Unsupported Operating System")
}

func (config *Config) WatchdogURL() string {
	return config.Watchdog
}
