package protocol

import (
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/net"
	"go.lumeweb.com/libs5-go/storage"
	"go.lumeweb.com/libs5-go/structs"
	"go.lumeweb.com/libs5-go/types"
	"net/url"
)

type Mediator interface {
	NetworkId() string
	NodeId() *encoding.NodeId
	SelfConnectionUris() []*url.URL
	SignMessageSimple(message []byte) ([]byte, error)
	GetCachedStorageLocations(hash *encoding.Multihash, kinds []types.StorageLocationType) (map[string]storage.StorageLocation, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	ProviderStore() storage.ProviderStore
	AddStorageLocation(hash *encoding.Multihash, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error
	HashQueryRoutingTable() structs.Map
	Peers() structs.Map
	RegistrySet(sre SignedRegistryEntry, trusted bool, receivedFrom net.Peer) error
	RegistryGet(pk []byte) (SignedRegistryEntry, error)
	ConnectToNode(connectionUris []*url.URL, retried bool, fromPeer net.Peer) error
	ServicesStarted() bool
	AddPeer(peer net.Peer) error
	SendPublicPeersToPeer(peer net.Peer, peersToSend []net.Peer) error
}
