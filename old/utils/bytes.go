package utils

import "bytes"

func ConcatBytes(slices ...[]byte) []byte {
	return bytes.Join(slices, nil)
}

func HashCode(bytes []byte) int {
	if len(bytes) < 4 {
		return 0
	}

	return int(bytes[0]) |
		int(bytes[1])<<8 |
		int(bytes[2])<<16 |
		int(bytes[3])<<24
}
