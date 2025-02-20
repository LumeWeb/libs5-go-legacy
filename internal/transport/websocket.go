package transport

import (
	"context"
	"errors"
	"net"
	"net/url"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

var (
	_ PeerFactory = (*WebSocketPeer)(nil)
	_ PeerStatic  = (*WebSocketPeer)(nil)
	_ Peer        = (*WebSocketPeer)(nil)
)

// WebSocketPeer implements Peer for WebSocket connections
type WebSocketPeer struct {
	BasePeer
	socket *websocket.Conn
	abuser bool
	ip     net.Addr
}

// Connect establishes a WebSocket connection
func (p *WebSocketPeer) Connect(uri *url.URL) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dial, _, err := websocket.Dial(ctx, uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return dial, nil
}

// NewPeer creates a new peer from a socket
func (p *WebSocketPeer) NewPeer(options *TransportPeerConfig) (Peer, error) {
	ws, ok := options.Socket.(*websocket.Conn)
	if !ok {
		return nil, errors.New("socket is not a websocket connection")
	}

	peer := &WebSocketPeer{
		BasePeer: BasePeer{
			connectionURIs: options.Uris,
			socket:         options.Socket,
		},
		socket: ws,
	}

	return peer, nil
}

// SendMessage sends a message over the WebSocket
func (p *WebSocketPeer) SendMessage(message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.socket.Write(ctx, websocket.MessageBinary, message)
	if err != nil {
		return err
	}

	return nil
}

// RenderLocationURI returns a description of the peer location
func (p *WebSocketPeer) RenderLocationURI() string {
	return "WebSocket client"
}

// ListenForMessages listens for WebSocket messages
func (p *WebSocketPeer) ListenForMessages(callback EventCallback, options ListenerOptions) {
	errChan := make(chan error, 10)
	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	for {
		// Read message with context that will be canceled on socket close
		ctx := context.Background()
		messageType, message, err := p.socket.Read(ctx)

		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				if options.OnError != nil {
					(*options.OnError)(err)
				}
			}
			break
		}

		if messageType != websocket.MessageBinary {
			// Skip non-binary messages
			continue
		}

		wg.Add(1)
		// Process each message in a separate goroutine
		go func(msg []byte) {
			defer wg.Done()
			// Call the callback and send any errors to the error channel
			if err := callback(msg); err != nil {
				select {
				case errChan <- err:
				case <-doneChan:
					// Stop sending errors if doneChan is closed
				}
			}
		}(message)

		// Non-blocking error check
		select {
		case err := <-errChan:
			if options.OnError != nil {
				(*options.OnError)(err)
			}
		default:
		}
	}

	if options.OnClose != nil {
		(*options.OnClose)()
	}

	// Close doneChan and wait for all goroutines to finish
	close(doneChan)
	wg.Wait()

	// Handle remaining errors
	close(errChan)
	for err := range errChan {
		if options.OnError != nil {
			(*options.OnError)(err)
		}
	}
}

// End closes the WebSocket connection normally
func (p *WebSocketPeer) End() error {
	err := p.socket.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		return err
	}

	return nil
}

// EndForAbuse closes the connection due to abuse
func (p *WebSocketPeer) EndForAbuse() error {
	p.BasePeer.lock.Lock()
	defer p.BasePeer.lock.Unlock()
	p.abuser = true
	err := p.socket.Close(websocket.StatusPolicyViolation, "")
	if err != nil {
		return err
	}

	return nil
}

// GetIP returns the peer's IP address
func (p *WebSocketPeer) GetIP() net.Addr {
	p.BasePeer.lock.RLock()
	defer p.BasePeer.lock.RUnlock()

	if p.ip != nil {
		return p.ip
	}

	// Get the IP from the WebSocket connection
	ctx, cancel := context.WithCancel(context.Background())
	netConn := websocket.NetConn(ctx, p.socket, websocket.MessageBinary)
	ipAddr := netConn.RemoteAddr()
	cancel()

	return ipAddr
}

// SetIP sets the peer's IP address
func (p *WebSocketPeer) SetIP(ip net.Addr) {
	p.BasePeer.lock.Lock()
	defer p.BasePeer.lock.Unlock()
	p.ip = ip
}

// GetIPString returns a string representation of the IP
func (p *WebSocketPeer) GetIPString() string {
	return p.GetIP().String()
}

// Abuser returns whether the peer is marked as an abuser
func (p *WebSocketPeer) Abuser() bool {
	p.BasePeer.lock.RLock()
	defer p.BasePeer.lock.RUnlock()
	return p.abuser
}
