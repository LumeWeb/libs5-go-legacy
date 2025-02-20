package types

type HashType int

const (
	HashTypeBlake3  HashType = 0x1f
	HashTypeEd25519 HashType = 0xed
)

var HashTypeMap = map[HashType]string{
	HashTypeBlake3:  "Blake3",
	HashTypeEd25519: "Ed25519",
}
