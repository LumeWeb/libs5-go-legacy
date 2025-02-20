package transport

import (
	"errors"
	"go.lumeweb.com/libs5-go/old/encoding"
	"go.uber.org/zap"
	"net"
	"net/url"
	"sync"
)

var (
	ErrTransportNotSupported = errors.New("no static method registered for type")
)

// EventCallback is a function that handles incoming messages
type EventCallback func(event []byte) error

// CloseCallback is called when a peer connection closes
type CloseCallback func()

// ErrorCallback is called when a peer connection encounters an error
type ErrorCallback func(args ...interface{})

// ListenerOptions contains options for message listeners
type ListenerOptions struct {
	OnClose *CloseCallback
	OnError *ErrorCallback
	Logger  *zap.Logger
}

// TransportPeerConfig contains configuration for new peers
type TransportPeerConfig struct {
	Socket interface{}
	Uris   []*url.URL
}

// Peer represents a network connection to another node
type Peer interface {
	// Message handling
	SendMessage(message []byte) error
	ListenForMessages(callback EventCallback, options ListenerOptions)

	// Connection state
	IsConnected() bool
	SetConnected(isConnected bool)

	// Connection information
	RenderLocationURI() string
	ConnectionURIs() []*url.URL
	SetConnectionURIs(uris []*url.URL)

	// Connection lifecycle
	End() error
	EndForAbuse() error

	// Identification
	Id() *encoding.NodeId
	SetId(id *encoding.NodeId)

	// Handshake
	Challenge() []byte
	SetChallenge(challenge []byte)
	IsHandshakeDone() bool
	SetHandshakeDone(status bool)

	// Socket management
	Socket() interface{}
	SetSocket(socket interface{})

	// IP address
	GetIPString() string
	GetIP() net.Addr
	SetIP(ip net.Addr)

	// Status
	Abuser() bool
}

// PeerFactory creates new peers
type PeerFactory interface {
	NewPeer(options *TransportPeerConfig) (Peer, error)
}

// PeerStatic handles static connection methods
type PeerStatic interface {
	Connect(uri *url.URL) (interface{}, error)
}

// BasePeer implements common Peer functionality
type BasePeer struct {
	connectionURIs []*url.URL
	isConnected    bool
	challenge      []byte
	socket         interface{}
	id             *encoding.NodeId
	handshaked     bool
	lock           sync.RWMutex
}

// IsConnected returns whether the peer is connected
func (b *BasePeer) IsConnected() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.isConnected
}

// SetConnected sets the connected state
func (b *BasePeer) SetConnected(isConnected bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.isConnected = isConnected
}

// SendMessage sends a message to the peer
func (b *BasePeer) SendMessage(message []byte) error {
	panic("must implement in child class")
}

// RenderLocationURI returns a string representation of the peer's location
func (b *BasePeer) RenderLocationURI() string {
	panic("must implement in child class")
}

// ListenForMessages listens for incoming messages
func (b *BasePeer) ListenForMessages(callback EventCallback, options ListenerOptions) {
	panic("must implement in child class")
}

// End gracefully ends the peer connection
func (b *BasePeer) End() error {
	panic("must implement in child class")
}

// EndForAbuse ends the peer connection due to abuse
func (b *BasePeer) EndForAbuse() error {
	panic("must implement in child class")
}

// GetIPString returns a string representation of the peer's IP
func (b *BasePeer) GetIPString() string {
	panic("must implement in child class")
}

// GetIP returns the peer's IP address
func (b *BasePeer) GetIP() net.Addr {
	panic("must implement in child class")
}

// SetIP sets the peer's IP address
func (b *BasePeer) SetIP(ip net.Addr) {
	panic("must implement in child class")
}

// Challenge returns the challenge used for handshaking
func (b *BasePeer) Challenge() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.challenge
}

// SetChallenge sets the challenge for handshaking
func (b *BasePeer) SetChallenge(challenge []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.challenge = challenge
}

// Socket returns the underlying socket
func (b *BasePeer) Socket() interface{} {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.socket
}

// SetSocket sets the underlying socket
func (b *BasePeer) SetSocket(socket interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.socket = socket
}

// Id returns the peer's node ID
func (b *BasePeer) Id() *encoding.NodeId {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.id
}

// SetId sets the peer's node ID
func (b *BasePeer) SetId(id *encoding.NodeId) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.id = id
}

// SetConnectionURIs sets the peer's connection URIs
func (b *BasePeer) SetConnectionURIs(uris []*url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.connectionURIs = uris
}

// ConnectionURIs returns the peer's connection URIs
func (b *BasePeer) ConnectionURIs() []*url.URL {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.connectionURIs
}

// IsHandshakeDone returns whether the handshake is complete
func (b *BasePeer) IsHandshakeDone() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.handshaked
}

// SetHandshakeDone sets the handshake status
func (b *BasePeer) SetHandshakeDone(status bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.handshaked = status
}

// Abuser returns whether the peer is an abuser
func (b *BasePeer) Abuser() bool {
	panic("must implement in child class")
}
