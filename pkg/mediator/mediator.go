package mediator

import (
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/protocol/registry"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.lumeweb.com/libs5-go/pkg/storage/location"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"net/url"
)

type Mediator interface {
	NetworkId() string
	NodeId() *encoding.NodeId
	SelfConnectionUris() []*url.URL
	SignMessageSimple(message []byte) ([]byte, error)
	GetCachedStorageLocations(hash *encoding.Blob, kinds []location.StorageLocationType) (map[string]storage.StorageLocation, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	ProviderStore() storage.ProviderStore
	AddStorageLocation(hash *encoding.Blob, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error
	HashQueryRoutingTable() structs.Map
	Peers() structs.Map
	RegistrySet(sre registry.SignedRegistryEntry, trusted bool, receivedFrom transport.Peer) error
	RegistryGet(pk []byte) (registry.SignedRegistryEntry, error)
	ConnectToNode(connectionUris []*url.URL, retried bool, fromPeer transport.Peer) error
	ServicesStarted() bool
	AddPeer(peer transport.Peer) error
	SendPublicPeersToPeer(peer transport.Peer, peersToSend []transport.Peer) error
}
