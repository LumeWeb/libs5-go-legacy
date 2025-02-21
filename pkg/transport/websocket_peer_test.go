package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestWebSocketPeer_SendMessage(t *testing.T) {
	t.Parallel()

	// Create test server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		// Read the test message
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			t.Fatal(err)
			return
		}

		if string(msg) != "test message" {
			t.Errorf("Expected 'test message', got %q", string(msg))
		}

		// Close cleanly after reading
		if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	// Convert http URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	peer := &WebSocketPeer{
		socket: c,
	}

	testMsg := []byte("test message")
	if err := peer.SendMessage(testMsg); err != nil {
		t.Fatal(err)
	}

	// Close cleanly after sending
	if err := peer.End(); err != nil {
		t.Error(err)
	}
}

func TestWebSocketPeer_ListenForMessages(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		// Send test message
		err = conn.Write(context.Background(), websocket.MessageBinary, []byte("test message"))
		if err != nil {
			t.Fatal(err)
		}

		// Wait briefly then close normally
		time.Sleep(100 * time.Millisecond)
		if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	// Convert http URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	peer := &WebSocketPeer{
		socket: c,
	}

	received := make(chan []byte)
	closed := make(chan struct{})

	options := ListenerOptions{
		OnClose: func() { close(closed) },
	}

	go peer.ListenForMessages(func(msg []byte) error {
		received <- msg
		return nil
	}, options)

	// Test message reception
	select {
	case msg := <-received:
		if string(msg) != "test message" {
			t.Errorf("Expected 'test message', got %q", string(msg))
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Test connection closure
	select {
	case <-closed:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for close callback")
	}
}
