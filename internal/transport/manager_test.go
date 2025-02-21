package transport_test

import (
	"context"
	"go.lumeweb.com/libs5-go/internal/transport"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.uber.org/zap"
	"net/url"
	"testing"
)

func createTestManager(_ *testing.T) transport.Manager {
	kp, _ := crypto.GenerateEd25519Key()
	logger, _ := zap.NewDevelopment()
	return transport.NewManager(kp, crypto.NewDefaultCrypto(), logger)
}

func TestManager_Connect(t *testing.T) {
	mgr := createTestManager(t)
	testURL, _ := url.Parse("ws://localhost:8080")

	t.Run("New connection", func(t *testing.T) {
		peer, err := mgr.Connect(context.Background(), testURL)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		if peer == nil {
			t.Fatal("Expected peer, got nil")
		}

		if len(mgr.AllPeers()) != 1 {
			t.Errorf("Expected 1 peer, got %d", len(mgr.AllPeers()))
		}
	})

	t.Run("Duplicate connection", func(t *testing.T) {
		_, err := mgr.Connect(context.Background(), testURL)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		if len(mgr.AllPeers()) != 1 {
			t.Errorf("Expected 1 peer after duplicate connect, got %d", len(mgr.AllPeers()))
		}
	})
}

func TestManager_Broadcast(t *testing.T) {
	mgr := createTestManager(t)
	testURL, _ := url.Parse("ws://localhost:8080")

	peer, _ := mgr.Connect(context.Background(), testURL)
	peer.SetConnected(true)

	testMsg := []byte("broadcast test")

	t.Run("Successful broadcast", func(t *testing.T) {
		err := mgr.Broadcast(testMsg)
		if err != nil {
			t.Fatalf("Broadcast failed: %v", err)
		}
	})

	t.Run("Skip peer", func(t *testing.T) {
		id, _ := peer.Id().ToString()
		err := mgr.Broadcast(testMsg, id)
		if err != nil {
			t.Fatalf("Broadcast failed: %v", err)
		}
	})
}
