package service

import (
	"context"
	"go.lumeweb.com/libs5-go/config"
	"go.lumeweb.com/libs5-go/db"
	"go.uber.org/zap"
)

type ServicesSetter interface {
	SetServices(services Services)
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Init(ctx context.Context) error
	Logger() *zap.Logger
	Config() *config.NodeConfig
	Db() db.KVStore
	ServicesSetter
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
