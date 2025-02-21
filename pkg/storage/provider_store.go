package storage

import (
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/storage/location"
)

type ProviderStore interface {
	CanProvide(hash *encoding.Blob, kind []location.StorageLocationType) bool
	Provide(hash *encoding.Blob, kind []location.StorageLocationType) (StorageLocation, error)
}
