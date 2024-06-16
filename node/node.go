package node

import (
	"context"
	"go.lumeweb.com/libs5-go/config"
	"go.lumeweb.com/libs5-go/db"
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/protocol"
	"go.lumeweb.com/libs5-go/service"
	_default "go.lumeweb.com/libs5-go/service/default"
	"go.uber.org/zap"
)

type Node struct {
	nodeConfig *config.NodeConfig
	services   service.Services
}

func (n *Node) Services() service.Services {
	return n.services
}

func NewNode(config *config.NodeConfig, services service.Services) *Node {
	return &Node{
		nodeConfig: config,
		services:   services, // Services are passed in, not created here
	}
}

func (n *Node) IsStarted() bool {
	return n.services.IsStarted()
}

func (n *Node) Config() *config.NodeConfig {
	return n.nodeConfig
}

func (n *Node) Logger() *zap.Logger {
	if n.nodeConfig != nil {
		return n.nodeConfig.Logger
	}
	return nil
}

func (n *Node) Db() db.KVStore {
	if n.nodeConfig != nil {
		return n.nodeConfig.DB
	}
	return nil
}

func (n *Node) Start(ctx context.Context) error {
	protocol.RegisterProtocols()
	protocol.RegisterSignedProtocols()

	return n.services.Start(ctx)
}

func (n *Node) Init(ctx context.Context) error {
	return n.services.Init(ctx)
}

func (n *Node) Stop(ctx context.Context) error {
	return n.services.Stop(ctx)
}

func (n *Node) WaitOnConnectedPeers() {
	n.services.P2P().WaitOnConnectedPeers()
}

func (n *Node) NetworkId() string {
	return n.services.P2P().NetworkId()
}

func (n *Node) NodeId() *encoding.NodeId {
	return n.services.P2P().NodeId()
}

func DefaultNode(config *config.NodeConfig) *Node {
	params := service.ServiceParams{
		Logger: config.Logger,
		Config: config,
		Db:     config.DB,
	}

	// Initialize services first
	p2pService := _default.NewP2P(params)
	registryService := _default.NewRegistry(params)
	httpService := _default.NewHTTP(params)
	storageService := _default.NewStorage(params)

	// Aggregate services
	services := NewServices(ServicesParams{
		P2P:      p2pService,
		Registry: registryService,
		HTTP:     httpService,
		Storage:  storageService,
	})

	// Now create the node with the services
	return NewNode(config, services)
}
