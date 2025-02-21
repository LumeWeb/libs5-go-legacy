package protocol

import (
	"context"
	"crypto/rand"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"io"
	"math"
)

func GenerateChallenge() []byte {
	challenge := make([]byte, 32)
	_, err := rand.Read(challenge)
	if err != nil {
		panic(err)
	}

	return challenge
}

func CalculateNodeScore(goodResponses, badResponses int) float64 {
	totalVotes := goodResponses + badResponses
	if totalVotes == 0 {
		return 0.5
	}

	average := float64(goodResponses) / float64(totalVotes)
	score := average - (average-0.5)*math.Pow(2, -math.Log(float64(totalVotes+1)))

	return score
}

var (
	_ msgpack.CustomDecoder = (*IncomingMessageReader)(nil)
)

type IncomingMessage interface {
	HandleMessage(message IncomingMessageData) error
	DecodeMessage(dec *msgpack.Decoder, message IncomingMessageData) error
	HandshakeRequirer
}
type EncodeableMessage interface {
	msgpack.CustomEncoder
}

type IncomingMessageData struct {
	Original []byte
	Data     []byte
	Ctx      context.Context
	Logger   *zap.Logger
	Peer     net.Peer
	Config   *config.NodeConfig
	VerifyId bool
	Mediator Mediator
}

type IncomingMessageReader struct {
	Kind int
	Data []byte
}

func (i *IncomingMessageReader) DecodeMsgpack(dec *msgpack.Decoder) error {
	kind, err := dec.DecodeInt()
	if err != nil {
		return err
	}

	i.Kind = kind

	raw, err := io.ReadAll(dec.Buffered())

	if err != nil {
		return err
	}

	i.Data = raw

	return nil
}

type HandshakeRequirer interface {
	RequiresHandshake() bool
	SetRequiresHandshake(value bool)
}

type HandshakeRequirement struct {
	requiresHandshake bool
}

func (hr *HandshakeRequirement) RequiresHandshake() bool {
	return hr.requiresHandshake
}

func (hr *HandshakeRequirement) SetRequiresHandshake(value bool) {
	hr.requiresHandshake = value
}
