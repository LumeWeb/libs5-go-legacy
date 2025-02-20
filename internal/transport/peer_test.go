package transport_test

import (
	"net/url"
	"testing"

	"go.lumeweb.com/libs5-go/internal/transport"
	"go.lumeweb.com/libs5-go/pkg/encoding"
)

func TestBasePeer_ConnectionState(t *testing.T) {
	peer := &transport.BasePeer{}

	t.Run("Default state", func(t *testing.T) {
		if peer.IsConnected() {
			t.Error("New peer should not be connected")
		}
	})

	t.Run("Set connected", func(t *testing.T) {
		peer.SetConnected(true)
		if !peer.IsConnected() {
			t.Error("Peer should be connected")
		}
	})
}

func TestPeer_URIs(t *testing.T) {
	testURIs := []*url.URL{
		{Scheme: "ws", Host: "host1"},
		{Scheme: "wss", Host: "host2"},
	}

	peer := &transport.BasePeer{}
	peer.SetConnectionURIs(testURIs)

	if len(peer.ConnectionURIs()) != 2 {
		t.Errorf("Expected 2 URIs, got %d", len(peer.ConnectionURIs()))
	}
}

func TestPeer_Identifiers(t *testing.T) {
	id := encoding.NewNodeId([]byte("test-id"))
	peer := &transport.BasePeer{}
	peer.SetId(id)

	if peer.Id() == nil {
		t.Fatal("Expected non-nil ID")
	}
}
