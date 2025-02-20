package metadata

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/serialize"
	"go.lumeweb.com/libs5-go/types"
	"go.lumeweb.com/libs5-go/utils"
	"io"
	"lukechampine.com/blake3"
	_ "lukechampine.com/blake3"
)

var (
	_ Metadata              = (*MediaMetadata)(nil)
	_ msgpack.CustomDecoder = (*MediaMetadata)(nil)
	_ msgpack.CustomEncoder = (*MediaMetadata)(nil)
	_ msgpack.CustomDecoder = (*mediaMap)(nil)
)

type mediaMap map[string][]MediaFormat

type MediaMetadata struct {
	Name          string
	MediaTypes    mediaMap
	Parents       []MetadataParentLink
	Details       MediaMetadataDetails
	Links         *MediaMetadataLinks
	ExtraMetadata ExtraMetadata
	provenPubKeys []*encoding.Multihash
	BaseMetadata
}

func (m *MediaMetadata) ProvenPubKeys() []*encoding.Multihash {
	return m.provenPubKeys
}

func NewMediaMetadata(name string, details MediaMetadataDetails, parents []MetadataParentLink, mediaTypes map[string][]MediaFormat, links *MediaMetadataLinks, extraMetadata ExtraMetadata) *MediaMetadata {
	return &MediaMetadata{
		Name:          name,
		Details:       details,
		Parents:       parents,
		MediaTypes:    mediaTypes,
		Links:         links,
		ExtraMetadata: extraMetadata,
	}
}
func NewEmptyMediaMetadata() *MediaMetadata {
	return &MediaMetadata{}
}

func (m *MediaMetadata) EncodeMsgpack(enc *msgpack.Encoder) error {
	return errors.New("Not implemented")
}

func (m *MediaMetadata) DecodeMsgpack(dec *msgpack.Decoder) error {
	kind, err := serialize.InitUnmarshaller(dec, types.MetadataTypeProof, types.MetadataTypeMedia)
	if err != nil {
		return err
	}

	switch kind {
	case types.MetadataTypeProof:
		return m.decodeProof(dec)
	case types.MetadataTypeMedia:
		return m.decodeMedia(dec)
	default:
		return errors.New("Invalid metadata type")
	}
}

func (m *MediaMetadata) decodeProof(dec *msgpack.Decoder) error {
	all, err := io.ReadAll(dec.Buffered())
	if err != nil {
		return err
	}

	proofSectionLength := utils.DecodeEndian(all[0:2])

	all = all[2:]

	bodyBytes := all[proofSectionLength:]

	if proofSectionLength == 0 {
		return nil
	}

	childDec := msgpack.NewDecoder(bytes.NewReader(all[:proofSectionLength]))

	hash := blake3.Sum256(bodyBytes)
	b3hash := append([]byte{byte(types.HashTypeBlake3)}, hash[:]...)

	arrayLen, err := childDec.DecodeArrayLen()
	if err != nil {
		return err
	}

	provenPubKeys := make([]*encoding.Multihash, 0)

	for i := 0; i < arrayLen; i++ {
		proofData, err := childDec.DecodeSlice()
		if err != nil {
			return err
		}

		var hashType int8
		var pubkey []byte
		var signature []byte

		if len(proofData) != 4 {
			return errors.New("Invalid proof data length")
		}

		for j := 0; j < len(proofData); j++ {

			switch j {
			case 0:
				sigType := proofData[j].(int8)
				if types.MetadataProofType(sigType) != types.MetadataProofTypeSignature {
					return errors.New("Invalid proof type")
				}
			case 1:
				hashType = proofData[j].(int8)
				if types.HashType(hashType) != types.HashTypeBlake3 {
					return errors.New("Invalid hash type")
				}

			case 2:
				pubkey = proofData[j].([]byte)
				if types.HashType(pubkey[0]) != types.HashTypeEd25519 {
					return errors.New("Invalid public key type")
				}

				if len(pubkey) != 33 {
					return errors.New("Invalid public key length")
				}

			case 3:
				signature = proofData[j].([]byte)

				if valid := ed25519.Verify(pubkey[1:], b3hash[:], signature); !valid {
					return errors.New("Invalid signature")
				}

				provenPubKeys = append(provenPubKeys, encoding.NewMultihash(pubkey))
			}
		}
	}

	m.provenPubKeys = provenPubKeys

	mediaDec := msgpack.NewDecoder(bytes.NewReader(bodyBytes))

	mediaByte, err := mediaDec.DecodeUint8()
	if err != nil {
		return err
	}

	if types.MetadataType(mediaByte) != types.MetadataTypeMedia {
		return errors.New("Invalid metadata type")
	}

	return m.decodeMedia(mediaDec)
}

func (m *MediaMetadata) decodeMedia(dec *msgpack.Decoder) error {
	_, err := dec.DecodeArrayLen()
	if err != nil {
		return err
	}

	err = dec.Decode(&m.Name)
	if err != nil {
		return err
	}

	err = dec.Decode(&m.Details)
	if err != nil {
		return err
	}

	arrLen, err := dec.DecodeArrayLen()
	if err != nil {
		return err
	}

	parents := make([]MetadataParentLink, arrLen)
	for i := 0; i < arrLen; i++ {
		parents[i].SetParent(m)
		err = dec.Decode(&parents[i])
		if err != nil {
			return err
		}
	}

	err = dec.Decode(&m.MediaTypes)
	if err != nil {
		return err
	}

	err = dec.Decode(&m.Links)
	if err != nil {
		return err
	}

	err = dec.Decode(&m.ExtraMetadata)
	if err != nil {
		return err
	}

	return nil
}

func (m *mediaMap) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}

	*m = make(map[string][]MediaFormat, mapLen)

	for i := 0; i < mapLen; i++ {
		typ, err := dec.DecodeString()
		if err != nil {
			return err
		}

		var formats []MediaFormat

		err = dec.Decode(&formats)

		if err != nil {
			return err
		}

		(*m)[typ] = formats
	}

	return nil
}
