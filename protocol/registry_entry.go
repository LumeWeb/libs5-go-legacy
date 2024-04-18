package protocol

import (
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/internal/bases"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

var _ IncomingMessage = (*RegistryEntryRequest)(nil)
var _ EncodeableMessage = (*RegistryEntryRequest)(nil)

type RegistryEntryRequest struct {
	sre SignedRegistryEntry
	HandshakeRequirement
}

func NewEmptyRegistryEntryRequest() *RegistryEntryRequest {
	rer := &RegistryEntryRequest{}

	rer.SetRequiresHandshake(true)

	return rer
}
func NewRegistryEntryRequest(sre SignedRegistryEntry) *RegistryEntryRequest {
	return &RegistryEntryRequest{sre: sre}
}

func (s *RegistryEntryRequest) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeInt(int64(types.RecordTypeRegistryEntry))
	if err != nil {
		return err
	}

	err = enc.EncodeBytes(MarshalSignedRegistryEntry(s.sre))
	if err != nil {
		return err
	}

	return nil
}

func (s *RegistryEntryRequest) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error {
	sre, err := UnmarshalSignedRegistryEntry(message.Original)
	if err != nil {
		return err
	}

	s.sre = sre

	return nil
}

func (s *RegistryEntryRequest) HandleMessage(message IncomingMessageData) error {
	entry, err := encoding.CIDFromRegistryPublicKey(s.sre.PK())
	if err != nil {
		return err
	}
	pid, err := message.Peer.Id().ToString()
	if err != nil {
		return err
	}
	b64, err := bases.ToBase64Url(s.sre.PK())
	if err != nil {
		return err
	}
	message.Logger.Debug("Handling registry entry set request", zap.Any("entryCID", entry), zap.Any("entryBase64", b64), zap.Any("peer", pid))
	return message.Mediator.RegistrySet(s.sre, false, message.Peer)
}
