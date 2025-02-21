package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		defer func() {
			err := conn.Close(websocket.StatusInternalError, "")
			if err != nil {
				t.Error(err)
			}
		}()

		// Read the test message
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			t.Fatal(err)
			return
		}
		err = conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			t.Error(err)
		}

		if string(msg) != "test message" {
			t.Errorf("Expected 'test message', got %q", string(msg))
		}
	}))
	defer s.Close()

	// Connect client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := c.Close(websocket.StatusInternalError, "")
		if err != nil {
			t.Error(err)
		}
	}()

	peer := &WebSocketPeer{
		socket: c,
	}

	testMsg := []byte("test message")
	if err := peer.SendMessage(testMsg); err != nil {
		t.Fatal(err)
	}

	err = c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
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
		defer func() {
			err := conn.Close(websocket.StatusInternalError, "")
			if err != nil {
				t.Error(err)
			}
		}()

		// Send test message
		err = conn.Write(context.Background(), websocket.MessageBinary, []byte("test message"))
		if err != nil {
			t.Fatal(err)
		}

		// Wait briefly then close normally
		time.Sleep(100 * time.Millisecond)
		err = conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	peer := &WebSocketPeer{
		socket: c,
	}

	received := make(chan []byte)
	closed := make(chan struct{})

	onClose := func() { close(closed) }
	options := ListenerOptions{
		OnClose: onClose,
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
