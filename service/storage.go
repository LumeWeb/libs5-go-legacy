package service

import (
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/metadata"
	"github.com/LumeWeb/libs5-go/storage"
	"github.com/LumeWeb/libs5-go/types"
)

type StorageService interface {
	SetProviderStore(store storage.ProviderStore)
	ProviderStore() storage.ProviderStore
	GetCachedStorageLocations(hash *encoding.Multihash, kinds []types.StorageLocationType, local bool) (map[string]storage.StorageLocation, error)
	AddStorageLocation(hash *encoding.Multihash, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error
	DownloadBytesByHash(hash *encoding.Multihash) ([]byte, error)
	DownloadBytesByCID(cid *encoding.CID) ([]byte, error)
	GetMetadataByCID(cid *encoding.CID) (metadata.Metadata, error)
	ParseMetadata(bytes []byte, cid *encoding.CID) (metadata.Metadata, error)
	Service
}
