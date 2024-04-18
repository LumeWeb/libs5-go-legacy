package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/LumeWeb/libs5-go/utils"
	"github.com/vmihailenco/msgpack/v5"
	"net/url"
)

var _ IncomingMessageSigned = (*HandshakeDone)(nil)
var _ EncodeableMessage = (*HandshakeDone)(nil)

type HandshakeDone struct {
	challenge         []byte
	networkId         string
	supportedFeatures int
	connectionUris    []*url.URL
	handshake         []byte
	HandshakeRequirement
}

func NewHandshakeDoneRequest(handshake []byte, supportedFeatures int, connectionUris []*url.URL) *HandshakeDone {
	ho := &HandshakeDone{
		handshake:         handshake,
		supportedFeatures: supportedFeatures,
		connectionUris:    connectionUris,
	}

	ho.SetRequiresHandshake(false)

	return ho
}

func (m HandshakeDone) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeUint(uint64(types.ProtocolMethodHandshakeDone))
	if err != nil {
		return err
	}

	err = enc.EncodeBytes(m.handshake)
	if err != nil {
		return err
	}

	err = enc.EncodeInt(int64(m.supportedFeatures))
	if err != nil {
		return err
	}

	err = utils.EncodeMsgpackArray(enc, m.connectionUris)
	if err != nil {
		return err
	}

	return nil
}

func (m *HandshakeDone) SetChallenge(challenge []byte) {
	m.challenge = challenge
}

func (m *HandshakeDone) SetNetworkId(networkId string) {
	m.networkId = networkId
}

func NewHandshakeDone() *HandshakeDone {
	hn := &HandshakeDone{challenge: nil, networkId: "", supportedFeatures: -1}

	hn.SetRequiresHandshake(false)

	return hn
}

func (h HandshakeDone) HandleMessage(message IncomingMessageDataSigned) error {
	mediator := message.Mediator
	peer := message.Peer
	verifyId := message.VerifyId
	nodeId := message.NodeId
	logger := message.Logger

	if !mediator.ServicesStarted() {
		err := peer.End()
		if err != nil {
			return nil
		}
		return nil
	}

	if !bytes.Equal(peer.Challenge(), h.challenge) {
		return errors.New("Invalid challenge")
	}

	if !verifyId {
		peer.SetId(nodeId)
	} else {
		if !peer.Id().Equals(nodeId) {
			return fmt.Errorf("Invalid transports id on initial list")
		}
	}

	peer.SetConnected(true)
	peer.SetHandshakeDone(true)

	if h.supportedFeatures != types.SupportedFeatures {
		return fmt.Errorf("Remote node does not support required features")
	}
	err := mediator.AddPeer(peer)
	if err != nil {
		return err
	}

	if len(h.connectionUris) == 0 {
		return nil
	}

	peerId, err := peer.Id().ToString()

	if err != nil {
		return err
	}

	for _, uri := range h.connectionUris {
		uri.User = url.User(peerId)
	}

	peer.SetConnectionURIs(h.connectionUris)

	logger.Info(fmt.Sprintf("[+] %s (%s)", peerId, peer.RenderLocationURI()))

	err = mediator.ConnectToNode([]*url.URL{h.connectionUris[0]}, false, peer)
	if err != nil {
		return err
	}

	return nil
}

func (h *HandshakeDone) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageDataSigned) error {
	challenge, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	h.challenge = challenge

	supportedFeatures, err := dec.DecodeInt()

	if err != nil {
		return err
	}

	h.supportedFeatures = supportedFeatures

	connectionUris, err := utils.DecodeMsgpackURLArray(dec)

	if err != nil {
		return err
	}

	h.connectionUris = connectionUris
	return nil
}
