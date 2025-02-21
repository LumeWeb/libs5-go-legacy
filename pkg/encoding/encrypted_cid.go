package encoding

import (
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
)

type EncryptedCID struct {
	Multibase
	encryptedBlobHash   Multihash
	OriginalCID         CID
	encryptionAlgorithm byte
	padding             uint64
	chunkSizeAsPowerOf2 int
	encryptionKey       []byte
}

var _ msgpack.CustomEncoder = (*EncryptedCID)(nil)
var _ msgpack.CustomDecoder = (*EncryptedCID)(nil)
var _ json.Marshaler = (*EncryptedCID)(nil)
var _ json.Unmarshaler = (*EncryptedCID)(nil)

func NewEncryptedCID(encryptedBlobHash Multihash, originalCID CID, encryptionKey []byte, padding uint64, chunkSizeAsPowerOf2 int, encryptionAlgorithm byte) *EncryptedCID {
	e := &EncryptedCID{
		encryptedBlobHash:   encryptedBlobHash,
		OriginalCID:         originalCID,
		encryptionKey:       encryptionKey,
		padding:             padding,
		chunkSizeAsPowerOf2: chunkSizeAsPowerOf2,
		encryptionAlgorithm: encryptionAlgorithm,
	}

	m := NewMultibase(e)
	e.Multibase = m

	return e
}

func DecodeEncryptedCID(cid string) (*EncryptedCID, error) {
	data, err := MultibaseDecodeString(cid)
	if err != nil {
		return nil, err
	}
	return EncryptedCIDFromBytes(data)
}

func EncryptedCIDFromBytes(data []byte) (*EncryptedCID, error) {
	if CIDType(data[0]) != CIDTypeEncryptedStatic {
		return nil, errors.New("Invalid CID type")
	}

	cid, err := CIDFromBytes(data[72:])
	if err != nil {
		return nil, err
	}

	encryptedBlobHash := NewMultihash(data[3:36])
	encryptionKey := data[36:68]
	padding := DecodeEndian(data[68:72])
	chunkSizeAsPowerOf2 := int(data[2])
	encryptionAlgorithm := data[1]

	return NewEncryptedCID(*encryptedBlobHash, *cid, encryptionKey, padding, chunkSizeAsPowerOf2, encryptionAlgorithm), nil
}

func (c *EncryptedCID) ChunkSize() int {
	return 1 << uint(c.chunkSizeAsPowerOf2)
}

func (c *EncryptedCID) ToBytes() []byte {
	data := []byte{
		byte(CIDTypeEncryptedStatic),
		c.encryptionAlgorithm,
		byte(c.chunkSizeAsPowerOf2),
	}
	data = append(data, c.encryptedBlobHash.fullBytes...)
	data = append(data, c.encryptionKey...)
	data = append(data, EncodeEndian(c.padding, 4)...)
	data = append(data, c.OriginalCID.ToBytes()...)
	return data
}
func (c EncryptedCID) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(c.ToBytes())
}

func (c *EncryptedCID) DecodeMsgpack(dec *msgpack.Decoder) error {
	return decodeMsgpackCID(c, dec)
}
func (c EncryptedCID) MarshalJSON() ([]byte, error) {
	str, err := c.ToString()

	if err != nil {
		return nil, err
	}
	// Delegate to the MarshalJSON method of the Encoder
	return json.Marshal(str)
}

func (c *EncryptedCID) UnmarshalJSON(data []byte) error {
	decData, err := UnmarshalBase64UrlJSON(data)

	if err != nil {
		return err
	}

	decodedCid, err := EncryptedCIDFromBytes(decData)

	if err != nil {
		return err
	}

	*c = *decodedCid
	return nil
}
