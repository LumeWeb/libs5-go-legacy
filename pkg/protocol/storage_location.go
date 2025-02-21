package protocol

import (
	"crypto/ed25519"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/net"
	"go.lumeweb.com/libs5-go/storage"
	"go.lumeweb.com/libs5-go/structs"
	"go.lumeweb.com/libs5-go/types"
	"go.lumeweb.com/libs5-go/utils"
	"go.uber.org/zap"
)

var _ IncomingMessage = (*StorageLocation)(nil)

type StorageLocation struct {
	hash      *encoding.Multihash
	kind      int
	expiry    int64
	parts     []string
	publicKey []byte
	signature []byte
	HandshakeRequirement
}

func NewStorageLocation() *StorageLocation {
	sl := &StorageLocation{}

	sl.SetRequiresHandshake(true)

	return sl
}

func (s *StorageLocation) DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error {
	// nop, we use the incoming message -> original already stored
	return nil
}
func (s *StorageLocation) HandleMessage(message IncomingMessageData) error {
	msg := message.Original
	mediator := message.Mediator
	peer := message.Peer
	logger := message.Logger

	hash := encoding.NewMultihash(msg[1:34]) // Replace NewMultihash with appropriate function

	typeOfData := msg[34]

	expiry := utils.DecodeEndian(msg[35:39])

	partCount := msg[39]

	parts := []string{}
	cursor := 40
	for i := 0; i < int(partCount); i++ {
		length := utils.DecodeEndian(msg[cursor : cursor+2])
		cursor += 2
		if len(msg) < cursor+int(length) {
			return fmt.Errorf("Invalid message")
		}
		part := string(msg[cursor : cursor+int(length)])
		parts = append(parts, part)
		cursor += int(length)
	}

	cursor++

	publicKey := msg[cursor : cursor+33]
	signature := msg[cursor+33:]

	if types.HashType(publicKey[0]) != types.HashTypeEd25519 { // Replace CID_HASH_TYPES_ED25519 with actual constant
		return fmt.Errorf("Unsupported public key type %d", publicKey[0])
	}

	if !ed25519.Verify(publicKey[1:], msg[:cursor], signature) {
		return fmt.Errorf("Signature verification failed")
	}

	nodeId := encoding.NewNodeId(publicKey)

	err := mediator.AddStorageLocation(hash, nodeId, storage.NewStorageLocation(int(typeOfData), parts, int64(expiry)), msg)
	if err != nil {
		return fmt.Errorf("Failed to add storage location: %s", err)
	}

	hashStr, err := hash.ToString()
	if err != nil {
		return err
	}

	var list *structs.SetImpl
	listVal, ok := mediator.HashQueryRoutingTable().Get(hashStr) // Implement HashQueryRoutingTable method
	if !ok {
		list = structs.NewSet()
	} else {
		list = listVal.(*structs.SetImpl)
	}

	for _, peerIdVal := range list.Values() {
		peerId := peerIdVal.(*encoding.NodeId)

		if peerId.Equals(nodeId) || peerId.Equals(peer) {
			continue
		}
		peerIdStr, err := peerId.ToString()
		if err != nil {
			return err
		}

		if peerVal, ok := mediator.Peers().Get(peerIdStr); ok {
			foundPeer := peerVal.(net.Peer)
			err := foundPeer.SendMessage(msg)
			if err != nil {
				logger.Error("Failed to send message", zap.Error(err))
				continue
			}
		}

		mediator.HashQueryRoutingTable().Remove(hashStr)
	}

	return nil
}
