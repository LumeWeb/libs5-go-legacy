package types

type RegistryType int

const (
	RegistryTypeCID          RegistryType = 0x5a
	RegistryTypeEncryptedCID RegistryType = 0x5e
)

var RegistryTypeMap = map[RegistryType]string{
	RegistryTypeCID:          "CID",
	RegistryTypeEncryptedCID: "EncryptedCID",
}

const RegistryMaxDataSize = 64
