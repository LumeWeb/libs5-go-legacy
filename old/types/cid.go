package types

type CIDType int

const (
	CIDTypeRaw              CIDType = 0x26
	CIDTypeMetadataMedia    CIDType = 0xc5
	CIDTypeMetadataWebapp   CIDType = 0x59
	CIDTypeResolver         CIDType = 0x25
	CIDTypeUserIdentity     CIDType = 0x77
	CIDTypeBridge           CIDType = 0x3a
	CIDTypeEncryptedStatic  CIDType = 0xae
	CIDTypeEncryptedDynamic CIDType = 0xad
	CIDTypeDirectory        CIDType = 0x5d
)

var CIDTypeMap = map[CIDType]string{
	CIDTypeRaw:              "Raw",
	CIDTypeMetadataMedia:    "MetadataMedia",
	CIDTypeMetadataWebapp:   "MetadataWebapp",
	CIDTypeResolver:         "Resolver",
	CIDTypeUserIdentity:     "UserIdentity",
	CIDTypeBridge:           "Bridge",
	CIDTypeEncryptedStatic:  "EncryptedStatic",
	CIDTypeEncryptedDynamic: "EncryptedDynamic",
	CIDTypeDirectory:        "Directory",
}
