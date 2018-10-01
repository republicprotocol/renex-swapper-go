package swapper

import swapDomain "github.com/republicprotocol/renex-swapper-go/domain/swap"

type Adapter interface {
	BuildBinder(req swapDomain.Request) (BlockchainBinder, error)
}

type BlockchainBinder interface {
	Initiate() error
	Audit() error
	Redeem([32]byte) error
	AuditSecret() ([32]byte, error)
	Refund() error
}
