package metadata

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
)

var (
	_ msgpack.CustomDecoder = (*MediaMetadataLinks)(nil)
	_ msgpack.CustomEncoder = (*MediaMetadataLinks)(nil)
)

type MediaMetadataLinks struct {
	Count     int
	Head      []*encoding.CID
	Collapsed []*encoding.CID
	Tail      []*encoding.CID
}

func (m MediaMetadataLinks) EncodeMsgpack(enc *msgpack.Encoder) error {
	return errors.New("Not implemented")
}

func (m MediaMetadataLinks) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := decodeIntMap(dec)
	if err != nil {
		return err
	}

	for key, value := range data {
		switch key {
		case 1:
			m.Count = value.(int)
		case 2:
			head := value.([]interface{})
			for _, h := range head {
				cid, err := encoding.CIDFromBytes(h.([]byte))
				if err != nil {
					return err
				}
				m.Head = append(m.Head, cid)
			}
		case 3:
			collapsed := value.([]interface{})
			for _, c := range collapsed {
				cid, err := encoding.CIDFromBytes(c.([]byte))
				if err != nil {
					return err
				}
				m.Collapsed = append(m.Collapsed, cid)
			}
		case 4:
			tail := value.([]interface{})
			for _, t := range tail {
				cid, err := encoding.CIDFromBytes(t.([]byte))
				if err != nil {
					return err
				}
				m.Tail = append(m.Tail, cid)
			}
		}
	}

	return nil
}

func NewMediaMetadataLinks(head []*encoding.CID) *MediaMetadataLinks {
	return &MediaMetadataLinks{
		Count: len(head),
		Head:  head,
	}
}
