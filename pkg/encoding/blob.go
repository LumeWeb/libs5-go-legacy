package encoding

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"go.lumeweb.com/libs5-go/pkg/crypto"
)

type BlobCode = int

type Blob struct {
	fullBytes []byte
}

func (m *Blob) FullBytes() []byte {
	result := make([]byte, len(m.fullBytes))
	copy(result, m.fullBytes)
	return result
}

var _ json.Marshaler = (*Blob)(nil)
var _ json.Unmarshaler = (*Blob)(nil)

func NewMultihash(fullBytes []byte) *Blob {
	return &Blob{fullBytes: fullBytes}
}

func (m *Blob) FunctionType() crypto.HashType {
	return crypto.HashType(m.fullBytes[0])
}

func (m *Blob) HashBytes() []byte {
	return m.fullBytes[1:]
}

func MultihashFromBytes(bytes []byte, kind crypto.HashType) *Blob {
	return NewMultihash(append([]byte{byte(kind)}, bytes...))
}

func MultihashFromBase64Url(hash string) (*Blob, error) {
	ret, err := base64.RawURLEncoding.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	return NewMultihash(ret), nil
}

func (m *Blob) ToBase64Url() (string, error) {
	return base64.RawURLEncoding.EncodeToString(m.fullBytes), nil
}

func (m *Blob) ToBase32() (string, error) {
	return base32.StdEncoding.EncodeToString(m.fullBytes), nil
}

func (m *Blob) ToString() (string, error) {
	if m.FunctionType() == crypto.HashType(CIDTypeBridge) {
		return string(m.HashBytes()), nil
	}
	return m.ToBase64Url()
}

func (m *Blob) Equals(other *Blob) bool {
	return bytes.Equal(m.fullBytes, other.fullBytes)
}

func (b *Blob) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	decodedBlob, err := MultihashFromBase64Url(str)
	if err != nil {
		return err
	}

	*b = *decodedBlob
	return nil
}
func (b Blob) MarshalJSON() ([]byte, error) {
	url, err := b.ToBase64Url()
	if err != nil {
		return nil, err
	}

	return json.Marshal(url)
}
