package net

import (
	"errors"
	"net/url"
	"sync"
)

var (
	ErrTransportNotSupported = errors.New("no static method registered for type")
)

type TransportPeerConfig struct {
	Socket interface{}
	Uris   []*url.URL
}

type PeerStatic interface {
	Connect(uri *url.URL) (interface{}, error) // Returns a connection/socket
}

type PeerFactory interface {
	NewPeer(options *TransportPeerConfig) (Peer, error)
}

var (
	transports sync.Map
)

func init() {
	transports = sync.Map{}
	RegisterTransport("ws", &WebSocketPeer{})
	RegisterTransport("wss", &WebSocketPeer{})
}
func RegisterTransport(peerType string, factory interface{}) {
	if _, ok := factory.(PeerFactory); !ok {
		panic("factory must implement PeerFactory")
	}

	if _, ok := factory.(PeerStatic); !ok {
		panic("factory must implement PeerStatic")
	}

	transports.Store(peerType, factory)
}

func CreateTransportSocket(peerType string, uri *url.URL) (interface{}, error) {
	static, ok := transports.Load(peerType)
	if !ok {
		return nil, ErrTransportNotSupported
	}

	t, err := static.(PeerStatic).Connect(uri)

	return t, err
}

func CreateTransportPeer(peerType string, options *TransportPeerConfig) (Peer, error) {
	factory, ok := transports.Load(peerType)
	if !ok {
		return nil, errors.New("no factory registered for type: " + peerType)
	}

	t, err := factory.(PeerFactory).NewPeer(options)

	return t, err
}
