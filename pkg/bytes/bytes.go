package bytes

import "bytes"

func ConcatBytes(slices ...[]byte) []byte {
	return bytes.Join(slices, nil)
}
