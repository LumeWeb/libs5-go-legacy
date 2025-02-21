package transport

import (
	"context"
	"net/url"
	"testing"

	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.uber.org/zap"
)

// socket defines a network connection abstraction
type socket interface {
	Close() error
}

// mockSocket implements socket for testing
type mockSocket struct {
	closed bool
}

func (s *mockSocket) Close() error {
	s.closed = true
	return nil
}

// mockPeerFactory implements PeerFactory for testing
type mockPeerFactory struct {
	peers map[string]*mockPeer
}

func newMockPeerFactory() *mockPeerFactory {
	return &mockPeerFactory{
		peers: make(map[string]*mockPeer),
	}
}

// NewPeer creates a new mock peer
func (f *mockPeerFactory) NewPeer(config *TransportPeerConfig) (Peer, error) {
	peer := newMockPeer()
	peer.SetSocket(config.Socket)
	peer.SetConnectionURIs(config.Uris)

	uri := config.Uris[0]
	f.peers[uri.String()] = peer
	return peer, nil
}

// Connect returns a mock socket
func (f *mockPeerFactory) Connect(uri *url.URL) (socket, error) {
	return &mockSocket{}, nil
}

// Test helper function
func createTestManager(t *testing.T) (Manager, *mockPeerFactory) {
	kp, err := crypto.GenerateEd25519Key()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	mockFactory := newMockPeerFactory()

	manager := NewManager(
		kp,
		crypto.NewDefaultCrypto(),
		logger,
		WithTransport("ws", mockFactory),
	)

	return manager, mockFactory
}

func TestManager_Connect(t *testing.T) {
	mgr, factory := createTestManager(t)
	testURL, _ := url.Parse("ws://localhost:8080")

	t.Run("New connection", func(t *testing.T) {
		peer, err := mgr.Connect(context.Background(), testURL)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		if peer == nil {
			t.Fatal("Expected peer, got nil")
		}

		mockPeer := factory.peers[testURL.String()]
		if mockPeer == nil {
			t.Fatal("Mock peer not created")
		}

		// Verify the peer was added to the manager
		if len(mgr.AllPeers()) != 1 {
			t.Errorf("Expected 1 peer, got %d", len(mgr.AllPeers()))
		}
	})

	t.Run("Duplicate connection", func(t *testing.T) {
		peer1, err := mgr.Connect(context.Background(), testURL)
		if err != nil {
			t.Fatalf("First connect failed: %v", err)
		}

		peer2, err := mgr.Connect(context.Background(), testURL)
		if err != nil {
			t.Fatalf("Second connect failed: %v", err)
		}

		// Should return the same peer
		if peer1 != peer2 {
			t.Error("Expected same peer instance for duplicate connection")
		}

		if len(mgr.AllPeers()) != 1 {
			t.Errorf("Expected 1 peer after duplicate connect, got %d", len(mgr.AllPeers()))
		}
	})
}

func TestManager_Broadcast(t *testing.T) {
	mgr, _ := createTestManager(t)
	testURL, _ := url.Parse("ws://localhost:8080")

	// Create and connect a peer
	peer, err := mgr.Connect(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	mockPeer, ok := peer.(*mockPeer)
	if !ok {
		t.Fatal("Failed to cast peer to mockPeer")
	}

	// Set peer as connected
	mockPeer.SetConnected(true)

	// Set a node ID for the peer
	nodeID := encoding.NewNodeId([]byte("test-peer-id"))
	mockPeer.SetId(nodeID)

	testMsg := []byte("broadcast test")

	t.Run("Successful broadcast", func(t *testing.T) {
		err := mgr.Broadcast(testMsg)
		if err != nil {
			t.Fatalf("Broadcast failed: %v", err)
		}

		// Verify message was sent to peer
		if len(mockPeer.messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(mockPeer.messages))
		}
	})

	t.Run("Skip peer", func(t *testing.T) {
		id, _ := mockPeer.Id().ToString()
		err := mgr.Broadcast(testMsg, id)
		if err != nil {
			t.Fatalf("Broadcast failed: %v", err)
		}

		// Message count should not increase since peer was skipped
		if len(mockPeer.messages) != 1 {
			t.Errorf("Expected 1 message (unchanged), got %d", len(mockPeer.messages))
		}
	})
}
