package keystore

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/republicprotocol/renex-swapper-go/utils"
)

type BitcoinKey struct {
	Network       *chaincfg.Params
	AddressString string
	Address       btcutil.Address
	PKHash        *btcutil.AddressPubKeyHash
	WIF           *btcutil.WIF
	PrivateKey    *btcec.PrivateKey
}

func (btcKey BitcoinKey) IsKey() {
}

func NewBitcoinKey(wifString string, network string) (BitcoinKey, error) {
	net := utils.GetChainParams(network)
	wif, err := btcutil.DecodeWIF(wifString)
	if err != nil {
		return BitcoinKey{}, err
	}

	pubKey, err := btcutil.NewAddressPubKey(wif.SerializePubKey(), net)
	if err != nil {
		return BitcoinKey{}, err
	}

	addrString := pubKey.EncodeAddress()
	addr, err := btcutil.DecodeAddress(addrString, net)
	if err != nil {
		return BitcoinKey{}, err
	}

	pkHash := pubKey.AddressPubKeyHash()

	return BitcoinKey{
		Network:       net,
		WIF:           wif,
		AddressString: addrString,
		Address:       addr,
		PKHash:        pkHash,
	}, nil
}

func RandomBitcoinKey(network string) (BitcoinKey, error) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return BitcoinKey{}, err
	}
	wif, err := btcutil.NewWIF(key, utils.GetChainParams(network), true)
	if err != nil {
		return BitcoinKey{}, err
	}
	return NewBitcoinKey(wif.String(), network)
}
