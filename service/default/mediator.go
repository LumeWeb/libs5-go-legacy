package _default

import (
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/net"
	"github.com/LumeWeb/libs5-go/protocol"
	"github.com/LumeWeb/libs5-go/service"
	"github.com/LumeWeb/libs5-go/storage"
	"github.com/LumeWeb/libs5-go/structs"
	"github.com/LumeWeb/libs5-go/types"
	"net/url"
)

var _ protocol.Mediator = (*MediatorDefault)(nil)

type MediatorDefault struct {
	service.ServiceBase
}

func (m MediatorDefault) NetworkId() string {
	return m.Services().P2P().NetworkId()
}

func (m MediatorDefault) NodeId() *encoding.NodeId {
	return m.Services().P2P().NodeId()
}

func (m MediatorDefault) SelfConnectionUris() []*url.URL {
	return m.Services().P2P().SelfConnectionUris()
}

func (m MediatorDefault) SignMessageSimple(message []byte) ([]byte, error) {
	return m.Services().P2P().SignMessageSimple(message)
}

func (m MediatorDefault) GetCachedStorageLocations(hash *encoding.Multihash, kinds []types.StorageLocationType) (map[string]storage.StorageLocation, error) {
	return m.Services().Storage().GetCachedStorageLocations(hash, kinds, false)
}

func (m MediatorDefault) SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error) {
	return m.Services().P2P().SortNodesByScore(nodes)
}

func (m MediatorDefault) ProviderStore() storage.ProviderStore {
	return m.Services().Storage().ProviderStore()
}

func (m MediatorDefault) AddStorageLocation(hash *encoding.Multihash, nodeId *encoding.NodeId, location storage.StorageLocation, message []byte) error {
	return m.Services().Storage().AddStorageLocation(hash, nodeId, location, message)
}

func (m MediatorDefault) HashQueryRoutingTable() structs.Map {
	return m.Services().P2P().HashQueryRoutingTable()
}

func (m MediatorDefault) Peers() structs.Map {
	return m.Services().P2P().Peers()
}

func (m MediatorDefault) RegistrySet(sre protocol.SignedRegistryEntry, trusted bool, receivedFrom net.Peer) error {
	return m.Services().Registry().Set(sre, trusted, receivedFrom)
}

func (m MediatorDefault) RegistryGet(pk []byte) (protocol.SignedRegistryEntry, error) {
	return m.Services().Registry().Get(pk)
}

func (m MediatorDefault) ConnectToNode(connectionUris []*url.URL, retried bool, fromPeer net.Peer) error {
	return m.Services().P2P().ConnectToNode(connectionUris, 0, fromPeer)
}

func (m MediatorDefault) ServicesStarted() bool {
	return m.Services().IsStarted()
}

func (m MediatorDefault) AddPeer(peer net.Peer) error {
	return m.Services().P2P().AddPeer(peer)
}

func (m MediatorDefault) SendPublicPeersToPeer(peer net.Peer, peersToSend []net.Peer) error {
	return m.Services().P2P().SendPublicPeersToPeer(peer, peersToSend)
}

func NewMediator(params service.ServiceParams) *MediatorDefault {
	return &MediatorDefault{
		ServiceBase: service.NewServiceBase(params.Logger, params.Config, params.Db),
	}
}
