package config

import (
	"github.com/LumeWeb/libs5-go/db"
	"github.com/LumeWeb/libs5-go/ed25519"
	"go.uber.org/zap"
)

type NodeConfig struct {
	P2P     P2PConfig `mapstructure:"p2p"`
	KeyPair *ed25519.KeyPairEd25519
	DB      db.KVStore
	Logger  *zap.Logger
	HTTP    HTTPConfig `mapstructure:"http"`
}
type P2PConfig struct {
	Network                 string      `mapstructure:"network"`
	Peers                   PeersConfig `mapstructure:"peers"`
	MaxOutgoingPeerFailures uint        `mapstructure:"max_outgoing_peer_failures"`
	MaxConnectionAttempts   uint        `mapstructure:"max_connection_attempts"`
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
