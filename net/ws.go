package net

import (
	"context"
	"github.com/LumeWeb/libs5-go/encoding"
	"net"
	"net/url"
	"nhooyr.io/websocket"
	"sync"
)

var (
	_ PeerFactory = (*WebSocketPeer)(nil)
	_ PeerStatic  = (*WebSocketPeer)(nil)
	_ Peer        = (*WebSocketPeer)(nil)
)

type WebSocketPeer struct {
	BasePeer
	socket *websocket.Conn
	abuser bool
	ip     net.Addr
}

func (p *WebSocketPeer) Connect(uri *url.URL) (interface{}, error) {
	dial, _, err := websocket.Dial(context.Background(), uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return dial, nil
}

func (p *WebSocketPeer) NewPeer(options *TransportPeerConfig) (Peer, error) {
	peer := &WebSocketPeer{
		BasePeer: BasePeer{
			connectionURIs: options.Uris,
			socket:         options.Socket,
		},
		socket: options.Socket.(*websocket.Conn),
	}

	return peer, nil
}

func (p *WebSocketPeer) SendMessage(message []byte) error {
	err := p.socket.Write(context.Background(), websocket.MessageBinary, message)
	if err != nil {
		return err
	}

	return nil
}

func (p *WebSocketPeer) RenderLocationURI() string {
	return "WebSocket client"
}

func (p *WebSocketPeer) ListenForMessages(callback EventCallback, options ListenerOptions) {
	errChan := make(chan error, 10)
	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	for {
		_, message, err := p.socket.Read(context.Background())
		if err != nil {
			if options.OnError != nil {
				(*options.OnError)(err)
			}
			break
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

func (p *WebSocketPeer) End() error {
	err := p.socket.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		return err
	}

	return nil
}
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
func (p *WebSocketPeer) SetId(id *encoding.NodeId) {
	p.BasePeer.lock.Lock()
	defer p.BasePeer.lock.Unlock()
	p.id = id
}

func (p *WebSocketPeer) SetChallenge(challenge []byte) {
	p.BasePeer.lock.Lock()
	defer p.BasePeer.lock.Unlock()
	p.challenge = challenge
}

func (p *WebSocketPeer) GetChallenge() []byte {
	p.BasePeer.lock.RLock()
	defer p.BasePeer.lock.RUnlock()
	return p.challenge
}

func (p *WebSocketPeer) GetIP() net.Addr {
	p.BasePeer.lock.RLock()
	defer p.BasePeer.lock.RUnlock()
	if p.ip != nil {
		return p.ip
	}

	ctx, cancel := context.WithCancel(context.Background())
	netConn := websocket.NetConn(ctx, p.socket, websocket.MessageBinary)

	ipAddr := netConn.RemoteAddr()

	cancel()

	return ipAddr
}

func (p *WebSocketPeer) SetIP(ip net.Addr) {
	p.BasePeer.lock.Lock()
	defer p.BasePeer.lock.Unlock()
	p.ip = ip
}

func (b *WebSocketPeer) GetIPString() string {
	return b.GetIP().String()
}

func (p *WebSocketPeer) Abuser() bool {
	p.BasePeer.lock.RLock()
	defer p.BasePeer.lock.RUnlock()
	return p.abuser
}
