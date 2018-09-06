package btc

import "fmt"

var ErrCompleteSignTransaction = NewErrSignTransaction(fmt.Errorf("incomplete signature"))
var ErrContractOutput = fmt.Errorf("transaction does not contain a contract output")
var ErrInitiated = fmt.Errorf("atomic swap already initiated")

func NewErrDecodeAddress(addr string, err error) error {
	return fmt.Errorf("failed to decode address (%s): %v", addr, err)
}

func NewErrDecodeScript(script []byte, err error) error {
	return fmt.Errorf("failed to decode script (%s): %v", script, err)
}

func NewErrSignTransaction(err error) error {
	return fmt.Errorf("failed to sign Transaction: %v", err)
}

func NewErrPublishTransaction(err error) error {
	return fmt.Errorf("failed to publish signed Transaction: %v", err)
}

func NewErrBuildScript(err error) error {
	return fmt.Errorf("failed to build bitcoin script: %v", err)
}

func NewErrInitiate(err error) error {
	return fmt.Errorf("failed to initiate: %v", err)
}
