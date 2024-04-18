package storage

import (
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/types"
)

type ProviderStore interface {
	CanProvide(hash *encoding.Multihash, kind []types.StorageLocationType) bool
	Provide(hash *encoding.Multihash, kind []types.StorageLocationType) (StorageLocation, error)
}
