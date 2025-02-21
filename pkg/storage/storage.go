package storage

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"strconv"
	"time"
)

var (
	_ msgpack.CustomDecoder = (*StorageLocationMap)(nil)
	_ msgpack.CustomEncoder = (*StorageLocationMap)(nil)
	_ StorageLocation       = (*StorageLocationImpl)(nil)
	_ SignedStorageLocation = (*SignedStorageLocationImpl)(nil)
)

type StorageLocationMap map[int]NodeStorage
type NodeStorage map[string]NodeDetailsStorage
type NodeDetailsStorage map[int]interface{}

func (s *StorageLocationImpl) BytesURL() string {
	return s.parts[0]
}

func (s *StorageLocationImpl) OutboardBytesURL() string {
	if len(s.parts) == 1 {
		return s.parts[0] + ".obao"
	}
	return s.parts[1]
}

func (s *StorageLocationImpl) String() string {
	expiryDate := time.Unix(s.expiry, 0)
	return "StorageLocationImpl(" + strconv.Itoa(s.Type()) + ", " + fmt.Sprint(s.parts) + ", expiry: " + expiryDate.Format(time.RFC3339) + ")"
}

type SignedStorageLocationImpl struct {
	nodeID   *encoding.NodeId
	location StorageLocation
}

func NewSignedStorageLocation(NodeID *encoding.NodeId, Location StorageLocation) SignedStorageLocation {
	return &SignedStorageLocationImpl{
		nodeID:   NodeID,
		location: Location,
	}
}

func (ssl *SignedStorageLocationImpl) String() string {
	nodeString, _ := ssl.nodeID.ToString()

	if nodeString == "" {
		nodeString = "failed to decode node id"
	}

	return "SignedStorageLocationImpl(" + ssl.location.String() + ", " + nodeString + ")"
}

func (ssl *SignedStorageLocationImpl) NodeId() *encoding.NodeId {
	return ssl.nodeID
}
func (ssl *SignedStorageLocationImpl) Location() StorageLocation {
	return ssl.location
}

func (s *StorageLocationMap) DecodeMsgpack(dec *msgpack.Decoder) error {
	if *s == nil {
		*s = make(StorageLocationMap)
	}

	// Decode directly into a temp map
	temp := make(map[int]map[string]map[int]interface{})
	err := dec.Decode(&temp)
	if err != nil {
		return fmt.Errorf("error decoding msgpack: %w", err)
	}

	// Convert temp map to StorageLocationMap
	for k, v := range temp {
		nodeStorage, exists := (*s)[k]
		if !exists {
			nodeStorage = make(NodeStorage, len(v)) // preallocate if size is known
			(*s)[k] = nodeStorage
		}

		for nk, nv := range v {
			nodeDetailsStorage, exists := nodeStorage[nk]
			if !exists {
				nodeDetailsStorage = make(NodeDetailsStorage, len(nv)) // preallocate if size is known
				nodeStorage[nk] = nodeDetailsStorage
			}

			for ndk, ndv := range nv {
				nodeDetailsStorage[ndk] = ndv
			}
		}
	}

	return nil
}

func (s StorageLocationMap) EncodeMsgpack(enc *msgpack.Encoder) error {
	// Create a temporary map to hold the encoded data
	tempMap := make(map[int]map[string]map[int]interface{})

	// Populate the temporary map with data from storageLocationMap
	for storageKey, nodeStorages := range s {
		tempNodeStorages := make(map[string]map[int]interface{})
		for nodeId, nodeDetails := range nodeStorages {
			tempNodeStorages[nodeId] = nodeDetails
		}
		tempMap[storageKey] = tempNodeStorages
	}

	// Encode the temporary map using MessagePack
	return enc.Encode(tempMap)
}

func NewStorageLocationMap() StorageLocationMap {
	return StorageLocationMap{}
}

type StorageLocationProvider interface {
	Start() error
	Next() (SignedStorageLocation, error)
	All() ([]SignedStorageLocation, error)
	Upvote(uri SignedStorageLocation) error
	Downvote(uri SignedStorageLocation) error
}
