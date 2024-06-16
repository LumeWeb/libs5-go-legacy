package _default

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
	"go.lumeweb.com/libs5-go/db"
	"go.lumeweb.com/libs5-go/ed25519"
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/net"
	"go.lumeweb.com/libs5-go/protocol"
	"go.lumeweb.com/libs5-go/service"
	"go.lumeweb.com/libs5-go/structs"
	"go.lumeweb.com/libs5-go/types"
	"go.uber.org/zap"
	"net/url"
	"sort"
	"sync"
	"time"
)

var _ service.P2PService = (*P2PServiceDefault)(nil)

var (
	errUnsupportedProtocol       = errors.New("unsupported protocol")
	errConnectionIdMissingNodeID = errors.New("connection id missing node id")
)

const nodeBucketName = "nodes"

type P2PServiceDefault struct {
	nodeKeyPair             *ed25519.KeyPairEd25519
	localNodeID             *encoding.NodeId
	networkID               string
	nodesBucket             *bolt.Bucket
	inited                  bool
	reconnectDelay          structs.Map
	peers                   structs.Map
	peersPending            structs.Map
	selfConnectionUris      []*url.URL
	outgoingPeerBlocklist   structs.Map
	incomingPeerBlockList   structs.Map
	incomingIPBlocklist     structs.Map
	outgoingPeerFailures    structs.Map
	maxOutgoingPeerFailures uint
	connections             sync.WaitGroup
	hashQueryRoutingTable   structs.Map
	bucket                  db.KVStore
	service.ServiceBase
}

func NewP2P(params service.ServiceParams) *P2PServiceDefault {
	uri, err := url.Parse(fmt.Sprintf("wss://%s:%d/s5/p2p", params.Config.HTTP.API.Domain, params.Config.HTTP.API.Port))
	if err != nil {
		params.Logger.Fatal("failed to parse HTTP API URL", zap.Error(err))
	}

	service := &P2PServiceDefault{
		nodeKeyPair:             params.Config.KeyPair,
		networkID:               params.Config.P2P.Network,
		inited:                  false,
		reconnectDelay:          structs.NewMap(),
		peers:                   structs.NewMap(),
		peersPending:            structs.NewMap(),
		selfConnectionUris:      []*url.URL{uri},
		outgoingPeerBlocklist:   structs.NewMap(),
		incomingPeerBlockList:   structs.NewMap(),
		incomingIPBlocklist:     structs.NewMap(),
		outgoingPeerFailures:    structs.NewMap(),
		maxOutgoingPeerFailures: params.Config.P2P.MaxOutgoingPeerFailures,
		hashQueryRoutingTable:   structs.NewMap(),
		ServiceBase:             service.NewServiceBase(params.Logger, params.Config, params.Db),
	}

	return service
}

func (p *P2PServiceDefault) SelfConnectionUris() []*url.URL {
	return p.selfConnectionUris
}

func (p *P2PServiceDefault) Peers() structs.Map {
	return p.peers
}

func (p *P2PServiceDefault) Start(ctx context.Context) error {
	config := p.Config()
	if len(config.P2P.Peers.Initial) > 0 {
		initialPeers := config.P2P.Peers.Initial

		for _, peer := range initialPeers {
			u, err := url.Parse(peer)
			if err != nil {
				return err
			}

			peer := peer
			go func() {
				err := p.ConnectToNode([]*url.URL{u}, 0, nil)
				if err != nil {
					p.Logger().Error("failed to connect to initial peer", zap.Error(err), zap.String("peer", peer))
				}
			}()
		}
	}

	return nil
}

func (p *P2PServiceDefault) Stop(ctx context.Context) error {
	return nil
}

