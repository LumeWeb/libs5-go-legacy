package config

import (
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.uber.org/zap"
)

type NodeConfig struct {
	P2P     *P2PConfig `mapstructure:"p2p"`
	KeyPair *crypto.KeyPairEd25519
	DB      kv.KVStore
	Logger  *zap.Logger
	HTTP    HTTPConfig `mapstructure:"http"`
}
type P2PConfig struct {
	Network                 string       `mapstructure:"network"`
	Peers                   *PeersConfig `mapstructure:"peers"`
	MaxOutgoingPeerFailures uint         `mapstructure:"max_outgoing_peer_failures"`
	MaxConnectionAttempts   uint         `mapstructure:"max_connection_attempts"`
}

type PeersConfig struct {
	Initial   []string `mapstructure:"initial"`
	Blocklist []string `mapstructure:"blocklist"`
}

type HTTPAPIConfig struct {
	Domain string `mapstructure:"domain"`
	Port   uint   `mapstructure:"port"`
}

type HTTPConfig struct {
	API HTTPAPIConfig `mapstructure:"api"`
}
