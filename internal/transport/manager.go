package transport

import (
	"context"
	"go.lumeweb.com/libs5-go/old/encoding"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager handles peer connections and network transport
type Manager interface {
	// Connection management
	Connect(ctx context.Context, uri *url.URL) (Peer, error)
	Disconnect(peerID string) error
	GetPeer(peerID string) (Peer, bool)
	AllPeers() []Peer
	ConnectedPeers() []Peer

	// Network state
	IsConnectedToNetwork() bool
	WaitForConnection(ctx context.Context) error

	// Message broadcasting
	Broadcast(message []byte, skipPeers ...string) error
	BroadcastToConnected(message []byte, skipPeers ...string) error

	// Node identity
	NodeID() *encoding.NodeId

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	peers        map[string]Peer        // All peers by ID
	peersByURI   map[string]Peer        // Peers indexed by URI
	nodeID       *encoding.NodeId       // This node's ID
	keyPair      *crypto.KeyPairEd25519 // This node's keypair
	transports   map[string]PeerFactory // Available transport types
	logger       *zap.Logger
	mutex        sync.RWMutex
	challengeGen func() []byte // Challenge generator function
}

type ManagerOption func(*DefaultManager)

// NewManager creates a transport manager with the given options
func NewManager(keyPair *crypto.KeyPairEd25519, logger *zap.Logger, options ...ManagerOption) Manager {
	nodeID := encoding.NewNodeId(keyPair.PublicKey())

	m := &DefaultManager{
		peers:      make(map[string]Peer),
		peersByURI: make(map[string]Peer),
		nodeID:     nodeID,
		keyPair:    keyPair,
		transports: make(map[string]PeerFactory),
		logger:     logger,
		challengeGen: func() []byte {
			challenge := make([]byte, 64)
			// Generate random challenge and embed public key
			// similar to JS implementation
			crypto.GenerateSecureRandomBytes(challenge)
			copy(challenge[31:], keyPair.PublicKeyRaw())
			return challenge
		},
	}

	// Register default transports
	m.transports["ws"] = &WebSocketPeer{}
	m.transports["wss"] = &WebSocketPeer{}

	// Apply options
	for _, option := range options {
		option(m)
	}

	return m
}

// WithCustomChallengeGenerator sets a custom challenge generator
func WithCustomChallengeGenerator(generator func() []byte) ManagerOption {
	return func(m *DefaultManager) {
		m.challengeGen = generator
	}
}

// WithTransport registers a transport type
func WithTransport(scheme string, factory PeerFactory) ManagerOption {
	return func(m *DefaultManager) {
		m.transports[scheme] = factory
	}
}

func (m *DefaultManager) Connect(ctx context.Context, uri *url.URL) (Peer, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if we already have a connection to this URI
	if peer, exists := m.peersByURI[uri.String()]; exists {
		return peer, nil
	}

	// Get transport factory for this scheme
	factory, ok := m.transports[uri.Scheme]
	if !ok {
		return nil, ErrTransportNotSupported
	}

	// Create transport socket
	socket, err := factory.(PeerStatic).Connect(uri)
	if err != nil {
		return nil, err
	}

	// Create peer with socket
	config := &TransportPeerConfig{
		Socket: socket,
		Uris:   []*url.URL{uri},
	}

	peer, err := factory.NewPeer(config)
	if err != nil {
		return nil, err
	}

	// Set up challenge for handshake
	challenge := m.challengeGen()
	peer.SetChallenge(challenge)

	// Store the peer
	if peerID := peer.Id(); peerID != nil {
		idStr, _ := peerID.ToString()
		m.peers[idStr] = peer
	}
	m.peersByURI[uri.String()] = peer

	// Set up message handler
	m.setupMessageHandler(peer)

	return peer, nil
}

func (m *DefaultManager) setupMessageHandler(peer Peer) {
	// Set up listener options
	onClose := func() {
		m.handlePeerDisconnect(peer)
	}
	onError := func(args ...interface{}) {
		m.logger.Error("Peer error", zap.Any("error", args), zap.String("peer", peer.GetIPString()))
	}

	// Start listening for messages
	go peer.ListenForMessages(
		func(event []byte) error {
			// Message handling logic (will be moved to protocol layer)
			return nil
		},
		ListenerOptions{
			OnClose: &onClose,
			OnError: &onError,
			Logger:  m.logger,
		},
	)
}

func (m *DefaultManager) handlePeerDisconnect(peer Peer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove from maps
	if peerID := peer.Id(); peerID != nil {
		idStr, _ := peerID.ToString()
		delete(m.peers, idStr)
	}

	// Remove from URI map
	for uri, p := range m.peersByURI {
		if p == peer {
			delete(m.peersByURI, uri)
		}
	}
}

func (m *DefaultManager) GetPeer(peerID string) (Peer, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	peer, exists := m.peers[peerID]
	return peer, exists
}

func (m *DefaultManager) AllPeers() []Peer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	peers := make([]Peer, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (m *DefaultManager) ConnectedPeers() []Peer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	peers := make([]Peer, 0)
	for _, peer := range m.peers {
		if peer.IsConnected() {
			peers = append(peers, peer)
		}
	}
	return peers
}

func (m *DefaultManager) IsConnectedToNetwork() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, peer := range m.peers {
		if peer.IsConnected() {
			return true
		}
	}
	return false
}

func (m *DefaultManager) WaitForConnection(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if m.IsConnectedToNetwork() {
				return nil
			}
		}
	}
}

func (m *DefaultManager) Broadcast(message []byte, skipPeers ...string) error {
	skipMap := make(map[string]struct{})
	for _, id := range skipPeers {
		skipMap[id] = struct{}{}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for id, peer := range m.peers {
		if _, skip := skipMap[id]; skip {
			continue
		}

		if err := peer.SendMessage(message); err != nil {
			m.logger.Warn("Failed to send message to peer",
				zap.String("peerID", id),
				zap.Error(err))
		}
	}

	return nil
}

func (m *DefaultManager) BroadcastToConnected(message []byte, skipPeers ...string) error {
	skipMap := make(map[string]struct{})
	for _, id := range skipPeers {
		skipMap[id] = struct{}{}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for id, peer := range m.peers {
		if _, skip := skipMap[id]; skip {
			continue
		}

		if peer.IsConnected() {
			if err := peer.SendMessage(message); err != nil {
				m.logger.Warn("Failed to send message to connected peer",
					zap.String("peerID", id),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (m *DefaultManager) NodeID() *encoding.NodeId {
	return m.nodeID
}

func (m *DefaultManager) Disconnect(peerID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	peer, exists := m.peers[peerID]
	if !exists {
		return nil // Already disconnected
	}

	// End the connection
	err := peer.End()
	if err != nil {
		return err
	}

	// Remove from maps
	delete(m.peers, peerID)

	// Remove from URI map
	for uri, p := range m.peersByURI {
		if p == peer {
			delete(m.peersByURI, uri)
		}
	}

	return nil
}

func (m *DefaultManager) Start(ctx context.Context) error {
	// No initialization needed for now
	return nil
}

func (m *DefaultManager) Stop(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Disconnect all peers
	for id, peer := range m.peers {
		err := peer.End()
		if err != nil {
			m.logger.Warn("Error disconnecting peer",
				zap.String("peerID", id),
				zap.Error(err))
		}
	}

	// Clear maps
	m.peers = make(map[string]Peer)
	m.peersByURI = make(map[string]Peer)

	return nil
}
