package client

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/republicprotocol/renex-swapper-go/adapters/configs/network"
)

type WatchdogClient interface {
	ComplainDelayedAddressSubmission([32]byte) error
	ComplainDelayedRequestorInitiation([32]byte) error
	ComplainWrongRequestorInitiation([32]byte) error
	ComplainDelayedResponderInitiation([32]byte) error
	ComplainWrongResponderInitiation([32]byte) error
	ComplainDelayedRequestorRedemption([32]byte) error
}
type watchdogHTTPClient struct {
	ipAddress string
}

// NewWatchdogHTTPClient creates a new WatchdogClient interface, that interacts
// with Watchdog over http.
func NewWatchdogHTTPClient(net network.Config) WatchdogClient {
	return &watchdogHTTPClient{
		ipAddress: net.Watchdog,
	}
}

func (client *watchdogHTTPClient) ComplainDelayedAddressSubmission(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) ComplainDelayedRequestorInitiation(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) ComplainWrongRequestorInitiation(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) ComplainDelayedResponderInitiation(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) ComplainWrongResponderInitiation(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) ComplainDelayedRequestorRedemption(orderID [32]byte) error {
	return client.watch(orderID)
}

func (client *watchdogHTTPClient) watch(orderID [32]byte) error {
	resp, err := http.Post(fmt.Sprintf("https://"+client.ipAddress+"/watch?orderID="+hex.EncodeToString(orderID[:])), "text", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	}
	return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
}
