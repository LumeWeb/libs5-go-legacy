package registry

import (
	"context"
	ed25519p "crypto/ed25519"
	"errors"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/protocol"
	"go.lumeweb.com/libs5-go/pkg/registry/entry"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
)

var (
	_ entry.SignedRegistryEntry = (*SignedRegistryEntryImpl)(nil)
	_ entry.SignedRegistryEntry = (*SignedRegistryEntryImpl)(nil)
)

const RegistryMaxDataSize = 64

type RegistryService interface {
	Set(sre entry.SignedRegistryEntry, trusted bool, receivedFrom transport.Peer) error
	BroadcastEntry(sre entry.SignedRegistryEntry, receivedFrom transport.Peer) error
	SendRegistryRequest(pk []byte) error
	Get(pk []byte) (entry.SignedRegistryEntry, error)
	Listen(pk []byte, cb func(sre entry.SignedRegistryEntry)) (func(), error)
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	Logger() *zap.Logger
	Config() *config.NodeConfig
	DB() kv.KVStore
}

type SignedRegistryEntryImpl struct {
	pk        []byte
	revision  uint64
	data      []byte
	signature []byte
}

func (s *SignedRegistryEntryImpl) Verify() bool {
	return VerifyRegistryEntry(s)
}

func (s *SignedRegistryEntryImpl) PK() []byte {
	return s.pk
}

func (s *SignedRegistryEntryImpl) SetPK(pk []byte) {
	s.pk = pk
}

func (s *SignedRegistryEntryImpl) Revision() uint64 {
	return s.revision
}

func (s *SignedRegistryEntryImpl) SetRevision(revision uint64) {
	s.revision = revision
}

func (s *SignedRegistryEntryImpl) Data() []byte {
	return s.data
}

func (s *SignedRegistryEntryImpl) SetData(data []byte) {
	s.data = data
}

func (s *SignedRegistryEntryImpl) Signature() []byte {
	return s.signature
}

func (s *SignedRegistryEntryImpl) SetSignature(signature []byte) {
	s.signature = signature
}

func NewSignedRegistryEntry(pk []byte, revision uint64, data []byte, signature []byte) entry.SignedRegistryEntry {
	return &SignedRegistryEntryImpl{
		pk:        pk,
		revision:  revision,
		data:      data,
		signature: signature,
	}
}

type RegistryEntryImpl struct {
	kp       *crypto.KeyPairEd25519
	data     []byte
	revision uint64
}

func NewRegistryEntry(kp *crypto.KeyPairEd25519, data []byte, revision uint64) entry.RegistryEntry {
	return &RegistryEntryImpl{
		kp:       kp,
		data:     data,
		revision: revision,
	}
}

func (r *RegistryEntryImpl) Sign() entry.SignedRegistryEntry {
	return SignRegistryEntry(r.kp, r.data, r.revision)
}

func SignRegistryEntry(kp *crypto.KeyPairEd25519, data []byte, revision uint64) entry.SignedRegistryEntry {
	buffer := MarshalRegistryEntry(nil, data, revision)

	privateKey := kp.ExtractBytes()
	signature := ed25519p.Sign(privateKey, buffer)

	return NewSignedRegistryEntry(kp.PublicKey(), revision, data, signature)
}
func VerifyRegistryEntry(sre entry.SignedRegistryEntry) bool {
	buffer := MarshalRegistryEntry(nil, sre.Data(), sre.Revision())
	publicKey := sre.PK()[1:]

	return ed25519p.Verify(publicKey, buffer, sre.Signature())
}

func MarshalSignedRegistryEntry(sre entry.SignedRegistryEntry) []byte {
	buffer := MarshalRegistryEntry(sre.PK(), sre.Data(), sre.Revision())
	buffer = append(buffer, sre.Signature()...)

	return buffer
}

func MarshalRegistryEntry(pk []byte, data []byte, revision uint64) []byte {
	var buffer []byte
	buffer = append(buffer, byte(protocol.RecordTypeRegistryEntry))

	if pk != nil {
		buffer = append(buffer, pk...)
	}

	revBytes := encoding.EncodeEndian(revision, 8)
	buffer = append(buffer, revBytes...)

	buffer = append(buffer, byte(len(data)))
	buffer = append(buffer, data...)

	return buffer
}

func UnmarshalSignedRegistryEntry(event []byte) (sre entry.SignedRegistryEntry, err error) {
	if len(event) < 43 {
		return nil, errors.New("Invalid registry entry")
	}

	dataLength := int(event[42])
	if len(event) < 43+dataLength {
		return nil, errors.New("Invalid registry entry")
	}

	pk := event[1:34]
	revisionBytes := event[34:42]
	revision := encoding.DecodeEndian(revisionBytes)
	signatureStart := 43 + dataLength
	var signature []byte

	if signatureStart < len(event) {
		signature = event[signatureStart:]
	} else {
		return nil, errors.New("Invalid signature")
	}

	return NewSignedRegistryEntry(pk, revision, event[43:signatureStart], signature), nil
}

type SignedRegistryEntry interface {
	PK() []byte
	Revision() uint64
	Data() []byte
	Signature() []byte
	SetPK(pk []byte)
	SetRevision(revision uint64)
	SetData(data []byte)
	SetSignature(signature []byte)
	Verify() bool
}
type RegistryEntry interface {
	Sign() SignedRegistryEntry
}
