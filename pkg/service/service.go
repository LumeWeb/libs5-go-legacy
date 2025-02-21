package service

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"sync"
)

type P2PService interface {
	SelfConnectionUris() []*url.URL
	Peers() structs.Map
	ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer net.Peer) error
	OnNewPeer(peer net.Peer, verifyId bool) error
	GetNodeScore(nodeId *encoding.NodeId) (float64, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	SignMessageSimple(message []byte) ([]byte, error)
	AddPeer(peer net.Peer) error
	SendPublicPeersToPeer(peer net.Peer, peersToSend []net.Peer) error
	SendHashRequest(hash *encoding.Multihash, kinds []types.StorageLocationType) error
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
	Db() db.KVStore
}

type RegistryService interface {
	Set(sre protocol.SignedRegistryEntry, trusted bool, receivedFrom net.Peer) error
	BroadcastEntry(sre protocol.SignedRegistryEntry, receivedFrom net.Peer) error
	SendRegistryRequest(pk []byte) error
	Get(pk []byte) (protocol.SignedRegistryEntry, error)
	Listen(pk []byte, cb func(sre protocol.SignedRegistryEntry)) (func(), error)
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() db.KVStore
}

type HTTPService interface {
	GetHttpRouter() map[string]http.HandlerFunc
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	SetServices(services Services)
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() db.KVStore
}

type StorageService interface {
	GetCachedStorageLocations(hash *encoding.Multihash, kinds []types.StorageLocationType, local bool) (map[string]storage.StorageLocation, error)
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
	Db() db.KVStore
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Init(ctx context.Context) error
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() db.KVStore
	SetServices(services Services)
}

type ServicesSetter interface {
	SetServices(services Services)
}
type Services interface {
	P2P() P2PService
	Registry() RegistryService
	HTTP() HTTPService
	Storage() StorageService
	All() []Service
	Init(ctx context.Context) error
	IsStarted() bool
	IsStarting() bool
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ServiceParams struct {
	Logger *zap.Logger
	Config *config.NodeConfig
	Db     db.KVStore
}

type ServiceBase struct {
	logger   *zap.Logger
	config   *config.NodeConfig
	db       db.KVStore
	services Services
}

func NewServiceBase(logger *zap.Logger, config *config.NodeConfig, db db.KVStore) ServiceBase {
	return ServiceBase{logger: logger, config: config, db: db}
}

func (s *ServiceBase) SetServices(services Services) {
	s.services = services
}
func (s *ServiceBase) Services() Services {
	return s.services
}
func (s *ServiceBase) Logger() *zap.Logger {
	return s.logger
}
func (s *ServiceBase) Config() *config.NodeConfig {
	return s.config
}
func (s *ServiceBase) Db() db.KVStore {
	return s.db
}
