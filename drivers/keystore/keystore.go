package keystore

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/republicprotocol/renex-swapper-go/domains/tokens"

	"github.com/republicprotocol/renex-swapper-go/adapter/keystore"
	"github.com/republicprotocol/renex-swapper-go/utils"
)

// ErrKeyFileExists is returned when the keystore file exists, and the user is
// trying to overwrite it.
var ErrKeyFileExists = errors.New("Keystore file exists")

// ErrKeyFileDoesNotExist is returned when the keystore file doesnot exist, and
// the user is trying to read from it.
var ErrKeyFileDoesNotExist = errors.New("Keystore file doesnot exist")

// LoadFromFile
func LoadFromFile(repNetwork string, passphrase string) keystore.Keystore {
	var ethLoc, btcLoc string
	if passphrase == "" {
		ethLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/ethereum-%s-unsafe.json", repNetwork)
		btcLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/bitcoin-%s-unsafe.json", repNetwork)
	} else {
		ethLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/ethereum-%s.json", repNetwork)
		btcLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/bitcoin-%s.json", repNetwork)
	}

	ethNet, btcNet := getSpecificNetworks(repNetwork)
	ethKey, err := LoadKeyFromFile(ethLoc, "ethereum", ethNet, passphrase)
	if err != nil {
		panic("unimplemented")
	}
	btcKey, err := LoadKeyFromFile(btcLoc, "bitcoin", btcNet, passphrase)
	if err != nil {
		panic("unimplemented")
	}
	keyMap := keystore.KeyMap{}
	keyMap[tokens.TokenETH] = ethKey
	keyMap[tokens.TokenBTC] = btcKey
	return keystore.New(keyMap)
}

func GenerateRandom(repNetwork string) keystore.Keystore {
	ethNet, btcNet := getSpecificNetworks(repNetwork)
	ethKey, err := keystore.RandomEthereumKey(ethNet)
	if err != nil {
		panic("unimplemented")
	}
	btcKey, err := keystore.RandomBitcoinKey(btcNet)
	if err != nil {
		panic("unimplemented")
	}
	keyMap := keystore.KeyMap{}
	keyMap[tokens.TokenETH] = ethKey
	keyMap[tokens.TokenBTC] = btcKey
	return keystore.New(keyMap)
}

// GenerateFile
func GenerateFile(repNetwork string, passphrase string) error {
	var ethLoc, btcLoc string
	if passphrase == "" {
		ethLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/ethereum-%s-unsafe.json", repNetwork)
		btcLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/bitcoin-%s-unsafe.json", repNetwork)
	} else {
		ethLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/ethereum-%s.json", repNetwork)
		btcLoc = fmt.Sprintf(utils.GetHome()+"/.swapper/bitcoin-%s.json", repNetwork)
	}
	ethNet, btcNet := getSpecificNetworks(repNetwork)
	if err := StoreKeyToFile(ethLoc, "ethereum", ethNet, passphrase); err != nil {
		return err
	}
	return StoreKeyToFile(btcLoc, "bitcoin", btcNet, passphrase)
}

// LoadKeyFromFile loads a key from a file and tries to decrypt it using the
// given passphrase. If the passphrase is empty, then it tries to load an
// unencrypted key.
func LoadKeyFromFile(loc, chain, network, passphrase string) (keystore.Key, error) {
	data, err := ioutil.ReadFile(loc)
	if err != nil {
		return nil, ErrKeyFileDoesNotExist
	}
	return decodeKey(data, chain, network, passphrase)
}

// StoreKeyToFile stores a key to a file after encrypting it using the given
// passphrase. If the passphrase is empty, then it tries to load an unencrypted
// key.
func StoreKeyToFile(loc, chain, network, passphrase string) error {
	if _, err := ioutil.ReadFile(loc); err == nil {
		return ErrKeyFileExists
	}
	generatedKey, err := generateRandomKey(chain, network, passphrase)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(loc, generatedKey, 0400)
}

// LoadKeyFromNet loads a key from the network and tries to decrypt it using
// the given passphrase. If the  passphrase is empty, then it tries to load an
// unencrypted key.
func LoadKeyFromNet(url, chain, network, passphrase string) (keystore.Key, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, ErrKeyFileExists
	}
	if resp.StatusCode == 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return decodeKey(data, chain, network, passphrase)
	}
	return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
}

func generateRandomKey(chain, network, passphrase string) ([]byte, error) {
	switch chain {
	case "bitcoin":
		return GenerateRandomBitcoinKey(network, passphrase)
	case "ethereum":
		return GenerateRandomEthereumKey(passphrase)
	default:
		panic("unimplemented")
	}
}

func decodeKey(data []byte, chain, network, passphrase string) (keystore.Key, error) {
	switch chain {
	case "bitcoin":
		return DecodeBitcoinKey(data, network, passphrase)
	case "ethereum":
		return DecodeEthereumKey(data, network, passphrase)
	default:
		panic("unimplemented")
	}
}

func getSpecificNetworks(repNetwork string) (string, string) {
	switch repNetwork {
	case "nightly":
		return "kovan", "testnet"
	case "falcon":
		return "kovan", "testnet"
	case "testnet":
		return "kovan", "testnet"
	default:
		panic("unimplemented")
	}
}
