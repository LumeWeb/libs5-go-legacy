package protocol

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"io"
)

var (
	_ IncomingMessage       = (*SignedMessage)(nil)
	_ msgpack.CustomDecoder = (*signedMessageReader)(nil)
	_ msgpack.CustomEncoder = (*SignedMessage)(nil)
)

var (
	errInvalidSignature = errors.New("Invalid signature found")
)

type SignedMessage struct {
	nodeId    *encoding.NodeId
	signature []byte
	message   []byte
	HandshakeRequirement
}

func (s *SignedMessage) NodeId() *encoding.NodeId {
	return s.nodeId
}

func (s *SignedMessage) SetNodeId(nodeId *encoding.NodeId) {
	s.nodeId = nodeId
}

func (s *SignedMessage) SetSignature(signature []byte) {
	s.signature = signature
}

func (s *SignedMessage) SetMessage(message []byte) {
	s.message = message
}

func NewSignedMessageRequest(message []byte) *SignedMessage {
	return &SignedMessage{message: message}
}

type signedMessageReader struct {
	kind    int
	message msgpack.RawMessage
}

func (s *signedMessageReader) DecodeMsgpack(dec *msgpack.Decoder) error {
	kind, err := dec.DecodeInt()
	if err != nil {
		return err
	}

	s.kind = kind

	message, err := io.ReadAll(dec.Buffered())

	if err != nil {
		return err
	}

	s.message = message

	return nil
}

func NewSignedMessage() *SignedMessage {
	sm := &SignedMessage{}

	sm.SetRequiresHandshake(false)

	return sm
}

func (s *SignedMessage) HandleMessage(message IncomingMessageData) error {
	var payload signedMessageReader
	peer := message.Peer
	logger := message.Logger

	err := msgpack.Unmarshal(s.message, &payload)
	if err != nil {
		return err
	}

	if msgHandler, valid := GetSignedMessageType(payload.kind); valid {
		logger.Debug("SignedMessage", zap.Any("type", types.ProtocolMethodMap[types.ProtocolMethod(payload.kind)]))
		if msgHandler.RequiresHandshake() && !peer.IsHandshakeDone() {
			nid, err := s.nodeId.ToString()
			if err != nil {
				return err
			}
			logger.Debug("Peer is not handshake done, ignoring message", zap.Any("type", types.ProtocolMethodMap[types.ProtocolMethod(payload.kind)]), zap.String("peerIP", peer.GetIPString()), zap.String("nodeId", nid))
			return nil
		}

		signedDec := msgpack.NewDecoder(bytes.NewReader(payload.message))
		signedMsg := IncomingMessageDataSigned{
			IncomingMessageData: message,
			NodeId:              s.nodeId,
		}

		err = msgHandler.DecodeMessage(signedDec, signedMsg)

		if err != nil {
			logger.Error("Error decoding signed message", zap.Error(err))
			return err
		}

		if err = msgHandler.HandleMessage(signedMsg); err != nil {
			logger.Error("Error handling signed message", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *SignedMessage) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error {
	nodeId, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	s.nodeId = encoding.NewNodeId(nodeId)

	signature, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	s.signature = signature

	signedMessage, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	s.message = signedMessage

	if !ed25519.Verify(s.nodeId.Raw()[1:], s.message, s.signature) {
		return errInvalidSignature
	}

	return nil
}
func (s *SignedMessage) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeInt(int64(types.ProtocolMethodSignedMessage))

	if err != nil {
		return err
	}

	err = enc.EncodeBytes(s.nodeId.Raw())

	if err != nil {
		return err
	}

	err = enc.EncodeBytes(s.signature)

	if err != nil {
		return err
	}

	err = enc.EncodeBytes(s.message)

	if err != nil {
		return err
	}

	return nil
}
func (s *SignedMessage) Sign(cfg *config.NodeConfig) error {
	if s.nodeId == nil {
		panic("nodeId is nil")
	}

	if s.message == nil {
		panic("message is nil")
	}

	s.signature = ed25519.Sign(cfg.KeyPair.ExtractBytes(), s.message)

	return nil
}
