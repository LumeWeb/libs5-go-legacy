package protocol

import (
	"github.com/vmihailenco/msgpack/v5"
)

type IncomingMessageDataSigned struct {
	IncomingMessageData
	NodeId *encoding.NodeId
}

type IncomingMessageSigned interface {
	HandleMessage(message IncomingMessageDataSigned) error
	DecodeMessage(dec *msgpack.Decoder, message IncomingMessageDataSigned) error
	HandshakeRequirer
}

var (
	signedMessageTypes map[int]func() IncomingMessageSigned
)

func RegisterSignedProtocols() {
	signedMessageTypes = make(map[int]func() IncomingMessageSigned)

	RegisterSignedMessageType(int(types.ProtocolMethodHandshakeDone), func() IncomingMessageSigned {
		return NewHandshakeDone()
	})
	RegisterSignedMessageType(int(types.ProtocolMethodAnnouncePeers), func() IncomingMessageSigned {
		return NewAnnouncePeers()
	})
}

func RegisterSignedMessageType(messageType int, factoryFunc func() IncomingMessageSigned) {
	if factoryFunc == nil {
		panic("factoryFunc cannot be nil")
	}
	signedMessageTypes[messageType] = factoryFunc
}

func GetSignedMessageType(kind int) (IncomingMessageSigned, bool) {
	value, ok := signedMessageTypes[kind]
	if !ok {
		return nil, false
	}

	return value(), true
}
