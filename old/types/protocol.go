package types

type ProtocolMethod int

const (
	ProtocolMethodHandshakeOpen ProtocolMethod = 0x1
	ProtocolMethodHandshakeDone ProtocolMethod = 0x2
	ProtocolMethodSignedMessage ProtocolMethod = 0xA
	ProtocolMethodHashQuery     ProtocolMethod = 0x4
	ProtocolMethodAnnouncePeers ProtocolMethod = 0x8
	ProtocolMethodRegistryQuery ProtocolMethod = 0xD
	RecordTypeStorageLocation   ProtocolMethod = 0x05
	RecordTypeStreamEvent       ProtocolMethod = 0x09
	RecordTypeRegistryEntry     ProtocolMethod = 0x07
)

var ProtocolMethodMap = map[ProtocolMethod]string{
	ProtocolMethodHandshakeOpen: "HandshakeOpen",
	ProtocolMethodHandshakeDone: "IsHandshakeDone",
	ProtocolMethodSignedMessage: "SignedMessage",
	ProtocolMethodHashQuery:     "HashQuery",
	ProtocolMethodAnnouncePeers: "AnnouncePeers",
	ProtocolMethodRegistryQuery: "RegistryQuery",
	RecordTypeStorageLocation:   "StorageLocation",
	RecordTypeStreamEvent:       "StreamEvent",
	RecordTypeRegistryEntry:     "RegistryEntry",
}
