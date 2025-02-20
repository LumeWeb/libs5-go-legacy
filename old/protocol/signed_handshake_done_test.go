package protocol

import (
	"bytes"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/types"
	"net/url"
	"testing"
)

func TestHandshakeDone_EncodeMsgpack(t *testing.T) {
	type fields struct {
		supportedFeatures int
		connectionUris    []*url.URL
		handshake         []byte
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantErrFunc assert.ErrorAssertionFunc
	}{
		{
			name: "Empty Fields",
			fields: fields{
				supportedFeatures: 0,
				connectionUris:    []*url.URL{},
				handshake:         []byte{},
			},
			wantErr: false,
		},
		{
			name: "Valid Fields",
			fields: fields{
				supportedFeatures: 1,
				connectionUris:    []*url.URL{{ /* initialize with valid URL data */ }},
				handshake:         []byte{0x01, 0x02},
			},
			wantErr: false,
		},
		{
			name: "Invalid URL",
			fields: fields{
				supportedFeatures: 1,
				connectionUris:    []*url.URL{&url.URL{Host: "invalid-url"}},
				handshake:         []byte{0x01},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := HandshakeDone{
				supportedFeatures: tt.fields.supportedFeatures,
				connectionUris:    tt.fields.connectionUris,
				handshake:         tt.fields.handshake,
			}

			if tt.wantErr {
				tt.wantErrFunc = assert.Error
			} else {
				tt.wantErrFunc = assert.NoError
			}

			enc := msgpack.NewEncoder(new(bytes.Buffer))
			err := m.EncodeMsgpack(enc)

			assert.NoError(t, err)

			encodedData := enc.Writer().(*bytes.Buffer).Bytes()

			if len(encodedData) == 0 && tt.wantErr {
				t.Errorf("Expected non-empty encoded data, got empty")
			}

			dec := msgpack.NewDecoder(bytes.NewReader(encodedData))
			protocol, err := dec.DecodeUint()

			assert.EqualValues(t, protocol, types.ProtocolMethodHandshakeDone)

			handshake, err := dec.DecodeBytes()

			assert.EqualValues(t, handshake, tt.fields.handshake)

			supportedFeatures, err := dec.DecodeInt()

			assert.EqualValues(t, supportedFeatures, tt.fields.supportedFeatures)
		})
	}
}

func TestHandshakeDone_DecodeMessage_Success(t *testing.T) {
	data := "xBFleGFtcGxlX2NoYWxsZW5nZQM="

	h := HandshakeDone{}

	dataDec, err := base64.StdEncoding.DecodeString(data)
	assert.NoError(t, err)

	enc := msgpack.NewDecoder(bytes.NewReader(dataDec))
	err = h.DecodeMessage(enc)
	assert.NoError(t, err)

	assert.EqualValues(t, types.SupportedFeatures, h.supportedFeatures)
	assert.EqualValues(t, []byte("example_challenge"), h.challenge)
}
func TestHandshakeDone_DecodeMessage_InvalidFeatures(t *testing.T) {
	data := "xBFleGFtcGxlX2NoYWxsZW5nZSo="

	h := HandshakeDone{}

	dataDec, err := base64.StdEncoding.DecodeString(data)
	assert.NoError(t, err)

	enc := msgpack.NewDecoder(bytes.NewReader(dataDec))
	err = h.DecodeMessage(enc)
	assert.NotEqualValues(t, types.SupportedFeatures, h.supportedFeatures)
	assert.EqualValues(t, []byte("example_challenge"), h.challenge)
}
func TestHandshakeDone_DecodeMessage_BadChallenge(t *testing.T) {
	data := "xA1iYWRfY2hhbGxlbmdlAw=="

	h := HandshakeDone{}

	dataDec, err := base64.StdEncoding.DecodeString(data)
	assert.NoError(t, err)

	enc := msgpack.NewDecoder(bytes.NewReader(dataDec))
	err = h.DecodeMessage(enc)
	assert.EqualValues(t, types.SupportedFeatures, h.supportedFeatures)
	assert.NotEqualValues(t, []byte("example_challenge"), h.challenge)
}
