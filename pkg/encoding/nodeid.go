package encoding

import (
	"bytes"
	"errors"
	"github.com/multiformats/go-multibase"
	libbytes "go.lumeweb.com/libs5-go/pkg/bytes"
	"go.lumeweb.com/libs5-go/pkg/internal"
)

var (
	errorNotBase58BTC = errors.New("not a base58btc string")
)

type NodeIdCode = int

type NodeId struct {
	bytes []byte
}

func (nodeId *NodeId) Bytes() []byte {
	result := make([]byte, len(nodeId.bytes))
	copy(result, nodeId.bytes)
	return result
}

func NewNodeId(bytes []byte) *NodeId {
	// Make a defensive copy of the input bytes
	bytesCopy := make([]byte, len(bytes))
	copy(bytesCopy, bytes)
	return &NodeId{bytes: bytesCopy}
}

func DecodeNodeId(nodeId string) (*NodeId, error) {
	// Special case for empty string to match expected error
	if nodeId == "" {
		return nil, multibase.ErrUnsupportedEncoding
	}

	encoding, ret, err := multibase.Decode(nodeId)
	if err != nil {
		// If error is due to empty string, return consistent error
		if err.Error() == "cannot decode multibase for zero length string" {
			return nil, multibase.ErrUnsupportedEncoding
		}
		return nil, err
	}

	if encoding != multibase.Base58BTC {
		return nil, errorNotBase58BTC
	}

	return NewNodeId(ret), nil
}

func (nodeId *NodeId) Equals(other interface{}) bool {
	if otherNodeId, ok := other.(*NodeId); ok {
		return bytes.Equal(nodeId.bytes, otherNodeId.bytes)
	}
	return false
}

func (nodeId *NodeId) ToBase58() (string, error) {
	return bases.ToBase58BTC(nodeId.bytes)
}

func (nodeId *NodeId) ToString() (string, error) {
	return nodeId.ToBase58()
}

func (nodeId *NodeId) Raw() []byte {
	// Create a new slice and copy the bytes
	result := make([]byte, len(nodeId.bytes))
	copy(result, nodeId.bytes)
	return result
}
func (nodeId *NodeId) HashCode() int {
	return libbytes.HashCode(nodeId.bytes[:4])
}
