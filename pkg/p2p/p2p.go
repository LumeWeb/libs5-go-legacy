package p2p

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"net/url"
	"old/net"
	"old/types"
	"sort"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/protocol"
	"go.lumeweb.com/libs5-go/pkg/service"
	"go.lumeweb.com/libs5-go/pkg/structs" // Assuming you still want this
	"go.uber.org/zap"
)

var (
	errUnsupportedProtocol       = errors.New("unsupported protocol")
	errConnectionIdMissingNodeID = errors.New("connection id missing node id")
)

var _ P2PService = (*P2PServiceDefault)(nil)

type P2PService interface {
	service.BaseService
	SelfConnectionUris() []*url.URL
	Peers() structs.Map
	ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer transport.Peer) error
	OnNewPeer(peer transport.Peer, verifyId bool) error
	GetNodeScore(nodeId *encoding.NodeId) (float64, error)
	SortNodesByScore(nodes []*encoding.NodeId) ([]*encoding.NodeId, error)
	SignMessageSimple(message []byte) ([]byte, error)
	AddPeer(peer transport.Peer) error
	SendPublicPeersToPeer(peer transport.Peer, peersToSend []transport.Peer) error
	SendHashRequest(hash *encoding.Multihash, kinds []storage.StorageLocationType) error
	UpVote(nodeId *encoding.NodeId) error
	DownVote(nodeId *encoding.NodeId) error
	NodeId() *encoding.NodeId
	WaitOnConnectedPeers()
	ConnectionTracker() *sync.WaitGroup
	NetworkId() string
	HashQueryRoutingTable() structs.Map
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	Logger() *zap.Logger
	Config() *config.NodeConfig
	DB() kv.KVStore
}

type P2PServiceDefault struct {
	nodeConfig              *config.NodeConfig
	crypto                  crypto.CryptoImplementation
	db                      kv.KVStore
	localNodeID             *encoding.NodeId
	networkID               string
	selfConnectionUris      []*url.URL
	peers                   structs.Map
	connectionTracker       sync.WaitGroup
	hashQueryRoutingTable   structs.Map
	reconnectDelay          structs.Map
	outgoingPeerBlocklist   structs.Map
	incomingPeerBlockList   structs.Map
	incomingIPBlocklist     structs.Map
	outgoingPeerFailures    structs.Map
	maxOutgoingPeerFailures uint

	logger *zap.Logger
	// Add any other P2P-specific fields here
}

func NewP2PService(cfg *config.NodeConfig, crypto crypto.CryptoImplementation, db db.KVStore, logger *zap.Logger, selfConnectionUris []*url.URL) (*P2PServiceDefault, error) {
	localNodeID := encoding.NewNodeId(cfg.KeyPair.PublicKey())
	if selfConnectionUris == nil || len(selfConnectionUris) == 0 {
		uri, err := url.Parse(fmt.Sprintf("wss://%s:%d/s5/p2p", cfg.HTTP.API.Domain, cfg.HTTP.API.Port))
		if err != nil {
			return nil, err
		}
		selfConnectionUris = []*url.URL{uri}
	}

	return &P2PServiceDefault{
		nodeConfig:            cfg,
		crypto:                crypto,
		db:                    db,
		localNodeID:           localNodeID,
		networkID:             cfg.P2P.Network,
		selfConnectionUris:    selfConnectionUris,
		peers:                 structs.NewMap(),
		connectionTracker:     sync.WaitGroup{},
		reconnectDelay:        structs.NewMap(),
		outgoingPeerBlocklist: structs.NewMap(),
		incomingPeerBlockList: structs.NewMap(),
		incomingIPBlocklist:   structs.NewMap(),
		outgoingPeerFailures:  structs.NewMap(),
		hashQueryRoutingTable: structs.NewMap(),

		logger:                  logger,
		maxOutgoingPeerFailures: cfg.P2P.MaxOutgoingPeerFailures,
	}, nil
}

