package utils

func EncodeEndian(value uint64, length int) []byte {
	res := make([]byte, length)

	for i := 0; i < length; i++ {
		res[i] = byte(value & 0xff)
		value = value >> 8
	}
	return res
}
func DecodeEndian(bytes []byte) uint64 {
	var total uint64 = 0
	var multiplier uint64 = 1

	for _, b := range bytes {
		total += uint64(b) * multiplier
		multiplier *= 256
	}

	return total
}
