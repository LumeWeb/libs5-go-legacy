package protocol

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/types"
)

var _ EncodeableMessage = (*HandshakeOpen)(nil)
var _ IncomingMessage = (*HandshakeOpen)(nil)

type HandshakeOpen struct {
	challenge []byte
	networkId string
	handshake []byte
	HandshakeRequirement
}

func (h *HandshakeOpen) SetHandshake(handshake []byte) {
	h.handshake = handshake
}

func (h HandshakeOpen) Challenge() []byte {
	return h.challenge
}

func (h HandshakeOpen) NetworkId() string {
	return h.networkId
}

func NewHandshakeOpen(challenge []byte, networkId string) *HandshakeOpen {
	ho := &HandshakeOpen{
		challenge: challenge,
		networkId: networkId,
	}

	ho.SetRequiresHandshake(false)

	return ho
}
func (h HandshakeOpen) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeUint(uint64(types.ProtocolMethodHandshakeOpen))
	if err != nil {
		return err
	}

	err = enc.EncodeBytes(h.challenge)
	if err != nil {
		return err
	}

	if h.networkId != "" {
		err = enc.EncodeString(h.networkId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HandshakeOpen) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error {
	handshake, err := dec.DecodeBytes()

	if err != nil {
		return err
	}

	h.handshake = handshake

	_, err = dec.PeekCode()

	networkId := ""

	if err != nil {
		if err.Error() != "EOF" {
			return err
		}
		h.networkId = networkId
		return nil
	}

	networkId, err = dec.DecodeString()
	if err != nil {
		return err
	}

	h.networkId = networkId

	return nil
}

func (h *HandshakeOpen) HandleMessage(message IncomingMessageData) error {
	peer := message.Peer
	mediator := message.Mediator

	if h.networkId != mediator.NetworkId() {
		return fmt.Errorf("Peer is in different network: %s", h.networkId)
	}

	handshake := NewHandshakeDoneRequest(h.handshake, types.SupportedFeatures, mediator.SelfConnectionUris())
	hsMessage, err := msgpack.Marshal(handshake)

	if err != nil {
		return err
	}

	secureMessage, err := mediator.SignMessageSimple(hsMessage)

	if err != nil {
		return err
	}

	err = peer.SendMessage(secureMessage)
	if err != nil {
		return err
	}

	return nil
}