func (p *P2PServiceDefault) Init(ctx context.Context) error {
	if p.inited {
		return nil
	}

	p.localNodeID = encoding.NewNodeId(p.nodeKeyPair.PublicKey())

	bucket, err := p.Db().Bucket(nodeBucketName)
	if err != nil {
		return err
	}

	err = bucket.Open()
	if err != nil {
		return err
	}

	p.bucket = bucket
	p.inited = true

	for _, peer := range p.Config().P2P.Peers.Blocklist {
		_, err := encoding.DecodeNodeId(peer)
		if err != nil {
			return err
		}

		p.incomingPeerBlockList.Put(peer, true)
		p.outgoingPeerBlocklist.Put(peer, true)
	}

	return nil
}
func (p *P2PServiceDefault) ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer net.Peer) error {
	if !p.Services().IsStarted() {
		if !p.Services().IsStarting() {
			return nil
		}
	}

	unsupported, _ := url.Parse("http://0.0.0.0")
	unsupported.Scheme = "unsupported"

	var connectionUri *url.URL

	for _, uri := range connectionUris {
		if uri.Scheme == "ws" || uri.Scheme == "wss" {
			connectionUri = uri
			break
		}
	}

	if connectionUri == nil {
		for _, uri := range connectionUris {
			if uri.Scheme == "tcp" {
				connectionUri = uri
				break
			}
		}
	}

	if connectionUri == nil {
		connectionUri = unsupported
	}

	if connectionUri.Scheme == "unsupported" {
		return errUnsupportedProtocol
	}

	scheme := connectionUri.Scheme

	if connectionUri.User == nil {
		return errConnectionIdMissingNodeID
	}

	username := connectionUri.User.Username()
	id, err := encoding.DecodeNodeId(username)
	if err != nil {
		return err
	}

	idString, err := id.ToString()
	if err != nil {
		return err
	}

	if p.peersPending.Contains(idString) || p.peers.Contains(idString) {
		p.Logger().Debug("already connected", zap.String("node", connectionUri.String()))
		return nil
	}

	if p.outgoingPeerBlocklist.Contains(idString) {
		p.Logger().Debug("outgoing peer is on blocklist", zap.String("node", connectionUri.String()))

		var fromPeerId string

		if fromPeer != nil {
			blocked := false
			if fromPeer.Id() != nil {
				fromPeerId, err = fromPeer.Id().ToString()
				if err != nil {
					return err
				}
				if !p.incomingPeerBlockList.Contains(fromPeerId) {
					p.incomingPeerBlockList.Put(fromPeerId, true)
					blocked = true
				}
			}

			fromPeerIP := fromPeer.GetIPString()

			if !p.incomingIPBlocklist.Contains(fromPeerIP) {
				p.incomingIPBlocklist.Put(fromPeerIP, true)
				blocked = true
			}
			err = fromPeer.EndForAbuse()
			if err != nil {
				return err
			}

			if blocked {
				p.Logger().Debug("blocking peer for sending peer on blocklist", zap.String("node", connectionUri.String()), zap.String("peer", fromPeerId), zap.String("ip", fromPeerIP))
			}
		}
		return nil
	}

	reconnectDelay := p.reconnectDelay.GetUInt(idString)
	if reconnectDelay == nil {
		delay := uint(1)
		reconnectDelay = &delay
	}

	if id.Equals(p.localNodeID) {
		return nil
	}

	p.Logger().Debug("connect", zap.String("node", connectionUri.String()))

	socket, err := net.CreateTransportSocket(scheme, connectionUri)
	if err != nil {
		if retry > p.Config().P2P.MaxConnectionAttempts {
			p.Logger().Error("failed to connect, too many retries", zap.String("node", connectionUri.String()), zap.Error(err))
			counter := uint(0)
			if p.outgoingPeerFailures.Contains(idString) {
				tmp := *p.outgoingPeerFailures.GetUInt(idString)
				counter = tmp
			}

			counter++

			p.outgoingPeerFailures.PutUInt(idString, counter)

			if counter >= p.maxOutgoingPeerFailures {

				if fromPeer != nil {
					blocked := false
					var fromPeerId string
					if fromPeer.Id() != nil {
						fromPeerId, err = fromPeer.Id().ToString()
						if err != nil {
							return err
						}
						if !p.incomingPeerBlockList.Contains(fromPeerId) {
							p.incomingPeerBlockList.Put(fromPeerId, true)
							blocked = true
						}
					}

					fromPeerIP := fromPeer.GetIPString()
					if !p.incomingIPBlocklist.Contains(fromPeerIP) {
						p.incomingIPBlocklist.Put(fromPeerIP, true)
						blocked = true
					}
					err = fromPeer.EndForAbuse()
					if err != nil {
						return err
					}

					if blocked {
						p.Logger().Debug("blocking peer for sending peer on blocklist", zap.String("node", connectionUri.String()), zap.String("peer", fromPeerId), zap.String("ip", fromPeerIP))
					}
				}
				p.outgoingPeerBlocklist.Put(idString, true)
				p.Logger().Debug("blocking peer for too many failures", zap.String("node", connectionUri.String()))
			}

			return nil
		}

		if errors.Is(err, net.ErrTransportNotSupported) {
			p.Logger().Debug("failed to connect, unsupported transport", zap.String("node", connectionUri.String()), zap.Error(err))
			return err
		}

		retry++

		p.Logger().Error("failed to connect", zap.String("node", connectionUri.String()), zap.Error(err))

		delay := p.reconnectDelay.GetUInt(idString)
		if delay == nil {
			tmp := uint(1)
			delay = &tmp
		}
		delayDeref := *delay
		p.reconnectDelay.PutUInt(idString, delayDeref*2)

		time.Sleep(time.Duration(delayDeref) * time.Second)

		return p.ConnectToNode(connectionUris, retry, fromPeer)
	}

	if p.outgoingPeerFailures.Contains(idString) {
		p.outgoingPeerFailures.Remove(idString)
	}

	peer, err := net.CreateTransportPeer(scheme, &net.TransportPeerConfig{
		Socket: socket,
		Uris:   []*url.URL{connectionUri},
	})

	if err != nil {
		return err
	}

	peer.SetId(id)

	p.Services().P2P().ConnectionTracker().Add(1)

	peerId, err := peer.Id().ToString()
	if err != nil {
		return err
	}
	p.peersPending.Put(peerId, peer)

	go func() {
		err := p.OnNewPeer(peer, true)
		if err != nil && !peer.Abuser() {
			p.Logger().Error("peer error", zap.Error(err))
		}
		p.Services().P2P().ConnectionTracker().Done()
	}()

	return nil

}

