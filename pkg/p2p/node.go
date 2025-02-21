package p2p

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/http"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/protocol"
	"go.lumeweb.com/libs5-go/pkg/registry"
	"go.lumeweb.com/libs5-go/pkg/service"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.uber.org/zap"
	_default "old/service/default"
)

// Node represents the main application node that coordinates all services
type Node struct {
	config   *config.NodeConfig
	p2p      P2PService
	registry registry.RegistryService
	http     http.HTTPService
	storage  service.StorageService
	logger   *zap.Logger
	db       kv.KVStore
}

// NewNode creates a new Node instance with all required services
func NewNode(params NodeParams) (*Node, error) {
	return &Node{
		config:   params.Config,
		p2p:      params.P2P,
		registry: params.Registry,
		http:     params.HTTP,
		storage:  params.Storage,
		logger:   params.Config.Logger,
		db:       params.Config.DB,
	}, nil
}

// NodeParams contains all dependencies needed to create a new Node
type NodeParams struct {
	Config   *config.NodeConfig
	P2P      P2PService
	Registry registry.RegistryService
	HTTP     http.HTTPService
	Storage  service.StorageService
}

// Start initializes and starts all services
func (n *Node) Start(ctx context.Context) error {
	protocol.RegisterProtocols()
	protocol.RegisterSignedProtocols()

	services := []service.BaseService{
		n.p2p,
		n.registry,
		n.http,
		n.storage,
	}

	for _, svc := range services {
		if err := svc.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// DefaultNode creates a new Node with default service implementations
func DefaultNode(config *config.NodeConfig) (*Node, error) {
	p2p, err := NewP2PService(config, crypto.NewDefaultCrypto(), config.DB, config.Logger, nil)
	if err != nil {
		return nil, err
	}

	params := NodeParams{
		Config:   config,
		P2P:      p2p,
		Registry: registry.NewRegistry(config, config.Logger, config.DB),
		HTTP:     http.NewHTTP(config, config.Logger, config.DB),
		Storage:  storage.NewStorage(config, config.Logger, config.DB),
	}

	return NewNode(params)
}

// Logger returns the node's logger instance
func (n *Node) Logger() *zap.Logger {
	return n.logger
}

// Config returns the node's configuration
func (n *Node) Config() *config.NodeConfig {
	return n.config
}

// DB returns the node's database instance
func (n *Node) DB() kv.KVStore {
	return n.db
}

// P2P returns the P2P service instance
func (n *Node) P2P() service.P2PService {
	return n.p2p
}

// Registry returns the Registry service instance
func (n *Node) Registry() service.RegistryService {
	return n.registry
}

// HTTP returns the HTTP service instance
func (n *Node) HTTP() service.HTTPService {
	return n.http
}

// Storage returns the Storage service instance
func (n *Node) Storage() service.StorageService {
	return n.storage
}
