package encoding

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/LumeWeb/libs5-go/utils"
)

type MultihashCode = int

type Multihash struct {
	fullBytes []byte
}

func (m *Multihash) FullBytes() []byte {
	return m.fullBytes
}

var _ json.Marshaler = (*Multihash)(nil)
var _ json.Unmarshaler = (*Multihash)(nil)

func NewMultihash(fullBytes []byte) *Multihash {
	return &Multihash{fullBytes: fullBytes}
}

func (m *Multihash) FunctionType() types.HashType {
	return types.HashType(m.fullBytes[0])
}

func (m *Multihash) HashBytes() []byte {
	return m.fullBytes[1:]
}

func MultihashFromBytes(bytes []byte, kind types.HashType) *Multihash {
	return NewMultihash(append([]byte{byte(kind)}, bytes...))
}

func MultihashFromBase64Url(hash string) (*Multihash, error) {
	ret, err := base64.RawURLEncoding.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	return NewMultihash(ret), nil
}

func (m *Multihash) ToBase64Url() (string, error) {
	return base64.RawURLEncoding.EncodeToString(m.fullBytes), nil
}

func (m *Multihash) ToBase32() (string, error) {
	return base32.StdEncoding.EncodeToString(m.fullBytes), nil
}

func (m *Multihash) ToString() (string, error) {
	if m.FunctionType() == types.HashType(types.CIDTypeBridge) {
		return string(m.fullBytes), nil // Assumes the bytes are valid UTF-8
	}
	return m.ToBase64Url()
}

func (m *Multihash) Equals(other *Multihash) bool {
	return bytes.Equal(m.fullBytes, other.fullBytes)
}

func (m *Multihash) HashCode() MultihashCode {
	return utils.HashCode(m.fullBytes[:4])
}

func (b *Multihash) UnmarshalJSON(data []byte) error {
	decodedData, err := MultihashFromBase64Url(string(data))
	if err != nil {
		return err
	}

	b.fullBytes = decodedData.fullBytes
	return nil
}
func (b Multihash) MarshalJSON() ([]byte, error) {
	url, err := b.ToBase64Url()
	if err != nil {
		return nil, err
	}

	return []byte(url), nil

}