func (p *P2PServiceDefault) OnNewPeer(peer net.Peer, verifyId bool) error {
	var wg sync.WaitGroup

	var pid string

	if peer.Id() != nil {
		pid, _ = peer.Id().ToString()
	} else {
		pid = "unknown"
	}

	pip := peer.GetIPString()

	if p.incomingIPBlocklist.Contains(pid) {
		p.Logger().Error("peer is on identity blocklist", zap.String("peer", pid))
		err := peer.EndForAbuse()
		if err != nil {
			return err
		}
		return nil
	}
	if p.incomingPeerBlockList.Contains(pip) {
		p.Logger().Debug("peer is on ip blocklist", zap.String("peer", pid), zap.String("ip", pip))
		err := peer.EndForAbuse()
		if err != nil {
			return err
		}
		return nil
	}

	p.Logger().Debug("OnNewPeer started", zap.String("peer", pid))

	challenge := protocol.GenerateChallenge()
	peer.SetChallenge(challenge)

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.OnNewPeerListen(peer, verifyId)
	}()

	handshakeOpenMsg, err := msgpack.Marshal(protocol.NewHandshakeOpen(challenge, p.networkID))
	if err != nil {
		return err
	}

	err = peer.SendMessage(handshakeOpenMsg)
	if err != nil {
		return err
	}
	p.Logger().Debug("OnNewPeer sent handshake", zap.String("peer", pid))

	p.Logger().Debug("OnNewPeer before Wait", zap.String("peer", pid))
	wg.Wait() // Wait for OnNewPeerListen goroutine to finish
	p.Logger().Debug("OnNewPeer ended", zap.String("peer", pid))
	return nil
}
func (p *P2PServiceDefault) OnNewPeerListen(peer net.Peer, verifyId bool) {
	onDone := net.CloseCallback(func() {
		if peer.Id() != nil {
			pid, err := peer.Id().ToString()
			if err != nil {
				p.Logger().Error("failed to get peer id", zap.Error(err))
				return
			}
			// Handle closure of the connection
			if p.peers.Contains(pid) {
				p.peers.Remove(pid)
			}
			if p.peersPending.Contains(pid) {
				p.peersPending.Remove(pid)
			}
		}
	})

	onError := net.ErrorCallback(func(args ...interface{}) {
		if !peer.Abuser() {
			p.Logger().Error("peer error", zap.Any("args", args))
		}
	})

	peer.ListenForMessages(func(message []byte) error {
		var reader protocol.IncomingMessageReader

		err := msgpack.Unmarshal(message, &reader)
		if err != nil {
			p.Logger().Error("Error decoding basic message info", zap.Error(err))
			return err
		}

		// Now, get the specific message handler based on the message kind
		handler, ok := protocol.GetMessageType(reader.Kind)
		if !ok {
			p.Logger().Error("Unknown message type", zap.Int("type", reader.Kind))
			return fmt.Errorf("unknown message type: %d", reader.Kind)
		}

		if handler.RequiresHandshake() && !peer.IsHandshakeDone() {
			p.Logger().Debug("Peer is not handshake done, ignoring message", zap.Any("type", types.ProtocolMethodMap[types.ProtocolMethod(reader.Kind)]))
			return nil
		}

		data := protocol.IncomingMessageData{
			Original: message,
			Data:     reader.Data,
			Ctx:      context.Background(),
			Peer:     peer,
			VerifyId: verifyId,
			Config:   p.Config(),
			Logger:   p.Logger(),
			Mediator: NewMediator(service.ServiceParams{
				Logger: p.Logger(),
				Config: p.Config(),
				Db:     p.Db(),
			}),
		}

		if mediator, ok := data.Mediator.(service.ServicesSetter); ok {
			mediator.SetServices(p.Services())
		} else {
			p.Logger().Fatal("failed to cast mediator to service.Service")
		}

		dec := msgpack.NewDecoder(bytes.NewReader(reader.Data))

		err = handler.DecodeMessage(dec, data)
		if err != nil {
			p.Logger().Error("Error decoding message", zap.Error(err))
			return err
		}

		// Directly decode and handle the specific message type
		if err = handler.HandleMessage(data); err != nil {
			p.Logger().Error("Error handling message", zap.Error(err))
			return err
		}

		return nil
	}, net.ListenerOptions{
		OnClose: &onDone,
		OnError: &onError,
		Logger:  p.Logger(),
	})
}

