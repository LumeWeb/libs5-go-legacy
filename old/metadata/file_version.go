package metadata

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
)

type FileVersion struct {
	Ts           uint64                 `json:"ts"`
	EncryptedCID *encoding.EncryptedCID `json:"encryptedCID,string"`
	PlaintextCID *encoding.CID          `json:"cid,string"`
	Thumbnail    *FileVersionThumbnail  `json:"thumbnail"`
	Hashes       []*encoding.Multihash  `json:"hashes"`
	Ext          map[string]interface{} `json:"ext"`
}

func NewFileVersion(ts uint64, encryptedCID *encoding.EncryptedCID, plaintextCID *encoding.CID, thumbnail *FileVersionThumbnail, hashes []*encoding.Multihash, ext map[string]interface{}) *FileVersion {
	return &FileVersion{
		Ts:           ts,
		EncryptedCID: encryptedCID,
		PlaintextCID: plaintextCID,
		Thumbnail:    thumbnail,
		Hashes:       hashes,
		Ext:          ext,
	}
}

func (fv *FileVersion) EncodeMsgpack(enc *msgpack.Encoder) error {

	fmap := &fileVersionSerializationMap{*linkedhashmap.New()}

	fmap.Put(8, fv.Ts)

	if fv.EncryptedCID != nil {
		fmap.Put(1, fv.EncryptedCID)
	}

	if fv.PlaintextCID != nil {
		fmap.Put(2, fv.PlaintextCID)
	}

	if len(fv.Hashes) > 0 {
		hashesData := make([][]byte, len(fv.Hashes))
		for i, hash := range fv.Hashes {
			hashesData[i] = hash.FullBytes()
		}
		fmap.Put(9, hashesData)
	}

	if fv.Thumbnail != nil {
		fmap.Put(10, fv.Thumbnail)
	}

	return enc.Encode(fmap)
}

func (fv *FileVersion) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return err
	}

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt8()
		if err != nil {
			return err
		}
		switch key {

		case int8(1):
			err := dec.Decode(&fv.EncryptedCID)
			if err != nil {
				return err
			}

		case int8(2):
			err := dec.Decode(&fv.PlaintextCID)
			if err != nil {
				return err
			}
		case int8(8):
			err := dec.Decode(&fv.Ts)
			if err != nil {
				return err
			}
		case int8(9):
			hashesData, err := dec.DecodeSlice()
			if err != nil {
				return err
			}

			fv.Hashes = make([]*encoding.Multihash, len(hashesData))
			for i, hashData := range hashesData {
				hashBytes := hashData.([]byte)
				fv.Hashes[i] = encoding.NewMultihash(hashBytes)
			}

		case int8(10):
			err := dec.Decode(&fv.Thumbnail)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fv *FileVersion) CID() *encoding.CID {
	if fv.PlaintextCID != nil {
		return fv.PlaintextCID
	}
	return &fv.EncryptedCID.OriginalCID
}
