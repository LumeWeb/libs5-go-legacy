package registry

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/protocol"
	"go.lumeweb.com/libs5-go/pkg/service"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"time"
)

const registryBucketName = "registry"

var _ service.RegistryService = (*RegistryServiceDefault)(nil)

type RegistryServiceDefault struct {
	streams structs.Map
	subs    structs.Map
	bucket  kv.KVStore
	logger  *zap.Logger
	config  *config.NodeConfig
	db      kv.KVStore
	p2p     service.P2PService
}

func (r *RegistryServiceDefault) Logger() *zap.Logger {
	return r.logger
}

func (r *RegistryServiceDefault) Config() *config.NodeConfig {
	return r.config
}

func (r *RegistryServiceDefault) DB() kv.KVStore {
	return r.db
}

func (r *RegistryServiceDefault) Start(ctx context.Context) error {
	return nil
}

func (r *RegistryServiceDefault) Stop(ctx context.Context) error {
	return nil
}

func (r *RegistryServiceDefault) Init(ctx context.Context) error {
	bucket, err := r.db.Bucket(registryBucketName)
	if err != nil {
		return err
	}

	err = bucket.Open()
	if err != nil {
		return err
	}

	r.bucket = bucket

	return nil
}

func NewRegistry(config *config.NodeConfig, logger *zap.Logger, db kv.KVStore, p2p service.P2PService) *RegistryServiceDefault {
	return &RegistryServiceDefault{
		streams: structs.NewMap(),
		subs:    structs.NewMap(),
		logger:  logger,
		config:  config,
		db:      db,
		p2p:     p2p,
	}
}
func (r *RegistryServiceDefault) Set(sre registry.SignedRegistryEntry, trusted bool, receivedFrom transport.Peer) error {
	hash := encoding.NewMultihash(sre.PK())
	hashString, err := hash.ToString()
	if err != nil {
		return err
	}
	pid, err := receivedFrom.Id().ToString()
	if err != nil {
		return err
	}
	r.Logger().Debug("[registry] set", zap.String("pk", hashString), zap.Uint64("revision", sre.Revision()), zap.String("receivedFrom", pid))

	if !trusted {
		if len(sre.PK()) != 33 {
			return errors.New("Invalid pubkey")
		}
		if int(sre.PK()[0]) != int(crypto.HashTypeEd25519) {
			return errors.New("Only ed25519 keys are supported")
		}
		if sre.Revision() < 0 || sre.Revision() > 281474976710656 {
			return errors.New("Invalid revision")
		}
		if len(sre.Data()) > protocol.RegistryMaxDataSize {
			return errors.New("Data too long")
		}

		if !sre.Verify() {
			return errors.New("Invalid signature found")
		}
	}

	existingEntry, err := r.getFromDB(sre.PK())
	if err != nil {
		return err
	}

	if existingEntry != nil {
		if receivedFrom != nil {
			if existingEntry.Revision() == sre.Revision() {
				return nil
			} else if existingEntry.Revision() > sre.Revision() {
				updateMessage := protocol.MarshalSignedRegistryEntry(existingEntry)
				err := receivedFrom.SendMessage(updateMessage)
				if err != nil {
					return err
				}
				return nil
			}
		}

		if existingEntry.Revision() >= sre.Revision() {
			return errors.New("Revision number too low")
		}
	}

	key := encoding.NewMultihash(sre.PK())
	keyString, err := key.ToString()
	if err != nil {
		return err
	}

	eventObj, ok := r.streams.Get(keyString)
	if ok {
		event := eventObj.(*emitter.Emitter)
		go event.Emit("fire", sre)
	}

	err = r.bucket.Put(sre.PK(), protocol.MarshalSignedRegistryEntry(sre))
	if err != nil {
		return err
	}

	err = r.BroadcastEntry(sre, receivedFrom)
	if err != nil {
		return err
	}

	return nil
}
func (r *RegistryServiceDefault) BroadcastEntry(sre protocol.SignedRegistryEntry, receivedFrom transport.Peer) error {
	hash := encoding.NewMultihash(sre.PK())
	hashString, err := hash.ToString()
	if err != nil {
		return err
	}
	pid, err := receivedFrom.Id().ToString()
	if err != nil {
		return err
	}
	r.Logger().Debug("[registry] broadcastEntry", zap.String("pk", hashString), zap.Uint64("revision", sre.Revision()), zap.String("receivedFrom", pid))
	updateMessage := protocol.MarshalSignedRegistryEntry(sre)

	for _, p := range r.p2p.Peers().Values() {
		peer, ok := p.(transport.Peer)
		if !ok {
			continue
		}
		if receivedFrom == nil || peer.Id().Equals(receivedFrom.Id()) {
			err := peer.SendMessage(updateMessage)
			if err != nil {
				pid, err := peer.Id().ToString()
				if err != nil {
					return err
				}
				r.Logger().Error("Failed to send registry broadcast", zap.String("peer", pid), zap.Error(err))
				return err
			}
		}
	}

	return nil
}
func (r *RegistryServiceDefault) SendRegistryRequest(pk []byte) error {
	query := protocol.NewRegistryQuery(pk)

	request, err := msgpack.Marshal(query)
	if err != nil {
		return err
	}

	// Iterate over peers and send the request
	for _, peerVal := range r.p2p.Peers().Values() {
		peer, ok := peerVal.(transport.Peer)
		if !ok {
			continue
		}
		err := peer.SendMessage(request)
		if err != nil {
			pid, err := peer.Id().ToString()
			if err != nil {
				return err
			}
			r.logger.Error("Failed to send registry request", zap.String("peer", pid), zap.Error(err))
			return err
		}
	}

	return nil
}
func (r *RegistryServiceDefault) Get(pk []byte) (protocol.SignedRegistryEntry, error) {
	key := encoding.NewMultihash(pk)
	keyString, err := key.ToString()
	if err != nil {
		return nil, err
	}

	if r.subs.Contains(keyString) {
		r.Logger().Debug("[registry] get (cached)", zap.String("key", keyString))
		res, err := r.getFromDB(pk)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}

		err = r.SendRegistryRequest(pk)
		if err != nil {
			return nil, err
		}
		time.Sleep(200 * time.Millisecond)
		return r.getFromDB(pk)
	}

	err = r.SendRegistryRequest(pk)
	if err != nil {
		return nil, err
	}
	r.subs.Put(keyString, key)
	if !r.streams.Contains(keyString) {
		event := &emitter.Emitter{}
		r.streams.Put(keyString, event)
	}

	res, err := r.getFromDB(pk)
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res, nil
	}

	r.Logger().Debug("[registry] get (cached)", zap.String("key", keyString))
	for i := 0; i < 200; i++ {
		time.Sleep(10 * time.Millisecond)
		res, err = r.getFromDB(pk)
		if err != nil {
			return nil, err
		}
		if res != nil {
			break
		}
	}

	return res, nil
}

