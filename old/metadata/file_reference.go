package metadata

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/vmihailenco/msgpack/v5"
)

var _ SerializableMetadata = (*FileReference)(nil)
var _ SerializableMetadata = (*FileHistoryMap)(nil)
var _ SerializableMetadata = (*ExtMap)(nil)

type FileHistoryMap struct {
	linkedhashmap.Map
}
type ExtMap struct {
	linkedhashmap.Map
}

func NewExtMap() ExtMap {
	return ExtMap{*linkedhashmap.New()}
}

func NewFileHistoryMap() FileHistoryMap {
	return FileHistoryMap{*linkedhashmap.New()}
}

type FileReference struct {
	Name     string         `json:"name"`
	Created  uint64         `json:"created"`
	Version  uint64         `json:"version"`
	File     *FileVersion   `json:"file"`
	Ext      ExtMap         `json:"ext"`
	History  FileHistoryMap `json:"history"`
	MimeType string         `json:"mimeType"`
	URI      string         `json:"uri,omitempty"`
	Key      string         `json:"key,omitempty"`
}

func NewFileReference(name string, created, version uint64, file *FileVersion, ext ExtMap, history FileHistoryMap, mimeType string) *FileReference {
	return &FileReference{
		Name:     name,
		Created:  created,
		Version:  version,
		File:     file,
		Ext:      ext,
		History:  history,
		MimeType: mimeType,
		URI:      "",
		Key:      "",
	}
}

func (fr *FileReference) Modified() uint64 {
	return fr.File.Ts
}

func (fr *FileReference) EncodeMsgpack(enc *msgpack.Encoder) error {
	tempMap := &fileReferenceSerializationMap{*linkedhashmap.New()}

	tempMap.Put(1, fr.Name)
	tempMap.Put(2, fr.Created)
	tempMap.Put(4, fr.File)
	tempMap.Put(5, fr.Version)

	if fr.MimeType != "" {
		tempMap.Put(6, fr.MimeType)
	}

	if !fr.Ext.Empty() {
		tempMap.Put(7, fr.Ext)
	}

	if !fr.History.Empty() {
		tempMap.Put(8, fr.History)
	}

	return enc.Encode(tempMap)
}
func (fr *FileReference) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return err
	}

	hasExt := false
	hasHistory := false

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt8()
		if err != nil {
			return err
		}

		switch key {
		case int8(1):
			err := dec.Decode(&fr.Name)
			if err != nil {
				return err
			}
		case int8(2):
			err := dec.Decode(&fr.Created)
			if err != nil {
				return err
			}
		case int8(4):
			err := dec.Decode(&fr.File)
			if err != nil {
				return err
			}
		case int8(5):
			val, err := dec.DecodeInt()
			if err != nil {
				return err
			}

			fr.Version = uint64(val)
		case int8(6):
			err := dec.Decode(&fr.MimeType)
			if err != nil {
				return err
			}
		case int8(7):
			err := dec.Decode(&fr.Ext)
			if err != nil {
				return err
			}

			hasExt = true
		case int8(8):
			err := dec.Decode(&fr.History)
			if err != nil {
				return err
			}

			hasHistory = true
		}
	}

	if !hasExt {
		fr.Ext = ExtMap{*linkedhashmap.New()}
	}

	if !hasHistory {
		fr.History = FileHistoryMap{*linkedhashmap.New()}
	}

	return nil
}

func (ext ExtMap) EncodeMsgpack(enc *msgpack.Encoder) error {
	return marshallMapMsgpack(enc, &ext.Map)
}
func (ext *ExtMap) DecodeMsgpack(dec *msgpack.Decoder) error {
	return unmarshalMapMsgpack(dec, &ext.Map, &ExtMap{}, true)
}
func (fhm FileHistoryMap) EncodeMsgpack(enc *msgpack.Encoder) error {
	return marshallMapMsgpack(enc, &fhm.Map)
}
func (fhm *FileHistoryMap) DecodeMsgpack(dec *msgpack.Decoder) error {
	return unmarshalMapMsgpack(dec, &fhm.Map, &ExtMap{}, false)
}

func (m *FileHistoryMap) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		m.Map = *linkedhashmap.New()
		return nil
	}
	return m.FromJSON(bytes)
}

func (m *ExtMap) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		m.Map = *linkedhashmap.New()
		return nil
	}
	return m.FromJSON(bytes)
}
