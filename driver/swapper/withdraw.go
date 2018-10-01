package swapper

import (
	"strconv"

	"github.com/republicprotocol/renex-swapper-go/adapter/btc"
	"github.com/republicprotocol/renex-swapper-go/adapter/eth"
	"github.com/republicprotocol/renex-swapper-go/adapter/keystore"
	"github.com/republicprotocol/renex-swapper-go/domain/token"
)

func (swapper *Swapper) Withdraw(t token.Token, to, value, fee string) error {
	switch t {
	case token.BTC:
		return swapper.withdrawBitcoin(to, value, fee)
	case token.ETH:
		return swapper.withdrawEthereum(to, value, fee)
	default:
		return token.ErrUnsupportedToken
	}
}

func (swapper *Swapper) withdrawBitcoin(to, valueString, feeString string) error {
	conn, err := btc.NewConnWithConfig(swapper.Bitcoin)
	if err != nil {
		return err
	}

	var value, fee int64
	if valueString == "" {
		value, err = conn.Balance(to)
	} else {
		value, err = strconv.ParseInt(valueString, 10, 64)
	}
	if err != nil {
		return err
	}

	if feeString == "" {
		fee = 3000
	} else {
		fee, err = strconv.ParseInt(feeString, 10, 64)
	}
	if err != nil {
		return err
	}

	return conn.Transfer(to, swapper.GetKey(token.BTC).(keystore.BitcoinKey), value, fee)
}

func (swapper *Swapper) withdrawEthereum(to, value, fee string) error {
	conn, err := eth.NewConnWithConfig(swapper.Ethereum)
	if err != nil {
		return err
	}

	return nil
}
