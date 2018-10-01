package swapper

import (
	"github.com/republicprotocol/renex-swapper-go/adapter/btc"
	"github.com/republicprotocol/renex-swapper-go/adapter/config"
	"github.com/republicprotocol/renex-swapper-go/adapter/eth"
	"github.com/republicprotocol/renex-swapper-go/adapter/keystore"
	"github.com/republicprotocol/renex-swapper-go/domain/swap"
	"github.com/republicprotocol/renex-swapper-go/domain/token"
	"github.com/republicprotocol/renex-swapper-go/service/logger"
	"github.com/republicprotocol/renex-swapper-go/service/swapper"
)

type swapperAdapter struct {
	keys   keystore.Keystore
	config config.Config
	logger logger.Logger
}

func New(keys keystore.Keystore, config config.Config, logger logger.Logger) swapper.Adapter {
	return &swapperAdapter{
		keys:   keys,
		config: config,
		logger: logger,
	}
}

func (adapter *swapperAdapter) BuildBinder(req swap.Request) (swapper.BlockchainBinder, error) {
	sendBlockchainBinder, err := buildBlockchainBinder(adapter.keys, adapter.config, adapter.logger, req.SendToken, req)
	if err != nil {
		return nil, err
	}
	receiveBlockchainBinder, err := buildBlockchainBinder(adapter.keys, adapter.config, adapter.logger, req.ReceiveToken, req)
	if err != nil {
		return nil, err
	}
	return &swapAdapter{
		sendBlockchainBinder:    sendBlockchainBinder,
		receiveBlockchainBinder: receiveBlockchainBinder,
	}, nil
}

func buildBlockchainBinder(keys keystore.Keystore, config config.Config, logger logger.Logger, t token.Token, req swap.Request) (swapper.BlockchainBinder, error) {
	switch t {
	case token.BTC:
		btcKey := keys.GetKey(t).(keystore.BitcoinKey)
		return btc.NewBitcoinAtom(config.Bitcoin, btcKey, logger, req)
	case token.ETH:
		ethKey := keys.GetKey(t).(keystore.EthereumKey)
		return eth.NewEthereumAtom(config.Ethereum, ethKey, logger, req)
	}
	return nil, token.ErrUnsupportedToken
}

type swapAdapter struct {
	sendBlockchainBinder    swapper.BlockchainBinder
	receiveBlockchainBinder swapper.BlockchainBinder
}

func (adapter *swapAdapter) Initiate() error {
	return adapter.sendBlockchainBinder.Initiate()
}

func (adapter *swapAdapter) Audit() error {
	return adapter.receiveBlockchainBinder.Audit()
}

func (adapter *swapAdapter) Redeem(secret [32]byte) error {
	return adapter.receiveBlockchainBinder.Redeem(secret)
}

func (adapter *swapAdapter) AuditSecret() ([32]byte, error) {
	return adapter.sendBlockchainBinder.AuditSecret()
}

func (adapter *swapAdapter) Refund() error {
	return adapter.sendBlockchainBinder.Refund()
}
