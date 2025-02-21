package storage

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

const cacheBucketName = "object-cache"

var (
	_ service.Service        = (*StorageService)(nil)
	_ service.StorageService = (*StorageService)(nil)
)

var (
	ErrUnsupportedMetaFormat = errors.New("unsupported metadata format")
)

type StorageService struct {
	metadataCache structs.Map
	providerStore storage.ProviderStore
	bucket        db.KVStore
	service.ServiceBase
}

func NewStorage(params service.ServiceParams) *StorageService {
	return &StorageService{
		metadataCache: structs.NewMap(),
		ServiceBase:   service.NewServiceBase(params.Logger, params.Config, params.Db),
	}
}

func (s *StorageService) Start(ctx context.Context) error {

	bucket, err := s.Db().Bucket(cacheBucketName)
	if err != nil {
		return err
	}

	err = bucket.Open()
	if err != nil {
		return err
	}

	s.bucket = bucket

	return nil
}

func (s *StorageService) Stop(ctx context.Context) error {
	return nil
}

func (s *StorageService) Init(ctx context.Context) error {
	return nil
}

func (n *StorageService) SetProviderStore(store storage.ProviderStore) {
	n.providerStore = store
}

func (n *StorageService) ProviderStore() storage.ProviderStore {
	return n.providerStore
}

func (s *StorageService) GetCachedStorageLocations(hash *encoding.Multihash, kinds []types.StorageLocationType, local bool) (map[string]storage.StorageLocation, error) {
	locations := make(map[string]storage.StorageLocation)

	locationMap, err := s.readStorageLocationsFromDB(hash)
	if err != nil {
		return nil, err
	}

	if local {
		localLocation := s.getLocalStorageLocation(hash, kinds)
		if localLocation != nil {
			nodeIDStr, err := s.Services().P2P().NodeId().ToString()
			if err != nil {
				return nil, err
			}

			locations[nodeIDStr] = localLocation
		}
	}

	if len(locationMap) == 0 {
		return locations, nil
	}

	ts := time.Now().Unix()

	for _, t := range kinds {
		nodeMap, ok := (locationMap)[int(t)]
		if !ok {
			continue
		}

		for key, value := range nodeMap {
			expiry, ok := value[3].(int64)
			if !ok || expiry < ts {
				continue
			}

			addressesInterface, ok := value[1].([]interface{})
			if !ok {
				continue
			}

			// Create a slice to hold the strings
			addresses := make([]string, len(addressesInterface))

			// Convert each element to string
			for i, v := range addressesInterface {
				str, ok := v.(string)
				if !ok {
					// Handle the error, maybe skip this element or set a default value
					continue
				}
				addresses[i] = str
			}

			storageLocation := storage.NewStorageLocation(int(t), addresses, expiry)

			if providerMessage, ok := value[4].([]byte); ok {
				(storageLocation).SetProviderMessage(providerMessage)
			}

			locations[key] = storageLocation
		}
	}
	return locations, nil
}

func (s *StorageService) getLocalStorageLocation(hash *encoding.Multihash, kinds []types.StorageLocationType) storage.StorageLocation {
	if s.ProviderStore() != nil {
		if s.ProviderStore().CanProvide(hash, kinds) {
			location, _ := s.ProviderStore().Provide(hash, kinds)

			message := storage.PrepareProvideMessage(s.Services().P2P().Config().KeyPair, hash, location)

			location.SetProviderMessage(message)

			return location
		}
	}

	return nil
}

func (s *StorageService) readStorageLocationsFromDB(hash *encoding.Multihash) (storage.StorageLocationMap, error) {
	var locationMap storage.StorageLocationMap

	value, err := s.bucket.Get(hash.FullBytes())
	if err != nil {
		return nil, err
	}

	if value == nil {
		return storage.NewStorageLocationMap(), nil
	}

	locationMap = storage.NewStorageLocationMap()

	err = msgpack.Unmarshal(value, &locationMap)
	if err != nil {
		return nil, err
	}

	return locationMap, nil
}

