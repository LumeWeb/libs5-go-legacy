package protocol

import ()

var (
	messageTypes map[int]func() IncomingMessage
)

func RegisterProtocols() {
	messageTypes = make(map[int]func() IncomingMessage)

	// Register factory functions instead of instances
	RegisterMessageType(int(types.ProtocolMethodHandshakeOpen), func() IncomingMessage {
		return NewHandshakeOpen([]byte{}, "")
	})
	RegisterMessageType(int(types.ProtocolMethodHashQuery), func() IncomingMessage {
		return NewHashQuery()
	})
	RegisterMessageType(int(types.RecordTypeStorageLocation), func() IncomingMessage {
		return NewStorageLocation()
	})
	RegisterMessageType(int(types.RecordTypeRegistryEntry), func() IncomingMessage {
		return NewEmptyRegistryEntryRequest()
	})
	RegisterMessageType(int(types.ProtocolMethodRegistryQuery), func() IncomingMessage {
		return NewEmptyRegistryQuery()
	})
	RegisterMessageType(int(types.ProtocolMethodSignedMessage), func() IncomingMessage {
		return NewSignedMessage()
	})

}

func RegisterMessageType(messageType int, factoryFunc func() IncomingMessage) {
	if factoryFunc == nil {
		panic("factoryFunc cannot be nil")
	}
	messageTypes[messageType] = factoryFunc
}

func GetMessageType(kind int) (IncomingMessage, bool) {
	value, ok := messageTypes[kind]
	if !ok {
		return nil, false
	}

	return value(), true
}
