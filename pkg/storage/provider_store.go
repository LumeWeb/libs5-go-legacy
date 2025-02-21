package storage

import (
	"go.lumeweb.com/libs5-go/pkg/encoding"
)

type ProviderStore interface {
	CanProvide(hash *encoding.Blob, kind []StorageLocationType) bool
	Provide(hash *encoding.Blob, kind []StorageLocationType) (StorageLocation, error)
}
