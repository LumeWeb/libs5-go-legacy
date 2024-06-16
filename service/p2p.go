package service

import (
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/net"
	"go.lumeweb.com/libs5-go/structs"
	"go.lumeweb.com/libs5-go/types"
	"net/url"
	"sync"
)

type P2PService interface {
	SelfConnectionUris() []*url.URL
	Peers() structs.Map
	ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer net.Peer) error
	OnNewPeer(peer net.Peer, verifyId bool) error
	GetNodeScore(nodeId *encoding.NodeId) (float64, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	SignMessageSimple(message []byte) ([]byte, error)
	AddPeer(peer net.Peer) error
	SendPublicPeersToPeer(peer net.Peer, peersToSend []net.Peer) error
	SendHashRequest(hash *encoding.Multihash, kinds []types.StorageLocationType) error
	UpVote(nodeId *encoding.NodeId) error
	DownVote(nodeId *encoding.NodeId) error
	NodeId() *encoding.NodeId
	WaitOnConnectedPeers()
	ConnectionTracker() *sync.WaitGroup
	NetworkId() string
	HashQueryRoutingTable() structs.Map
	Service
}
