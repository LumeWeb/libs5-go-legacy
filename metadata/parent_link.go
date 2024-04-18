package metadata

import (
	"errors"
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	_ msgpack.CustomDecoder = (*MetadataParentLink)(nil)
	_ msgpack.CustomEncoder = (*MetadataParentLink)(nil)
)

// MetadataParentLink represents the structure for Metadata Parent Link.
type MetadataParentLink struct {
	CID    *encoding.CID
	Type   types.ParentLinkType
	Role   string
	Signed bool
	parent *MediaMetadata
}

func (m *MetadataParentLink) SetParent(parent *MediaMetadata) {
	m.parent = parent
}

func (m *MetadataParentLink) EncodeMsgpack(enc *msgpack.Encoder) error {
	return errors.New("Not implemented")
}

func (m *MetadataParentLink) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return err
	}

	cid := &encoding.CID{}

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt8()
		if err != nil {
			return err
		}
		value, err := dec.DecodeInterface()
		if err != nil {
			return err
		}

		switch key {
		case 0:
			m.Type = types.ParentLinkType(value.(int))
		case 1:
			cid, err = encoding.CIDFromBytes(value.([]byte))
			if err != nil {
				return err
			}

			m.CID = cid
		}
	}

	if m.Type == 0 {
		m.Type = types.ParentLinkTypeUserIdentity
	}

	m.Signed = false

	if m.parent != nil {
		for _, key := range m.parent.ProvenPubKeys() {
			if cid.Hash.Equals(key) {
				m.Signed = true
				break
			}
		}
	}

	return nil
}

// NewMetadataParentLink creates a new MetadataParentLink with the provided values.
func NewMetadataParentLink(cid *encoding.CID, role string, signed bool) *MetadataParentLink {
	return &MetadataParentLink{
		CID:    cid,
		Type:   types.ParentLinkTypeUserIdentity,
		Role:   role,
		Signed: signed,
	}
}
