package protocol

import (
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.lumeweb.com/libs5-go/pkg/storage/location"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"log"
	"old/net"
	"old/types"
)

var _ EncodeableMessage = (*HashQuery)(nil)
var _ IncomingMessage = (*HashQuery)(nil)

type HashQuery struct {
	hash                 *encoding.Blob                 `json:"hash,omitempty"`
	kinds                []location.StorageLocationType `json:"kinds,omitempty"`
	HandshakeRequirement `json:"handshakeRequirement"`
}

func NewHashQuery() *HashQuery {
	hq := &HashQuery{}

	hq.SetRequiresHandshake(true)

	return hq
}

func NewHashRequest(hash *encoding.Blob, kinds []location.StorageLocationType) *HashQuery {
	if len(kinds) == 0 {
		kinds = []location.StorageLocationType{location.StorageLocationTypeFile}
	}
	return &HashQuery{
		hash:  hash,
		kinds: kinds,
	}
}

func (h HashQuery) Hash() *encoding.Blob {
	return h.hash
}

func (h HashQuery) Kinds() []location.StorageLocationType {
	return h.kinds
}

func (h *HashQuery) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error {
	hash, err := dec.DecodeBytes()

	if err != nil {
		return err
	}

	h.hash = encoding.NewMultihash(hash)

	var kinds []location.StorageLocationType
	err = dec.Decode(&kinds)
	if err != nil {
		return err
	}

	h.kinds = kinds

	return nil
}

func (h HashQuery) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeInt(int64(types.ProtocolMethodHashQuery))
	if err != nil {
		return err
	}

	err = enc.EncodeBytes(h.hash.FullBytes())

	if err != nil {
		return err
	}

	err = enc.Encode(h.kinds)

	if err != nil {
		return err
	}

	return nil
}

func (h *HashQuery) HandleMessage(message IncomingMessageData) error {
	peer := message.Peer
	mediator := message.Mediator
	logger := message.Logger
	config := message.Config

	mapLocations, err := mediator.GetCachedStorageLocations(h.hash, h.kinds)
	if err != nil {
		log.Printf("Error getting cached storage locations: %v", err)
		return err
	}

	if len(mapLocations) > 0 {
		availableNodes := make([]*encoding.NodeId, 0, len(mapLocations))
		for key := range mapLocations {
			nodeId, err := encoding.DecodeNodeId(key)
			if err != nil {
				logger.Error("Error decoding node id", zap.Error(err))
				continue
			}

			availableNodes = append(availableNodes, nodeId)
		}

		score, err := mediator.SortNodesByScore(availableNodes)
		if err != nil {
			return err
		}

		sortedNodeId, err := (*score[0]).ToString()
		if err != nil {
			return err
		}

		entry, exists := mapLocations[sortedNodeId]
		if exists {
			err := peer.SendMessage(entry.ProviderMessage())
			if err != nil {
				return err
			}
		}
	}

	if mediator.ProviderStore() != nil {
		if mediator.ProviderStore().CanProvide(h.hash, h.kinds) {
			location, err := mediator.ProviderStore().Provide(h.hash, h.kinds)
			if err != nil {
				return err
			}

			message := storage.PrepareProvideMessage(config.KeyPair, h.hash, location)

			err = peer.SendMessage(message)
			if err != nil {
				return err
			}
		}
	}

	var peers *structs.SetImpl
	hashString, err := h.hash.ToString()
	logger.Debug("HashQuery", zap.Any("hashString", hashString))
	if err != nil {
		return err
	}
	peersVal, ok := mediator.HashQueryRoutingTable().Get(hashString)
	if ok {
		peers = peersVal.(*structs.SetImpl)
		if !peers.Contains(peer.Id()) {
			peers.Add(peer.Id())
		}

		return nil
	}

	peerList := structs.NewSet()
	peerList.Add(peer.Id())

	mediator.HashQueryRoutingTable().Put(hashString, peerList)

	for _, val := range mediator.Peers().Values() {
		peerVal := val.(transport.Peer)
		if !peerVal.Id().Equals(peer.Id()) {
			err := peerVal.SendMessage(message.Original)
			if err != nil {
				logger.Error("Failed to send message", zap.Error(err))
			}
		}
	}

	return nil
}
