package protocol

import (
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/net"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/vmihailenco/msgpack/v5"
	"net/url"
)

var (
	_ IncomingMessageSigned = (*AnnouncePeers)(nil)
)

type AnnouncePeers struct {
	peer           net.Peer
	connectionUris []*url.URL
	peersToSend    []net.Peer
	HandshakeRequirement
}

func (a *AnnouncePeers) PeersToSend() []net.Peer {
	return a.peersToSend
}

func (a *AnnouncePeers) SetPeersToSend(peersToSend []net.Peer) {
	a.peersToSend = peersToSend
}

func NewAnnounceRequest(peer net.Peer, peersToSend []net.Peer) *AnnouncePeers {
	return &AnnouncePeers{peer: peer, connectionUris: nil, peersToSend: peersToSend}
}

func NewAnnouncePeers() *AnnouncePeers {
	ap := &AnnouncePeers{peer: nil, connectionUris: nil}

	ap.SetRequiresHandshake(true)

	return ap
}

func (a *AnnouncePeers) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageDataSigned) error {
	// CIDFromString the number of peers.
	numPeers, err := dec.DecodeInt()
	if err != nil {
		return err
	}

	// Initialize the slice for storing connection URIs.
	var connectionURIs []*url.URL

	// Loop through each peer.
	for i := 0; i < numPeers; i++ {
		// CIDFromString peer ID.
		peerIdBytes, err := dec.DecodeBytes()
		if err != nil {
			return err
		}
		peerId := encoding.NewNodeId(peerIdBytes)

		// Skip decoding connection status as it is not used.
		_, err = dec.DecodeBool() // Connection status, not used.
		if err != nil {
			return err
		}

		// CIDFromString the number of connection URIs for this peer.
		numUris, err := dec.DecodeInt()
		if err != nil {
			return err
		}

		// CIDFromString each connection URI for this peer.
		for j := 0; j < numUris; j++ {
			uriStr, err := dec.DecodeString()
			if err != nil {
				return err
			}

			uri, err := url.Parse(uriStr)
			if err != nil {
				return err
			}

			pid, err := peerId.ToString()
			if err != nil {
				return err
			}

			passwd, empty := uri.User.Password()
			if empty {
				passwd = ""
			}

			// Incorporate the peer ID into the URI.
			uri.User = url.UserPassword(pid, passwd)

			connectionURIs = append(connectionURIs, uri)
		}
	}

	a.connectionUris = connectionURIs

	return nil
}

func (a AnnouncePeers) HandleMessage(message IncomingMessageDataSigned) error {
	mediator := message.Mediator
	peer := message.Peer
	if len(a.connectionUris) > 0 {
		err := mediator.ConnectToNode([]*url.URL{a.connectionUris[0]}, false, peer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a AnnouncePeers) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeUint(uint64(types.ProtocolMethodAnnouncePeers))
	if err != nil {
		return err
	}

	// Encode the number of peers.
	err = enc.EncodeInt(int64(len(a.peersToSend)))
	if err != nil {
		return err
	}

	// Loop through each peer.
	for _, peer := range a.peersToSend {
		err = enc.EncodeBytes(peer.Id().Raw())
		if err != nil {
			return err
		}

		// Encode connection status.
		err = enc.EncodeBool(peer.IsConnected())
		if err != nil {
			return err
		}

		// Encode the number of connection URIs for this peer.
		err = enc.EncodeInt(int64(len(peer.ConnectionURIs())))
		if err != nil {
			return err
		}

		// Encode each connection URI for this peer.
		for _, uri := range peer.ConnectionURIs() {
			err = enc.EncodeString(uri.String())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