func (s *StorageService) AddStorageLocation(hash *encoding.Multihash, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error {
	// Read existing storage locations
	locationDb, err := s.readStorageLocationsFromDB(hash)
	if err != nil {
		return err
	}

	nodeIdStr, err := nodeId.ToString()
	if err != nil {
		return err
	}

	// Get or create the inner map for the specific type
	innerMap, exists := locationDb[location.Type()]
	if !exists {
		innerMap = make(storage.NodeStorage, 1)
		innerMap[nodeIdStr] = make(storage.NodeDetailsStorage, 1)
	}

	// Create location map with new data
	locationMap := make(map[int]interface{}, 3)
	locationMap[1] = location.Parts()
	locationMap[3] = location.Expiry()
	locationMap[4] = message

	// Update the inner map with the new location
	innerMap[nodeIdStr] = locationMap
	locationDb[location.Type()] = innerMap

	// Serialize the updated map and store it in the database
	packedBytes, err := msgpack.Marshal(locationDb)
	if err != nil {
		return err
	}

	err = s.bucket.Put(hash.FullBytes(), packedBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageService) DownloadBytesByHash(hash *encoding.Multihash) ([]byte, error) {
	// Initialize the download URI provider
	dlUriProvider := provider.NewStorageLocationProvider(provider.StorageLocationProviderParams{
		Services: s.Services(),
		Hash:     hash,
		LocationTypes: []types.StorageLocationType{
			types.StorageLocationTypeFull,
			types.StorageLocationTypeFile,
		},
		ServiceParams: service.ServiceParams{
			Logger: s.Logger(),
			Config: s.Config(),
			Db:     s.Db(),
		},
	})
	err := dlUriProvider.Start()
	if err != nil {
		return nil, err
	}

	retryCount := 0
	for {
		dlUri, err := dlUriProvider.Next()
		if err != nil {
			return nil, err
		}

		s.Logger().Debug("Trying to download from", zap.String("url", dlUri.Location().BytesURL()))

		req := rq.Get(dlUri.Location().BytesURL())
		httpReq, err := req.ParseRequest()
		if err != nil {
			return nil, err
		}

		res, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			err := dlUriProvider.Downvote(dlUri)
			if err != nil {
				return nil, err
			}
			retryCount++
			if retryCount > 32 {
				return nil, errors.New("too many retries")
			}
			continue
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				s.Logger().Error("error closing body", zap.Error(err))
			}
		}(res.Body)

		if res.StatusCode != 200 {
			err := dlUriProvider.Downvote(dlUri)
			retryCount++
			if retryCount > 32 {
				return nil, errors.New("too many retries")
			}
			if err != nil {
				return nil, err
			}
			continue
		}

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return bodyBytes, nil
	}
}

func (s *StorageService) DownloadBytesByCID(cid *encoding.CID) (bytes []byte, err error) {
	bytes, err = s.DownloadBytesByHash(&cid.Hash)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (s *StorageService) GetMetadataByCID(cid *encoding.CID) (md metadata.Metadata, err error) {
	hashStr, err := cid.Hash.ToString()
	if err != nil {
		return nil, err
	}

	if s.metadataCache.Contains(hashStr) {
		md, _ := s.metadataCache.Get(hashStr)

		return md.(metadata.Metadata), nil
	}

	bytes, err := s.DownloadBytesByHash(&cid.Hash)
	if err != nil {
		return nil, err
	}

	md, err = s.ParseMetadata(bytes, cid)
	if err != nil {
		return nil, err
	}

	s.metadataCache.Put(hashStr, md)

	return md, nil
}

func (s *StorageService) ParseMetadata(bytes []byte, cid *encoding.CID) (metadata.Metadata, error) {
	var md metadata.Metadata

	switch cid.Type {
	case types.CIDTypeMetadataMedia, types.CIDTypeBridge: // Both cases use the same deserialization method
		md = metadata.NewEmptyMediaMetadata()

		err := msgpack.Unmarshal(bytes, md)
		if err != nil {
			return nil, err
		}
	case types.CIDTypeMetadataWebapp:
		md = metadata.NewEmptyWebAppMetadata()

		err := msgpack.Unmarshal(bytes, md)
		if err != nil {
			return nil, err
		}
	case types.CIDTypeDirectory:
		md = metadata.NewEmptyDirectoryMetadata()

		err := msgpack.Unmarshal(bytes, md)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnsupportedMetaFormat
	}

	return md, nil
}