func (p *P2PServiceDefault) Start(ctx context.Context) error {
	config := p.nodeConfig
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
					p.logger.Error("failed to connect to initial peer", zap.Error(err), zap.String("peer", peer))
				}
			}()
		}
	}

	return nil
}

func (p *P2PServiceDefault) ConnectToNode(connectionUris []*url.URL, retry uint, fromPeer net.Peer) error {
	if len(connectionUris) == 0 {
		return errors.New("No connection URIs provided")
	}
	for _, connectionUri := range connectionUris {

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

		if p.outgoingPeerBlocklist.Contains(idString) {
			p.logger.Debug("outgoing peer is on blocklist", zap.String("node", connectionUri.String()))

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
					p.logger.Debug("blocking peer for sending peer on blocklist", zap.String("node", connectionUri.String()), zap.String("peer", fromPeerId), zap.String("ip", fromPeerIP))
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
			p.logger.Debug("skipping connection to self", zap.String("node", connectionUri.String()))
			continue
		}

		p.logger.Debug("connect", zap.String("node", connectionUri.String()))

		socket, err := net.CreateTransportSocket(scheme, connectionUri)
		if err != nil {
			if retry > p.nodeConfig.P2P.MaxConnectionAttempts {
				p.logger.Error("failed to connect, too many retries", zap.String("node", connectionUri.String()), zap.Error(err))
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
							p.logger.Debug("blocking peer for sending peer on blocklist", zap.String("node", connectionUri.String()), zap.String("peer", fromPeerId), zap.String("ip", fromPeerIP))
						}
					}
					p.outgoingPeerBlocklist.Put(idString, true)
					p.logger.Debug("blocking peer for too many failures", zap.String("node", connectionUri.String()))
				}

				return nil
			}

			if errors.Is(err, net.ErrTransportNotSupported) {
				p.logger.Debug("failed to connect, unsupported transport", zap.String("node", connectionUri.String()), zap.Error(err))
				return err
			}

			retry++

			p.logger.Error("failed to connect", zap.String("node", connectionUri.String()), zap.Error(err))

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

		p.connectionTracker.Add(1)

		go func() {
			err := p.OnNewPeer(peer, true)
			if err != nil && !peer.Abuser() {
				p.logger.Error("peer error", zap.Error(err))
			}
			p.connectionTracker.Done()
		}()

		return nil
	}

	return errUnsupportedProtocol
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

	if p.incomingIPBlocklist.Contains(pip) {
		p.logger.Error("peer is on ip blocklist", zap.String("peer", pid), zap.String("ip", pip))
		err := peer.EndForAbuse()
		if err != nil {
			return err
		}
		return nil
	}
	if p.incomingPeerBlockList.Contains(pid) {
		p.logger.Debug("peer is on identity blocklist", zap.String("peer", pid))
		err := peer.EndForAbuse()
		if err != nil {
			return err
		}
		return nil
	}

	p.logger.Debug("OnNewPeer started", zap.String("peer", pid))

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
	p.logger.Debug("OnNewPeer sent handshake", zap.String("peer", pid))

	p.logger.Debug("OnNewPeer before Wait", zap.String("peer", pid))
	wg.Wait() // Wait for OnNewPeerListen goroutine to finish
	p.logger.Debug("OnNewPeer ended", zap.String("peer", pid))
	return nil
}

