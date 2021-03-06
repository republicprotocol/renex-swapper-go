package match

import (
	"encoding/json"
	"math/big"
)

// Match is the order match interface
type Match interface {
	PersonalOrderID() [32]byte
	ForeignOrderID() [32]byte
	SendValue() *big.Int
	ReceiveValue() *big.Int
	SendCurrency() uint32
	ReceiveCurrency() uint32
	Serialize() ([]byte, error)
}

type match struct {
	personalOrderID [32]byte
	foreignOrderID  [32]byte
	sendValue       *big.Int
	receiveValue    *big.Int
	sendCurrency    uint32
	receiveCurrency uint32
}

// NewMatch creates a new Match interface
func NewMatch(personalOrderID, foreignOrderID [32]byte, sendValue, receiveValue *big.Int, sendCurrency, receiveCurrency uint32) Match {
	return &match{
		personalOrderID: personalOrderID,
		foreignOrderID:  foreignOrderID,
		sendValue:       sendValue,
		receiveValue:    receiveValue,
		sendCurrency:    sendCurrency,
		receiveCurrency: receiveCurrency,
	}
}

// NewMatchFromBytes creates a new Match interface from byte array
func NewMatchFromBytes(data []byte) (Match, error) {
	var m match
	err := json.Unmarshal(data, &m)
	return &m, err
}

// PersonalOrderID returns the personal (caller's) order id of the order match.
func (match *match) PersonalOrderID() [32]byte {
	return match.personalOrderID
}

// ForeignOrderID returns the foreign (counter party's) order id of the order match.
func (match *match) ForeignOrderID() [32]byte {
	return match.foreignOrderID
}

// SendValue returns the value the caller has to send according to this order match.
func (match *match) SendValue() *big.Int {
	return match.sendValue
}

// ReceiveValue returns the value the caller has to receive according to this order match.
func (match *match) ReceiveValue() *big.Int {
	return match.receiveValue
}

// SendCurrency returns the currency the caller has to send according to this order match.
func (match *match) SendCurrency() uint32 {
	return match.sendCurrency
}

// SendCurrency returns the currency the caller has to send according to this order match.
func (match *match) ReceiveCurrency() uint32 {
	return match.receiveCurrency
}

// Serialize serializes the match object into a byte array.
func (match *match) Serialize() ([]byte, error) {
	return json.Marshal(*match)
}