func (p *P2PServiceDefault) readNodeVotes(nodeId *encoding.NodeId) (service.NodeVotes, error) {
	var value []byte

	value, err := p.bucket.Get(nodeId.Raw())
	if err != nil {
		return nil, err
	}

	if value == nil {
		return service.NewNodeVotes(), nil
	}

	score := service.NewNodeVotes()
	err = msgpack.Unmarshal(value, &score)
	if err != nil {
		return nil, err
	}

	return score, nil
}

func (p *P2PServiceDefault) saveNodeVotes(nodeId *encoding.NodeId, votes service.NodeVotes) error {
	// Marshal the votes into data
	data, err := msgpack.Marshal(votes)
	if err != nil {
		return err
	}

	err = p.bucket.Put(nodeId.Raw(), data)
	if err != nil {
		return err
	}

	return nil
}

func (p *P2PServiceDefault) GetNodeScore(nodeId *encoding.NodeId) (float64, error) {
	if nodeId.Equals(p.localNodeID) {
		return 1, nil
	}

	score, err := p.readNodeVotes(nodeId)
	if err != nil {
		return 0.5, err
	}

	return protocol.CalculateNodeScore(score.Good(), score.Bad()), nil

}
func (p *P2PServiceDefault) SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error) {
	scores := make(map[encoding.NodeIdCode]float64)
	var errOccurred error

	for _, nodeId := range nodes {
		score, err := p.GetNodeScore(nodeId)
		if err != nil {
			errOccurred = err
			scores[nodeId.HashCode()] = 0 // You may choose a different default value for error cases
		} else {
			scores[nodeId.HashCode()] = score
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return scores[nodes[i].HashCode()] > scores[nodes[j].HashCode()]
	})

	return nodes, errOccurred
}
func (p *P2PServiceDefault) SignMessageSimple(message []byte) ([]byte, error) {
	signedMessage := protocol.NewSignedMessageRequest(message)
	signedMessage.SetNodeId(p.localNodeID)

	err := signedMessage.Sign(p.Config())

	if err != nil {
		return nil, err
	}

	result, err := msgpack.Marshal(signedMessage)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *P2PServiceDefault) AddPeer(peer net.Peer) error {
	peerId, err := peer.Id().ToString()
	if err != nil {
		return err
	}
	p.peers.Put(peerId, peer)
	p.reconnectDelay.PutUInt(peerId, 1)

	if p.peersPending.Contains(peerId) {
		p.peersPending.Remove(peerId)
	}

	return nil
}
func (p *P2PServiceDefault) SendPublicPeersToPeer(peer net.Peer, peersToSend []net.Peer) error {
	announceRequest := protocol.NewAnnounceRequest(peer, peersToSend)

	message, err := msgpack.Marshal(announceRequest)

	if err != nil {
		return err
	}

	signedMessage, err := p.SignMessageSimple(message)

	if err != nil {
		return err
	}

	err = peer.SendMessage(signedMessage)

	return nil
}
func (p *P2PServiceDefault) SendHashRequest(hash *encoding.Multihash, kinds []types.StorageLocationType) error {
	hashRequest := protocol.NewHashRequest(hash, kinds)
	message, err := msgpack.Marshal(hashRequest)
	if err != nil {
		return err
	}

	for _, peer := range p.peers.Values() {
		peerValue, ok := peer.(net.Peer)
		if !ok {
			p.Logger().Error("failed to cast peer to net.Peer")
			continue
		}
		err = peerValue.SendMessage(message)
	}

	return nil
}