func (p *P2PServiceDefault) OnNewPeerListen(peer net.Peer, verifyId bool) {
	onDone := func() {
		if peer.Id() != nil {
			pid, err := peer.Id().ToString()
			if err != nil {
				p.logger.Error("failed to get peer id", zap.Error(err))
				return
			}
			// Handle closure of the connection
			if p.peers.Contains(pid) {
				p.peers.Remove(pid)
			}

		}
	}

	onError := func(args ...interface{}) {
		if !peer.Abuser() {
			p.logger.Error("peer error", zap.Any("args", args))
		}
	}

	peer.ListenForMessages(func(message []byte) error {
		var reader protocol.IncomingMessageReader

		err := msgpack.Unmarshal(message, &reader)
		if err != nil {
			p.logger.Error("Error decoding basic message info", zap.Error(err))
			return err
		}

		// Now, get the specific message handler based on the message kind
		handler, ok := protocol.GetMessageType(reader.Kind)
		if !ok {
			p.logger.Error("Unknown message type", zap.Int("type", reader.Kind))
			return fmt.Errorf("unknown message type: %d", reader.Kind)
		}

		if handler.RequiresHandshake() && !peer.IsHandshakeDone() {
			p.logger.Debug("Peer is not handshake done, ignoring message", zap.Any("type", protocol.ProtocolMethodMap[types.ProtocolMethod(reader.Kind)]))
			return nil
		}

		data := protocol.IncomingMessageData{
			Original: message,
			Data:     reader.Data,
			Ctx:      context.Background(),
			Peer:     peer,
			VerifyId: verifyId,
			Config:   p.nodeConfig,
			Logger:   p.logger,
			Mediator: NewMediator(service.ServiceParams{
				Logger: p.logger,
				Config: p.nodeConfig,
				Db:     p.db,
			}),
		}

		dec := msgpack.NewDecoder(bytes.NewReader(reader.Data))

		err = handler.DecodeMessage(dec, data)
		if err != nil {
			p.logger.Error("Error decoding message", zap.Error(err))
			return err
		}

		// Directly decode and handle the specific message type
		if err = handler.HandleMessage(data); err != nil {
			p.logger.Error("Error handling message", zap.Error(err))
			return err
		}

		return nil
	}, net.ListenerOptions{
		OnClose: &onDone,
		OnError: &onError,
		Logger:  p.logger,
	})
}

// All other service methods
func (p *P2PServiceDefault) SelfConnectionUris() []*url.URL {
	return p.selfConnectionUris
}

func (p *P2PServiceDefault) Peers() structs.Map {
	return p.peers
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

func (p *P2PServiceDefault) readNodeVotes(nodeId *encoding.NodeId) (service.NodeVotes, error) {
	var value []byte

	value, err := p.db.Get(nodeId.Raw())
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

	err = p.db.Put(nodeId.Raw(), data)
	if err != nil {
		return err
	}

	return nil
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

	err := signedMessage.Sign(p.nodeConfig)

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
			p.logger.Error("failed to cast peer to net.Peer")
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
	p.connectionTracker.Wait()
}

func (p *P2PServiceDefault) ConnectionTracker() *sync.WaitGroup {
	return &p.connectionTracker
}

func (p *P2PServiceDefault) NetworkId() string {
	return p.networkID
}
func (n *P2PServiceDefault) HashQueryRoutingTable() structs.Map {
	return n.hashQueryRoutingTable
}

func (n *P2PServiceDefault) Init(ctx context.Context) error {
	if n.nodeConfig.P2P == nil {
		return errors.New("Nodeconfig is nil")
	}
	if n.nodeConfig.P2P.Peers == nil {
		return errors.New("Nodeconfig P2P peers is nil")
	}
	for _, peer := range n.nodeConfig.P2P.Peers.Blocklist {
		_, err := encoding.DecodeNodeId(peer)
		if err != nil {
			return err
		}

		n.incomingPeerBlockList.Put(peer, true)
		n.outgoingPeerBlocklist.Put(peer, true)
	}
	return nil
}
func (n *P2PServiceDefault) Stop(ctx context.Context) error {
	return nil
}
func (n *P2PServiceDefault) Logger() *zap.Logger {
	return n.logger
}
func (n *P2PServiceDefault) Config() *config.NodeConfig {
	return n.nodeConfig
}
func (n *P2PServiceDefault) Db() kv.KVStore {
	return n.db
}
