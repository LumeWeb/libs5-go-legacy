package service

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/registry"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"old/metadata"
	"sync"
)

// BaseService defines the common operations all services must implement
type BaseService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Init(ctx context.Context) error
}

type P2PService interface {
	BaseService
	SelfConnectionUris() []*url.URL
	Peers() structs.Map
	ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer transport.Peer) error
	OnNewPeer(peer transport.Peer, verifyId bool) error
	GetNodeScore(nodeId *encoding.NodeId) (float64, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	SignMessageSimple(message []byte) ([]byte, error)
	AddPeer(peer transport.Peer) error
	SendPublicPeersToPeer(peer transport.Peer, peersToSend []transport.Peer) error
	SendHashRequest(hash *encoding.Multihash, kinds []storage.StorageLocationType) error
	UpVote(nodeId *encoding.NodeId) error
	DownVote(nodeId *encoding.NodeId) error
	NodeId() *encoding.NodeId
	WaitOnConnectedPeers()
	ConnectionTracker() *sync.WaitGroup
	NetworkId() string
	HashQueryRoutingTable() structs.Map
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() kv.KVStore
}

type RegistryService interface {
	Set(sre registry.SignedRegistryEntry, trusted bool, receivedFrom transport.Peer) error
	BroadcastEntry(sre registry.SignedRegistryEntry, receivedFrom transport.Peer) error
	SendRegistryRequest(pk []byte) error
	Get(pk []byte) (registry.SignedRegistryEntry, error)
	Listen(pk []byte, cb func(sre registry.SignedRegistryEntry)) (func(), error)
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() kv.KVStore
}

type HTTPService interface {
	GetHttpRouter() map[string]http.HandlerFunc
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() kv.KVStore
}

type StorageService interface {
	GetCachedStorageLocations(hash *encoding.Multihash, kinds []storage.StorageLocationType, local bool) (map[string]storage.StorageLocation, error)
	AddStorageLocation(hash *encoding.Multihash, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error
	DownloadBytesByHash(hash *encoding.Multihash) ([]byte, error)
	DownloadBytesByCID(cid *encoding.CID) ([]byte, error)
	GetMetadataByCID(cid *encoding.CID) (metadata.Metadata, error)
	ParseMetadata(bytes []byte, cid *encoding.CID) (metadata.Metadata, error)
	SetProviderStore(store storage.ProviderStore)
	ProviderStore() storage.ProviderStore
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() kv.KVStore
}

