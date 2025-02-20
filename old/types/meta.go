package types

type MetadataExtension int

const (
	MetadataExtensionLicenses           MetadataExtension = 0x0B
	MetadataExtensionDonationKeys       MetadataExtension = 0x0C
	MetadataExtensionWikidataClaims     MetadataExtension = 0x0D
	MetadataExtensionLanguages          MetadataExtension = 0x0E
	MetadataExtensionSourceUris         MetadataExtension = 0x0F
	MetadataExtensionUpdateCID          MetadataExtension = 0x10
	MetadataExtensionPreviousVersions   MetadataExtension = 0x11
	MetadataExtensionTimestamp          MetadataExtension = 0x12
	MetadataExtensionTags               MetadataExtension = 0x13
	MetadataExtensionCategories         MetadataExtension = 0x14
	MetadataExtensionViewTypes          MetadataExtension = 0x15
	MetadataExtensionBasicMediaMetadata MetadataExtension = 0x16
	MetadataExtensionBridge             MetadataExtension = 0x17
	MetadataExtensionOriginalTimestamp  MetadataExtension = 0x18
	MetadataExtensionRoutingHints       MetadataExtension = 0x19
)

var MetadataMap = map[MetadataExtension]string{
	MetadataExtensionLicenses:           "MetadataExtensionLicenses",
	MetadataExtensionDonationKeys:       "MetadataExtensionDonationKeys",
	MetadataExtensionWikidataClaims:     "MetadataExtensionWikidataClaims",
	MetadataExtensionLanguages:          "MetadataExtensionLanguages",
	MetadataExtensionSourceUris:         "MetadataExtensionSourceUris",
	MetadataExtensionUpdateCID:          "MetadataExtensionUpdateCID",
	MetadataExtensionPreviousVersions:   "MetadataExtensionPreviousVersions",
	MetadataExtensionTimestamp:          "MetadataExtensionTimestamp",
	MetadataExtensionTags:               "MetadataExtensionTags",
	MetadataExtensionCategories:         "MetadataExtensionCategories",
	MetadataExtensionViewTypes:          "MetadataExtensionViewTypes",
	MetadataExtensionBasicMediaMetadata: "MetadataExtensionBasicMediaMetadata",
	MetadataExtensionBridge:             "MetadataExtensionBridge",
	MetadataExtensionOriginalTimestamp:  "MetadataExtensionOriginalTimestamp",
	MetadataExtensionRoutingHints:       "MetadataExtensionRoutingHints",
}

const MetadataMagicByte = 0x5f

type MetadataType uint8

const (
	MetadataTypeMedia        MetadataType = 0x02
	MetadataTypeWebApp       MetadataType = 0x03
	MetadataTypeDirectory    MetadataType = 0x04
	MetadataTypeProof        MetadataType = 0x05
	MetadataTypeUserIdentity MetadataType = 0x07
)

var MetadataTypeMap = map[string]MetadataType{
	"Media":        MetadataTypeMedia,
	"WebApp":       MetadataTypeWebApp,
	"Directory":    MetadataTypeDirectory,
	"Proof":        MetadataTypeProof,
	"UserIdentity": MetadataTypeUserIdentity,
}

type MetadataProofType uint8

const (
	MetadataProofTypeSignature MetadataProofType = 0x01
	MetadataProofTypeTimestamp MetadataProofType = 0x02
)

var MetadataProofTypeMap = map[string]MetadataProofType{
	"Signature": MetadataProofTypeSignature,
	"Timestamp": MetadataProofTypeTimestamp,
}
