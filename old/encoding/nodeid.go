package encoding

import (
	"bytes"
	"errors"
	"github.com/multiformats/go-multibase"
	"go.lumeweb.com/libs5-go/internal/bases"
	"go.lumeweb.com/libs5-go/utils"
)

var (
	errorNotBase58BTC = errors.New("not a base58btc string")
)

type NodeIdCode = int

type NodeId struct {
	bytes []byte
}

func (nodeId *NodeId) Bytes() []byte {
	return nodeId.bytes
}

func NewNodeId(bytes []byte) *NodeId {
	return &NodeId{bytes: bytes}
}

func DecodeNodeId(nodeId string) (*NodeId, error) {
	encoding, ret, err := multibase.Decode(nodeId)
	if err != nil {
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

func (nodeId *NodeId) HashCode() int {
	return utils.HashCode(nodeId.bytes[:4])
}

func (nodeId *NodeId) ToBase58() (string, error) {
	return bases.ToBase58BTC(nodeId.bytes)
}

func (nodeId *NodeId) ToString() (string, error) {
	return nodeId.ToBase58()
}

func (nodeId *NodeId) Raw() []byte {
	return nodeId.bytes
}
