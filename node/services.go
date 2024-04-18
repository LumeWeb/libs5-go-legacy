package node

import (
	"context"
	"github.com/LumeWeb/libs5-go/service"
)

var (
	_ service.Services = (*ServicesImpl)(nil)
)

type ServicesParams struct {
	P2P      service.P2PService
	Registry service.RegistryService
	HTTP     service.HTTPService
	Storage  service.StorageService
}

type ServicesImpl struct {
	p2p      service.P2PService
	registry service.RegistryService
	http     service.HTTPService
	storage  service.StorageService
	started  bool
	starting bool
}

func (s *ServicesImpl) HTTP() service.HTTPService {
	return s.http
}

func (s *ServicesImpl) Storage() service.StorageService {
	return s.storage
}

func (s *ServicesImpl) All() []service.Service {
	services := make([]service.Service, 0)
	services = append(services, s.p2p)
	services = append(services, s.registry)
	services = append(services, s.http)
	services = append(services, s.storage)

	return services
}

func (s *ServicesImpl) Registry() service.RegistryService {
	return s.registry
}

func NewServices(params ServicesParams) service.Services {
	sc := &ServicesImpl{
		p2p:      params.P2P,
		registry: params.Registry,
		http:     params.HTTP,
		storage:  params.Storage,
		started:  false,
	}

	for _, svc := range sc.All() {
		svc.SetServices(sc)
	}

	return sc
}

func (s *ServicesImpl) P2P() service.P2PService {
	return s.p2p
}

func (s *ServicesImpl) IsStarted() bool {
	return s.started
}

func (s *ServicesImpl) IsStarting() bool {
	return s.starting
}

func (s *ServicesImpl) Init(ctx context.Context) error {
	err := s.p2p.Config().DB.Open()
	if err != nil {
		return err
	}

	for _, svc := range s.All() {
		err := svc.Init(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ServicesImpl) Start(ctx context.Context) error {
	s.starting = true

	for _, svc := range s.All() {
		err := svc.Start(ctx)
		if err != nil {
			return err
		}
	}

	s.started = true
	s.starting = false

	return nil
}
func (s *ServicesImpl) Stop(ctx context.Context) error {
	for _, svc := range s.All() {
		err := svc.Stop(ctx)
		if err != nil {
			return err
		}
	}

	s.started = false

	return nil
}
