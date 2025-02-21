package storage

import (
	"go.lumeweb.com/libs5-go/pkg/encoding"
)

type ProviderStore interface {
	CanProvide(hash *encoding.Multihash, kind []StorageLocationType) bool
	Provide(hash *encoding.Multihash, kind []StorageLocationType) (StorageLocation, error)
}
