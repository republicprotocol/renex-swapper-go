package utils

import "errors"

// ToBytes32 converts dynamic byte array of size 32 to static byte array of size 32
func ToBytes32(b []byte) ([32]byte, error) {
	bytes32 := [32]byte{}
	if len(b) != 32 {
		return bytes32, errors.New("Length mismatch")
	}
	copy(bytes32[:], b[:32])
	return bytes32, nil
}

// ToBytes65 converts dynamic byte array of size 65 to static byte array of size 65
func ToBytes65(b []byte) ([65]byte, error) {
	bytes65 := [65]byte{}
	if len(b) != 65 {
		return bytes65, errors.New("Length mismatch")
	}
	copy(bytes65[:], b[:65])
	return bytes65, nil
}
