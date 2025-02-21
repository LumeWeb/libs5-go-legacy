package encoding

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	libbytes "go.lumeweb.com/libs5-go/pkg/bytes"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"go.lumeweb.com/libs5-go/pkg/internal"
)

var (
	errEmptyBytes       = errors.New("empty libbytes")
	errInvalidInputType = errors.New("invalid input type for libbytes")
)

type CID struct {
	Multibase
	Type CIDType
	Hash Multihash
	Size uint64
}

var _ json.Marshaler = (*CID)(nil)
var _ json.Unmarshaler = (*CID)(nil)
var _ msgpack.CustomEncoder = (*CID)(nil)
var _ msgpack.CustomDecoder = (*CID)(nil)

func NewCID(Type CIDType, Hash Multihash, Size uint64) *CID {
	c := &CID{
		Type: Type,
		Hash: Hash,
		Size: Size,
	}
	m := NewMultibase(c)
	c.Multibase = m

	return c
}

func (cid *CID) getPrefixBytes() []byte {
	return []byte{byte(cid.Type)}
}

func (cid *CID) ToBytes() []byte {
	if cid.Type == CIDTypeBridge {
		return cid.Hash.fullBytes
	} else if cid.Type == CIDTypeRaw {
		sizeBytes := EncodeEndian(cid.Size, 8)

		for len(sizeBytes) > 0 && sizeBytes[len(sizeBytes)-1] == 0 {
			sizeBytes = sizeBytes[:len(sizeBytes)-1]
		}
		if len(sizeBytes) == 0 {
			sizeBytes = []byte{0}
		}

		return libbytes.ConcatBytes(cid.getPrefixBytes(), cid.Hash.fullBytes, sizeBytes)
	}

	return libbytes.ConcatBytes(cid.getPrefixBytes(), cid.Hash.fullBytes)
}

func CIDFromString(cid string) (*CID, error) {
	decodedBytes, err := MultibaseDecodeString(cid)
	if err != nil {
		return nil, err
	}

	cidInstance, err := initCID(decodedBytes)

	if err != nil {
		return nil, err
	}

	return cidInstance, nil
}

func CIDFromRegistry(bytes []byte) (*CID, error) {
	if len(bytes) == 0 {
		return nil, errEmptyBytes
	}

	registryType := RegistryType(bytes[0])
	if _, exists := RegistryTypeMap[registryType]; !exists {
		return nil, fmt.Errorf("invalid registry type %d", bytes[0])
	}

	bytes = bytes[1:]

	cidInstance, err := initCID(bytes)

	if err != nil {
		return nil, err
	}

	return cidInstance, nil
}

func CIDFromBytes(bytes []byte) (*CID, error) {
	return initCID(bytes)
}

func CIDFromHash(bytes interface{}, size uint64, cidType CIDType, hashType crypto.HashType) (*CID, error) {
	var (
		byteSlice []byte
		err       error
	)

	switch v := bytes.(type) {
	case string:
		byteSlice, err = hex.DecodeString(v)
		if err != nil {
			return nil, err
		}
	case []byte:
		byteSlice = v
	default:
		return nil, errInvalidInputType
	}

	if _, exists := CIDTypeMap[cidType]; !exists {
		return nil, fmt.Errorf("invalid hash type %d", cidType)
	}

	return NewCID(cidType, *MultihashFromBytes(byteSlice, hashType), size), nil
}

func CIDVerify(bytes interface{}) bool {
	var (
		byteSlice []byte
		err       error
	)

	switch v := bytes.(type) {
	case string:
		byteSlice, err = MultibaseDecodeString(v) // Assuming MultibaseDecodeString function is defined
		if err != nil {
			return false
		}
	case []byte:
		byteSlice = v
	default:
		return false
	}

	_, err = initCID(byteSlice)
	return err == nil
}

func (cid *CID) CopyWith(newType int, newSize uint64) (*CID, error) {
	if newType == 0 {
		newType = int(cid.Type)
	}

	if _, exists := CIDTypeMap[CIDType(newType)]; !exists {
		return nil, fmt.Errorf("invalid cid type %d", newType)
	}

	return NewCID(CIDType(newType), cid.Hash, newSize), nil
}

func (cid *CID) ToRegistryEntry() []byte {
	registryType := RegistryTypeCID
	cidBytes := cid.ToBytes()
	return libbytes.ConcatBytes([]byte{byte(registryType)}, cidBytes)
}

func (cid *CID) ToRegistryCID() ([]byte, error) {
	registryCIDType := CIDTypeResolver
	copiedCID, err := cid.CopyWith(int(registryCIDType), cid.Size)
	if err != nil {
		return nil, err
	}
	return copiedCID.ToBytes(), nil
}

func (cid *CID) ToString() (string, error) {
	if cid.Type == CIDTypeBridge {
		return cid.Hash.ToString()
	}

	return bases.ToBase58BTC(cid.ToBytes())
}

func (cid *CID) Equals(other *CID) bool {
	return bytes.Equal(cid.ToBytes(), other.ToBytes())
}

func (cid *CID) HashCode() int {
	fullBytes := cid.ToBytes()
	if len(fullBytes) < 4 {
		return 0
	}

	return int(fullBytes[0]) +
		int(fullBytes[1])<<8 +
		int(fullBytes[2])<<16 +
		int(fullBytes[3])<<24
}

func (b CID) MarshalJSON() ([]byte, error) {
	url, err := b.ToBase64Url()
	if err != nil {
		return nil, err
	}

	return json.Marshal(url)
}

func (cid *CID) UnmarshalJSON(data []byte) error {
	decData, err := UnmarshalBase64UrlJSON(data)

	if err != nil {
		return err
	}

	decodedCid, err := CIDFromBytes(decData)

	if err != nil {
		return err
	}

	*cid = *decodedCid
	return nil
}

func (cid CID) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(cid.ToBytes())
}

func (cid *CID) DecodeMsgpack(dec *msgpack.Decoder) error {
	return decodeMsgpackCID(cid, dec)
}

func CIDFromRegistryPublicKey(pubkey interface{}) (*CID, error) {
	return CIDFromHash(pubkey, 0, CIDTypeResolver, crypto.HashTypeEd25519)
}

func decodeMsgpackCID(cid interface{}, dec *msgpack.Decoder) error {
	byt, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	switch v := cid.(type) {
	case *CID:
		cidInstance, err := CIDFromBytes(byt)
		if err != nil {
			return err
		}

		*v = *cidInstance
	case *EncryptedCID:
		cidInstance, err := EncryptedCIDFromBytes(byt)
		if err != nil {
			return err
		}

		*v = *cidInstance
	default:
		return errors.New("Unsupported type")
	}

	return nil
}

func initCID(bytes []byte) (*CID, error) {
	if len(bytes) == 0 {
		return nil, errEmptyBytes
	}

	cidType := CIDType(bytes[0])
	if cidType == CIDTypeBridge {
		hash := NewMultihash(bytes[1:35])
		return NewCID(cidType, *hash, 0), nil
	}

	if len(bytes) < 34 {
		return nil, errors.New("invalid cid")
	}

	hashBytes := bytes[1:34]
	hash := NewMultihash(hashBytes)

	var size uint64
	if len(bytes) > 34 {
		sizeBytes := bytes[34:]
		sizeValue := DecodeEndian(sizeBytes)
		size = sizeValue
	}

	if _, exists := CIDTypeMap[cidType]; !exists {
		return nil, fmt.Errorf("invalid cid type %d", cidType)
	}

	return NewCID(cidType, *hash, size), nil
}