func (p *P2PServiceDefault) UpVote(nodeId *encoding.NodeId) error {
	err := p.vote(nodeId, true)
	if err != nil {
		return err
	}

	return nil
}

func (p *P2PServiceDefault) DownVote(nodeId *encoding.NodeId) error {
	err := p.vote(nodeId, false)
	if err != nil {
		return err
	}

	return nil
}

func (p *P2PServiceDefault) vote(nodeId *encoding.NodeId, upvote bool) error {
	votes, err := p.readNodeVotes(nodeId)
	if err != nil {
		return err
	}

	if upvote {
		votes.Upvote()
	} else {
		votes.Downvote()
	}

	err = p.saveNodeVotes(nodeId, votes)
	if err != nil {
		return err
	}

	return nil
}
func (p *P2PServiceDefault) NodeId() *encoding.NodeId {
	return p.localNodeID
}

func (p *P2PServiceDefault) WaitOnConnectedPeers() {
	p.connections.Wait()
}

func (p *P2PServiceDefault) ConnectionTracker() *sync.WaitGroup {
	return &p.connections
}

func (p *P2PServiceDefault) NetworkId() string {
	return p.Config().P2P.Network
}
func (n *P2PServiceDefault) HashQueryRoutingTable() structs.Map {
	return n.hashQueryRoutingTable
}
