package p2p

import (
	"context"
	"old/config"
	"old/db"

	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/protocol"
	"go.lumeweb.com/libs5-go/pkg/service"
	"go.uber.org/zap"
)

type Node struct {
	nodeConfig      *config.NodeConfig // Ensure you get this from constructor.
	p2pService      service.P2PService
	registryService service.RegistryService
	httpService     service.HTTPService
	storageService  service.StorageService

	logger *zap.Logger
	db     db.KVStore
}

func NewNode(cfg *config.NodeConfig, p2p service.P2PService, registry service.RegistryService, http service.HTTPService, storage service.StorageService) (*Node, error) {

	node := &Node{
		nodeConfig:      cfg,
		p2pService:      p2p,
		registryService: registry,
		httpService:     http,
		storageService:  storage,
		logger:          cfg.Logger,
		db:              cfg.DB,
	}

	return node, nil
}
func (n *Node) Start(ctx context.Context) error { // This start just starts the service now
	protocol.RegisterProtocols()
	protocol.RegisterSignedProtocols() // You may have to move this elsewhere?

	err := n.p2pService.Start(ctx)
	if err != nil {
		return err
	}

	err = n.registryService.Start(ctx)
	if err != nil {
		return err
	}
	err = n.httpService.Start(ctx)
	if err != nil {
		return err
	}

	err = n.storageService.Start(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) GetP2PService() service.P2PService {
	return n.p2pService
}

func DefaultNode(config *config.NodeConfig) (*Node, error) {
	params := service.ServiceParams{
		Logger: config.Logger,
		Config: config,
		Db:     config.DB,
	}

	// Initialize services first
	p2pService, err := NewP2PService(config, crypto.NewDefaultCrypto(), config.DB, config.Logger, nil)
	if err != nil {
		return nil, err
	}
	registryService := _default.NewRegistry(params)
	httpService := _default.NewHTTP(params)
	storageService := _default.NewStorage(params)

	// Now create the node with the services
	return NewNode(config, p2pService, registryService, httpService, storageService)
}
func (n *Node) Logger() *zap.Logger {
	return n.nodeConfig.Logger
}
func (n *Node) GetRegistryService() service.RegistryService {
	return n.registryService
}
func (n *Node) GetHTTPService() service.HTTPService {
	return n.httpService
}
func (n *Node) GetStorageService() service.StorageService {
	return n.storageService
}
func (n *Node) Config() *config.NodeConfig {
	return n.nodeConfig
}
func (n *Node) GetDB() db.KVStore {
	return n.nodeConfig.DB
}
