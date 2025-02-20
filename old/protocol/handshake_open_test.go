package protocol

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/types"
	"testing"
)

func TestHandshakeOpen_EncodeMsgpack(t *testing.T) {
	type fields struct {
		challenge []byte
		networkId string
	}
	type args struct {
		enc *msgpack.Encoder
		buf *bytes.Buffer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Normal Case",
			fields: fields{
				challenge: []byte("test-challenge"),
				networkId: "test-network",
			},
			args: args{
				buf: new(bytes.Buffer),
			},
			wantErr: false,
		},
		{
			name: "Empty Network ID",
			fields: fields{
				challenge: []byte("test-challenge"),
				networkId: "",
			},
			args: args{
				buf: new(bytes.Buffer),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt.args.enc = msgpack.NewEncoder(tt.args.buf)
		t.Run(tt.name, func(t *testing.T) {
			h := HandshakeOpen{
				challenge: tt.fields.challenge,
				networkId: tt.fields.networkId,
			}
			if err := h.EncodeMsgpack(tt.args.enc); (err != nil) != tt.wantErr {
				t.Errorf("EncodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check the contents of the buffer to verify encoding
			encodedData := tt.args.buf.Bytes()
			if len(encodedData) == 0 && !tt.wantErr {
				t.Errorf("Expected non-empty encoded data, got empty")
			}

			dec := msgpack.NewDecoder(bytes.NewReader(encodedData))

			protocolMethod, err := dec.DecodeUint()
			if err != nil {
				t.Errorf("DecodeUint() error = %v", err)
			}

			assert.EqualValues(t, types.ProtocolMethodHandshakeOpen, protocolMethod)

			challenge, err := dec.DecodeBytes()
			if err != nil {
				t.Errorf("DecodeBytes() error = %v", err)
			}

			assert.EqualValues(t, tt.fields.challenge, challenge)

			networkId, err := dec.DecodeString()
			if err != nil {
				if err.Error() == "EOF" && tt.fields.networkId != "" {
					t.Logf("DecodeString() networkId missing, got EOF")
				}
				if err.Error() != "EOF" {
					t.Errorf("DecodeString() error = %v", err)
				}
			}

			assert.EqualValues(t, tt.fields.networkId, networkId)
		})
	}
}
func TestHandshakeOpen_DecodeMessage(t *testing.T) {
	type fields struct {
		challenge []byte
		networkId string
		handshake []byte
	}
	type args struct {
		base64EncodedData string // Base64 encoded msgpack data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Valid Handshake and NetworkID",
			fields: fields{}, // Fields are now empty
			args: args{
				base64EncodedData: "xBNzYW1wbGVIYW5kc2hha2VEYXRhr3NhbXBsZU5ldHdvcmtJRA==",
			},
			wantErr: assert.NoError,
		},
		{
			name:   "Valid Handshake and Empty NetworkID",
			fields: fields{}, // Fields are now empty
			args: args{
				base64EncodedData: "xBNzYW1wbGVIYW5kc2hha2VEYXRh",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HandshakeOpen{}

			decodedData, _ := base64.StdEncoding.DecodeString(tt.args.base64EncodedData)
			reader := bytes.NewReader(decodedData)
			dec := msgpack.NewDecoder(reader)

			tt.wantErr(t, h.DecodeMessage(dec), fmt.Sprintf("DecodeMessage(%v)", tt.args.base64EncodedData))
		})
	}
}

func TestHandshakeOpen_HandleMessage_Success(t *testing.T) {
	setup(t)
	testResultEncoded := "AsQWZXhhbXBsZSBoYW5kc2hha2UgZGF0YQMA"
	testResult, err := base64.StdEncoding.DecodeString(testResultEncoded)
	assert.NoError(t, err)

	node.EXPECT().Services().Return(services).Times(1)
	node.EXPECT().NetworkId().Return("").Times(1)
	services.EXPECT().P2P().Return(p2p).Times(1)
	p2p.EXPECT().SignMessageSimple(testResult).Return(testResult, nil).Times(1)
	peer.EXPECT().SendMessage(testResult).Return(nil).Times(1)

	handshake := []byte("example handshake data")
	handshakeOpen := NewHandshakeOpen([]byte{}, "")
	handshakeOpen.SetHandshake(handshake)

	assert.NoError(t, handshakeOpen.HandleMessage(node, peer, false))
}
func TestHandshakeOpen_HandleMessage_DifferentNetworkID(t *testing.T) {
	setup(t) // Assuming setup initializes the mocks and any required objects

	// Define a network ID that is different from the one in handshakeOpen
	differentNetworkID := "differentNetworkID"

	// Setup expectations for the mock objects
	node.EXPECT().NetworkId().Return(differentNetworkID).Times(1)
	// No other method calls are expected after the NetworkId check fails

	// Create a HandshakeOpen instance with a specific network ID that differs from `differentNetworkID`
	networkIDForHandshakeOpen := "expectedNetworkID"
	handshakeOpen := NewHandshakeOpen([]byte{}, networkIDForHandshakeOpen)
	handshakeOpen.SetHandshake([]byte("example handshake data"))

	// Invoke HandleMessage and expect an error
	err := handshakeOpen.HandleMessage(node, peer, false)
	assert.Error(t, err)

	// Optionally, assert that the error message is as expected
	expectedErrorMessage := fmt.Sprintf("Peer is in different network: %s", networkIDForHandshakeOpen)
	assert.Equal(t, expectedErrorMessage, err.Error())
}

func TestHandshakeOpen_HandleMessage_MarshalError(t *testing.T) {
	setup(t)

	node.EXPECT().Services().Return(services).Times(1)
	node.EXPECT().NetworkId().Return("").Times(1)
	services.EXPECT().P2P().Return(p2p).Times(1)
	p2p.EXPECT().SignMessageSimple(gomock.Any()).Return(nil, fmt.Errorf("marshal error")).Times(1)

	handshake := []byte("example handshake data")
	handshakeOpen := NewHandshakeOpen([]byte{}, "")
	handshakeOpen.SetHandshake(handshake)

	assert.Error(t, handshakeOpen.HandleMessage(node, peer, false))
}

func TestHandshakeOpen_HandleMessage_SendMessageError(t *testing.T) {
	setup(t)

	node.EXPECT().Services().Return(services).Times(1)
	node.EXPECT().NetworkId().Return("").Times(1)
	services.EXPECT().P2P().Return(p2p).Times(1)
	p2p.EXPECT().SignMessageSimple(gomock.Any()).Return([]byte{}, nil).Times(1)
	peer.EXPECT().SendMessage(gomock.Any()).Return(fmt.Errorf("send message error")).Times(1)

	handshake := []byte("example handshake data")
	handshakeOpen := NewHandshakeOpen([]byte{}, "")
	handshakeOpen.SetHandshake(handshake)

	assert.Error(t, handshakeOpen.HandleMessage(node, peer, false))
}
