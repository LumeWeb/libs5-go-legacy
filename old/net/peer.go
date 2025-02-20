package net

import (
	"go.lumeweb.com/libs5-go/encoding"
	"go.uber.org/zap"
	"net"
	"net/url"
	"sync"
)

var (
	_ Peer = (*BasePeer)(nil)
)

// EventCallback type for the callback function
type EventCallback func(event []byte) error

// CloseCallback type for the OnClose callback
type CloseCallback func()

// ErrorCallback type for the onError callback
type ErrorCallback func(args ...interface{})

// ListenerOptions struct for options
type ListenerOptions struct {
	OnClose *CloseCallback
	OnError *ErrorCallback
	Logger  *zap.Logger
}

type Peer interface {
	SendMessage(message []byte) error
	RenderLocationURI() string
	ListenForMessages(callback EventCallback, options ListenerOptions)
	End() error
	EndForAbuse() error
	SetId(id *encoding.NodeId)
	Id() *encoding.NodeId
	SetChallenge(challenge []byte)
	Challenge() []byte
	SetSocket(socket interface{})
	Socket() interface{}
	SetConnected(isConnected bool)
	IsConnected() bool
	SetConnectionURIs(uris []*url.URL)
	ConnectionURIs() []*url.URL
	IsHandshakeDone() bool
	SetHandshakeDone(status bool)
	GetIPString() string
	GetIP() net.Addr
	SetIP(ip net.Addr)
	Abuser() bool
}

type BasePeer struct {
	connectionURIs []*url.URL
	isConnected    bool
	challenge      []byte
	socket         interface{}
	id             *encoding.NodeId
	handshaked     bool
	lock           sync.RWMutex
}

func (b *BasePeer) IsConnected() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.isConnected
}

func (b *BasePeer) SetConnected(isConnected bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.isConnected = isConnected
}

func (b *BasePeer) SendMessage(message []byte) error {
	panic("must implement in child class")
}

func (b *BasePeer) RenderLocationURI() string {
	panic("must implement in child class")
}

func (b *BasePeer) ListenForMessages(callback EventCallback, options ListenerOptions) {
	panic("must implement in child class")
}

func (b *BasePeer) End() error {
	panic("must implement in child class")
}
func (b *BasePeer) EndForAbuse() error {
	panic("must implement in child class")
}
func (b *BasePeer) GetIPString() string {
	panic("must implement in child class")
}

func (b *BasePeer) GetIP() net.Addr {
	panic("must implement in child class")
}

func (b *BasePeer) SetIP(ip net.Addr) {
	panic("must implement in child class")
}

func (b *BasePeer) Challenge() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.challenge
}

func (b *BasePeer) SetChallenge(challenge []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.challenge = challenge
}

func (b *BasePeer) Socket() interface{} {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.socket
}

func (b *BasePeer) SetSocket(socket interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.socket = socket
}

func (b *BasePeer) Id() *encoding.NodeId {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.id
}

func (b *BasePeer) SetId(id *encoding.NodeId) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.id = id
}
func (b *BasePeer) SetConnectionURIs(uris []*url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.connectionURIs = uris
}
func (b *BasePeer) ConnectionURIs() []*url.URL {
	b.lock.RLock()
	b.lock.RUnlock()
	return b.connectionURIs
}

func (b *BasePeer) IsHandshakeDone() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.handshaked
}

func (b *BasePeer) SetHandshakeDone(status bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.handshaked = status
}

func (b *BasePeer) Abuser() bool {
	panic("must implement in child class")
}