func (r *RegistryServiceDefault) Listen(pk []byte, cb func(sre protocol.SignedRegistryEntry)) (func(), error) {
	key, err := encoding.NewMultihash(pk).ToString()
	if err != nil {
		return nil, err
	}

	cbProxy := func(event *emitter.Event) {
		sre, ok := event.Args[0].(protocol.SignedRegistryEntry)
		if !ok {
			r.Logger().Error("Failed to cast event to SignedRegistryEntry")
			return
		}

		cb(sre)
	}

	if !r.streams.Contains(key) {
		em := emitter.New(0)
		r.streams.Put(key, em)
		err := r.SendRegistryRequest(pk)
		if err != nil {
			return nil, err
		}
	}
	streamVal, _ := r.streams.Get(key)
	stream := streamVal.(*emitter.Emitter)
	channel := stream.On("fire", cbProxy)

	return func() {
		stream.Off("fire", channel)
	}, nil
}

func (r *RegistryServiceDefault) getFromDB(pk []byte) (sre protocol.SignedRegistryEntry, err error) {
	value, err := r.bucket.Get(pk)
	if err != nil {
		return nil, err
	}

	if value != nil {
		sre, err = protocol.UnmarshalSignedRegistryEntry(value)
		if err != nil {
			return nil, err
		}

		return sre, nil
	}

	return nil, nil
}
