package service

import (
	"github.com/LumeWeb/libs5-go/net"
	"github.com/LumeWeb/libs5-go/protocol"
)

type RegistryService interface {
	Set(sre protocol.SignedRegistryEntry, trusted bool, receivedFrom net.Peer) error
	BroadcastEntry(sre protocol.SignedRegistryEntry, receivedFrom net.Peer) error
	SendRegistryRequest(pk []byte) error
	Get(pk []byte) (protocol.SignedRegistryEntry, error)
	Listen(pk []byte, cb func(sre protocol.SignedRegistryEntry)) (func(), error)
	Service
}
