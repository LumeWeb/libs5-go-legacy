package storage

import (
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/types"
)

type ProviderStore interface {
	CanProvide(hash *encoding.Multihash, kind []types.StorageLocationType) bool
	Provide(hash *encoding.Multihash, kind []types.StorageLocationType) (StorageLocation, error)
}
