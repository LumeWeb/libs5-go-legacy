package transport

import (
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"net"
	"net/url"
	"sync"
)

type mockPeer struct {
	id            *encoding.NodeId
	connected     bool
	handshakeDone bool
	challenge     []byte
	messages      [][]byte
	socket        interface{}
	uris          []*url.URL
	ip            net.Addr
	isAbuser      bool
	mutex         sync.RWMutex
}

func newMockPeer() *mockPeer {
	return &mockPeer{
		messages: make([][]byte, 0),
		uris:     make([]*url.URL, 0),
	}
}

// Message handling
func (m *mockPeer) SendMessage(message []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockPeer) ListenForMessages(callback EventCallback, options ListenerOptions) {
	// No-op for mock
}

// Connection state
func (m *mockPeer) IsConnected() bool {
	return m.connected
}

func (m *mockPeer) SetConnected(isConnected bool) {
	m.connected = isConnected
}

// Connection information
func (m *mockPeer) RenderLocationURI() string {
	if len(m.uris) > 0 {
		return m.uris[0].String()
	}
	return ""
}

func (m *mockPeer) ConnectionURIs() []*url.URL {
	return m.uris
}

func (m *mockPeer) SetConnectionURIs(uris []*url.URL) {
	m.uris = uris
}

// Connection lifecycle
func (m *mockPeer) End() error {
	m.connected = false
	return nil
}

func (m *mockPeer) EndForAbuse() error {
	m.connected = false
	m.isAbuser = true
	return nil
}

// Identification
func (m *mockPeer) Id() *encoding.NodeId {
	return m.id
}

func (m *mockPeer) SetId(id *encoding.NodeId) {
	m.id = id
}

// Handshake
func (m *mockPeer) Challenge() []byte {
	return m.challenge
}

func (m *mockPeer) SetChallenge(challenge []byte) {
	m.challenge = challenge
}

func (m *mockPeer) IsHandshakeDone() bool {
	return m.handshakeDone
}

func (m *mockPeer) SetHandshakeDone(status bool) {
	m.handshakeDone = status
}

// Socket management
func (m *mockPeer) Socket() interface{} {
	return m.socket
}

func (m *mockPeer) SetSocket(socket interface{}) {
	m.socket = socket
}

// IP address
func (m *mockPeer) GetIPString() string {
	if m.ip != nil {
		return m.ip.String()
	}
	return "mock-ip"
}

func (m *mockPeer) GetIP() net.Addr {
	return m.ip
}

func (m *mockPeer) SetIP(ip net.Addr) {
	m.ip = ip
}

// Status
func (m *mockPeer) Abuser() bool {
	return m.isAbuser
}
