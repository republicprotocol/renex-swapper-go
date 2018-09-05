package swap

import (
	"github.com/republicprotocol/renex-swapper-go/domains/order"
)

type Adapter interface {
	Status([32]byte) string
	PutStatus([32]byte, string) error
	InitiateDetails([32]byte) (int64, [32]byte, error)
	PutInitiateDetails([32]byte, int64, [32]byte) error
	RedeemDetails([32]byte) ([32]byte, error)
	PutRedeemDetails([32]byte, [32]byte) error
	AtomDetails([32]byte) ([]byte, error)
	PutAtomDetails([32]byte, []byte) error
	PutRedeemable([32]byte) error
	Redeemed([32]byte) error
	SendOwnerAddress(order.ID, []byte) error
	ReceiveOwnerAddress(order.ID, int64) ([]byte, error)
	ReceiveSwapDetails(order.ID, int64) ([]byte, error)
	SendSwapDetails(order.ID, []byte) error
	ComplainDelayedAddressSubmission([32]byte) error
	ComplainDelayedRequestorInitiation([32]byte) error
	ComplainWrongRequestorInitiation([32]byte) error
	ComplainDelayedResponderInitiation([32]byte) error
	ComplainWrongResponderInitiation([32]byte) error
	ComplainDelayedRequestorRedemption([32]byte) error
	LogError([32]byte, string)
	LogInfo([32]byte, string)
	LogDebug([32]byte, string)
}
