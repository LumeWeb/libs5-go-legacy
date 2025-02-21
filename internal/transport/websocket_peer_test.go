package transport

import (
	"context"
	"net"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestWebSocketPeer_SendMessage(t *testing.T) {
	// Create in-memory WebSocket pair
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, server := net.Pipe()
	clientConn, _ := websocket.NewClient(ctx, client, "ws://localhost")
	serverConn, _ := websocket.NewServer(ctx, server, nil)

	peer := WebSocketPeer{
		socket: serverConn,
	}

	testMsg := []byte("test message")

	t.Run("Successful send", func(t *testing.T) {
		err := peer.SendMessage(testMsg)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		_, msg, err := clientConn.Read(ctx)
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		if string(msg) != string(testMsg) {
			t.Errorf("Expected %q, got %q", testMsg, msg)
		}
	})
}

func TestWebSocketPeer_ListenForMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, server := net.Pipe()
	clientConn, _ := websocket.NewClient(ctx, client, "ws://localhost")
	serverConn, _ := websocket.NewServer(ctx, server, nil)

	peer := &WebSocketPeer{
		socket: serverConn,
	}

	t.Run("Message reception", func(t *testing.T) {
		received := make(chan []byte)
		callback := func(msg []byte) error {
			received <- msg
			return nil
		}

		go peer.ListenForMessages(callback, ListenerOptions{})

		testMsg := []byte("test message")
		err := clientConn.Write(ctx, websocket.MessageBinary, testMsg)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case msg := <-received:
			if string(msg) != string(testMsg) {
				t.Errorf("Expected %q, got %q", testMsg, msg)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})

	t.Run("Connection closure", func(t *testing.T) {
		closed := make(chan struct{})
		options := ListenerOptions{
			OnClose: func() { close(closed) },
		}

		go peer.ListenForMessages(func([]byte) error { return nil }, options)
		clientConn.Close(websocket.StatusNormalClosure, "")

		select {
		case <-closed:
			// Expected
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for close callback")
		}
	})
}
