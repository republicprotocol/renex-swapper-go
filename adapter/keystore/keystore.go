package keystore

import "github.com/republicprotocol/renex-swapper-go/domains/tokens"

type Key interface {
	IsKey()
}

type Keystore interface {
	GetKey(token tokens.Token) Key
}

type KeyMap map[tokens.Token]Key

type keystore struct {
	keyMap KeyMap
}

func New(keyMap KeyMap) Keystore {
	return &keystore{
		keyMap,
	}
}

// GetKey returns the key object of the given token
func (keystore *keystore) GetKey(token tokens.Token) Key {
	return keystore.keyMap[token]
}
