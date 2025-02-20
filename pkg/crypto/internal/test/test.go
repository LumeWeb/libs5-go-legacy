package test

import "encoding/hex"

// mustDecodeHex is a helper function to decode hex strings for test vectors
func MustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid test vector hex: " + s)
	}
	return b
}
