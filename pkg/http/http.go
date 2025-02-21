package http

import (
	"context"
	"go.lumeweb.com/libs5-go/internal/transport"
	"go.lumeweb.com/libs5-go/pkg/service"
	transport2 "go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
)

var _ service.H = (*HTTPServiceDefault)(nil)

type P2PNodesResponse struct {
	Nodes []P2PNodeResponse `json:"nodes"`
}

type P2PNodeResponse struct {
	Id   string   `json:"id"`
	Uris []string `json:"uris"`
}

type HTTPServiceDefault struct {
	service.ServiceBase
}

func NewHTTP(params service.ServiceParams) *HTTPServiceDefault {
	return &HTTPServiceDefault{
		ServiceBase: service.NewServiceBase(params.Logger, params.Config, params.Db),
	}
}

func (h *HTTPServiceDefault) GetHttpRouter() map[string]http.HandlerFunc {
	routes := map[string]http.HandlerFunc{
		"GET /s5/version":   h.versionHandler,
		"GET /s5/p2p":       h.p2pHandler,
		"GET /s5/p2p/nodes": h.p2pNodesHandler,
		"GET /s5/p2p/peers": h.p2pPeersHandler,
	}

	return routes
}

func (h *HTTPServiceDefault) Start(ctx context.Context) error {
	return nil
}

func (h *HTTPServiceDefault) Stop(ctx context.Context) error {
	return nil
}

func (h *HTTPServiceDefault) Init(ctx context.Context) error {
	return nil
}

func (h *HTTPServiceDefault) versionHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(build.Version))
}
func (h *HTTPServiceDefault) p2pHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.Logger().Error("error accepting websocket connection", zap.Error(err))
		return
	}

	peer, err := transport.CreateTransportPeer("wss", &transport2.TransportPeerConfig{
		Socket: c,
		Uris:   []*url.URL{},
	})

	if err != nil {
		h.Logger().Error("error creating transport peer", zap.Error(err))
		err := c.Close(websocket.StatusInternalError, "the sky is falling")
		if err != nil {
			h.Logger().Error("error closing websocket connection", zap.Error(err))
		}
		return
	}

	// Check for reverse proxy headers
	realIP := r.Header.Get("X-Real-IP")
	forwardedFor := r.Header.Get("X-Forwarded-For")

	var clientIP net.IP
	if realIP != "" {
		clientIP = net.ParseIP(realIP)
	} else if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IP addresses separated by commas
		// We take the first IP in the list as the client's IP
		parts := strings.Split(forwardedFor, ",")
		clientIP = net.ParseIP(parts[0])
	}

	blockConnection := func(ip net.Addr) bool {
		// If we have a valid client IP from headers, use that for the loopback check
		if clientIP != nil {
			return clientIP.IsLoopback()
		}
		// Otherwise, fall back to the peer's IP
		switch v := ip.(type) {
		case *net.IPNet:
			return v.IP.IsLoopback()
		case *net.TCPAddr:
			return v.IP.IsLoopback()
		default:
			return false
		}
	}

	if blockConnection(peer.GetIP()) {
		err := peer.End()
		if err != nil {
			h.Logger().Error("error ending peer", zap.Error(err))
		}
		return
	}

	if clientIP != nil {
		peer.SetIP(&net.IPAddr{IP: clientIP})
	}

	h.Services().P2P().ConnectionTracker().Add(1)

	go func() {
		err := h.Services().P2P().OnNewPeer(peer, false)
		if err != nil {
			h.Logger().Error("error handling new peer", zap.Error(err))
		}
		h.Services().P2P().ConnectionTracker().Done()
	}()
}

func (h *HTTPServiceDefault) p2pNodesHandler(w http.ResponseWriter, r *http.Request) {
	localId, err := h.Services().P2P().NodeId().ToString()

	ctx := httputil.Context(r, w)

	if ctx.Check("error getting local node id", err) != nil {
		return
	}

	uris := h.Services().P2P().SelfConnectionUris()

	nodeList := make([]P2PNodeResponse, len(uris))

	for i, uri := range uris {
		nodeList[i] = P2PNodeResponse{
			Id:   localId,
			Uris: []string{uri.String()},
		}
	}

	ctx.Encode(P2PNodesResponse{
		Nodes: nodeList,
	})
}
func (h *HTTPServiceDefault) p2pPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers := h.Services().P2P().Peers().Values()
	peerList := make([]P2PNodeResponse, 0)

	ctx := httputil.Context(r, w)

	for _, p := range peers {
		peer, ok := p.(s5net.Peer)
		if !ok {
			continue
		}

		id, err := peer.Id().ToString()
		if err != nil {
			h.Logger().Error("error getting peer id", zap.Error(err))
			continue
		}

		if len(peer.ConnectionURIs()) == 0 {
			continue
		}

		uris := make([]string, len(peer.ConnectionURIs()))

		for i, uri := range peer.ConnectionURIs() {
			uris[i] = uri.String()
		}

		peerList = append(peerList, P2PNodeResponse{
			Id:   id,
			Uris: uris,
		})
	}

	ctx.Encode(P2PNodesResponse{
		Nodes: peerList,
	})
}
