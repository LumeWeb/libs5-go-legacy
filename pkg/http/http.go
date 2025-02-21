package http

import (
	"context"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/libs5-go/build"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/p2p"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
)

var _ HTTPService = (*HTTPServiceDefault)(nil)

type HTTPService interface {
	GetHttpRouter() map[string]http.HandlerFunc
	Init(ctx context.Context) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	Logger() *zap.Logger
	Config() *config.NodeConfig
	DB() kv.KVStore
}

type P2PNodesResponse struct {
	Nodes []P2PNodeResponse `json:"nodes"`
}

type P2PNodeResponse struct {
	Id   string   `json:"id"`
	Uris []string `json:"uris"`
}

type HTTPServiceDefault struct {
	logger *zap.Logger
	config *config.NodeConfig
	db     kv.KVStore
	p2p    p2p.P2PService
}

func (h *HTTPServiceDefault) Logger() *zap.Logger {
	return h.logger
}

func (h *HTTPServiceDefault) Config() *config.NodeConfig {
	return h.config
}

func (h *HTTPServiceDefault) DB() kv.KVStore {
	return h.db
}

func NewHTTP(config *config.NodeConfig, logger *zap.Logger, db kv.KVStore, p2p p2p.P2PService) *HTTPServiceDefault {
	return &HTTPServiceDefault{
		logger: logger,
		config: config,
		db:     db,
		p2p:    p2p,
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

func (h *HTTPServiceDefault) Start(_ context.Context) error {
	return nil
}

func (h *HTTPServiceDefault) Stop(_ context.Context) error {
	return nil
}

func (h *HTTPServiceDefault) Init(_ context.Context) error {
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

	peer, err := transport.CreateTransportPeer("wss", &transport.TransportPeerConfig{
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

	h.p2p.ConnectionTracker().Add(1)

	go func() {
		err := h.p2p.OnNewPeer(peer, false)
		if err != nil {
			h.logger.Error("error handling new peer", zap.Error(err))
		}
		h.p2p.ConnectionTracker().Done()
	}()
}

func (h *HTTPServiceDefault) p2pNodesHandler(w http.ResponseWriter, r *http.Request) {
	localId, err := h.p2p.NodeId().ToString()

	ctx := httputil.Context(r, w)

	if ctx.Check("error getting local node id", err) != nil {
		return
	}

	uris := h.p2p.SelfConnectionUris()

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
	peers := h.p2p.Peers().Values()
	peerList := make([]P2PNodeResponse, 0)

	ctx := httputil.Context(r, w)

	for _, p := range peers {
		peer, ok := p.(transport.Peer)
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
