package metadata

import "github.com/LumeWeb/libs5-go/encoding"

type UserIdentityMetadata struct {
	UserID         *encoding.CID
	Details        UserIdentityMetadataDetails
	SigningKeys    []UserIdentityPublicKey
	EncryptionKeys []UserIdentityPublicKey
	Links          map[int]*encoding.CID
}
