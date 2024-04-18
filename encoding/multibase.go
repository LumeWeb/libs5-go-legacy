package encoding

import (
	"errors"
	"github.com/LumeWeb/libs5-go/internal/bases"
	"github.com/multiformats/go-multibase"
)

var (
	ErrMultibaseEncodingNotSupported = errors.New("multibase encoding not supported")
	errMultibaseDecodeZeroLength     = errors.New("cannot decode multibase for zero length string")
)

type Encoder interface {
	ToBytes() []byte
}

type multibaseImpl struct {
	Multibase
	Encoder Encoder
}

type Multibase interface {
	ToHex() (string, error)
	ToBase32() (string, error)
	ToBase64Url() (string, error)
	ToBase58() (string, error)
	ToString() (string, error)
}

var _ Multibase = (*multibaseImpl)(nil)

func NewMultibase(encoder Encoder) Multibase {
	return &multibaseImpl{Encoder: encoder}
}

func MultibaseDecodeString(data string) (bytes []byte, err error) {
	if len(data) == 0 {
		return nil, errMultibaseDecodeZeroLength
	}

	switch data[0] {
	case 'z', 'f', 'u', 'b':
		_, bytes, err = multibase.Decode(data)
	case ':':
		bytes = []byte(data)
	default:
		err = ErrMultibaseEncodingNotSupported
	}

	return bytes, err
}

func (m *multibaseImpl) ToHex() (string, error) {
	return bases.ToHex(m.Encoder.ToBytes())
}

func (m *multibaseImpl) ToBase32() (string, error) {
	return bases.ToBase32(m.Encoder.ToBytes())
}

func (m *multibaseImpl) ToBase64Url() (string, error) {
	return bases.ToBase64Url(m.Encoder.ToBytes())
}

func (m *multibaseImpl) ToBase58() (string, error) {
	return bases.ToBase58BTC(m.Encoder.ToBytes())
}

func (m *multibaseImpl) ToString() (string, error) {
	return m.ToBase58()
}
