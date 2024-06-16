package metadata

import (
	"errors"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
	"go.lumeweb.com/libs5-go/serialize"
	"go.lumeweb.com/libs5-go/types"
	"sort"
)

var (
	_ Metadata             = (*WebAppMetadata)(nil)
	_ SerializableMetadata = (*WebAppMetadata)(nil)
	_ SerializableMetadata = (*WebAppFileMap)(nil)
	_ SerializableMetadata = (*WebAppErrorPages)(nil)
)

type WebAppErrorPages map[int]string

type WebAppMetadata struct {
	BaseMetadata
	Name          string           `json:"name"`
	TryFiles      []string         `json:"tryFiles"`
	ErrorPages    WebAppErrorPages `json:"errorPages"`
	ExtraMetadata ExtraMetadata    `json:"extraMetadata"`
	Paths         *WebAppFileMap   `json:"paths"`
}

func NewWebAppMetadata(name string, tryFiles []string, extraMetadata ExtraMetadata, errorPages WebAppErrorPages, paths *WebAppFileMap) *WebAppMetadata {
	return &WebAppMetadata{
		Name:          name,
		TryFiles:      tryFiles,
		ExtraMetadata: extraMetadata,
		ErrorPages:    errorPages,
		Paths:         paths,
	}
}
func NewEmptyWebAppMetadata() *WebAppMetadata {
	return &WebAppMetadata{}
}

func (wm *WebAppMetadata) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := serialize.InitMarshaller(enc, types.MetadataTypeWebApp)
	if err != nil {
		return err
	}

	items := make([]interface{}, 5)

	if wm.ErrorPages == nil {
		wm.ErrorPages = make(WebAppErrorPages)
	}

	items[0] = wm.Name
	items[1] = wm.TryFiles
	items[2] = &wm.ErrorPages
	items[3] = wm.Paths
	items[4] = wm.ExtraMetadata

	return enc.Encode(items)
}

func (wm *WebAppMetadata) DecodeMsgpack(dec *msgpack.Decoder) error {
	_, err := serialize.InitUnmarshaller(dec, types.MetadataTypeWebApp)
	if err != nil {
		return err
	}

	val, err := dec.DecodeArrayLen()

	if err != nil {
		return err
	}

	if val != 5 {
		return errors.New(" Corrupted metadata")
	}

	for i := 0; i < val; i++ {
		switch i {
		case 0:
			wm.Name, err = dec.DecodeString()
			if err != nil {
				return err
			}
		case 1:
			err = dec.Decode(&wm.TryFiles)
			if err != nil {
				return err
			}
		case 2:
			err = dec.Decode(&wm.ErrorPages)
			if err != nil {
				return err
			}
		case 3:
			err = dec.Decode(&wm.Paths)
			if err != nil {
				return err
			}

		case 4:
			err = dec.Decode(&wm.ExtraMetadata)
			if err != nil {
				return err
			}
		default:
			return errors.New(" Corrupted metadata")
		}
	}

	wm.Type = "web_app"

	return nil
}

type WebAppFileMap struct {
	linkedhashmap.Map
}

func NewWebAppFileMap() *WebAppFileMap {
	return &WebAppFileMap{*linkedhashmap.New()}
}

func (wafm *WebAppFileMap) Put(key string, value WebAppMetadataFileReference) {
	wafm.Map.Put(key, value)
}

func (wafm *WebAppFileMap) Get(key string) (WebAppMetadataFileReference, bool) {
	value, found := wafm.Map.Get(key)
	if !found {
		return WebAppMetadataFileReference{}, false
	}
	return value.(WebAppMetadataFileReference), true
}

func (wafm *WebAppFileMap) Remove(key string) {
	wafm.Map.Remove(key)
}

func (wafm *WebAppFileMap) Keys() []string {
	keys := wafm.Map.Keys()
	ret := make([]string, len(keys))
	for i, key := range keys {
		ret[i] = key.(string)
	}
	return ret
}

func (wafm *WebAppFileMap) Values() []WebAppMetadataFileReference {
	values := wafm.Map.Values()
	ret := make([]WebAppMetadataFileReference, len(values))
	for i, value := range values {
		ret[i] = value.(WebAppMetadataFileReference)
	}
	return ret
}

func (wafm *WebAppFileMap) Sort() {
	keys := wafm.Keys()
	newMap := NewWebAppFileMap()

	sort.Strings(keys)

	for _, key := range keys {
		value, _ := wafm.Get(key)
		newMap.Put(key, value)
	}
	wafm.Map = newMap.Map
}

func (wafm *WebAppFileMap) EncodeMsgpack(encoder *msgpack.Encoder) error {
	wafm.Sort()

	err := encoder.EncodeArrayLen(wafm.Size())
	if err != nil {
		return err
	}

	for _, key := range wafm.Keys() {
		value, _ := wafm.Get(key)

		data :=
			make([]interface{}, 3)

		data[0] = key
		data[1] = value.Cid.ToBytes()
		data[2] = value.ContentType

		err := encoder.Encode(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wafm *WebAppFileMap) DecodeMsgpack(decoder *msgpack.Decoder) error {
	arrLen, err := decoder.DecodeArrayLen()
	if err != nil {
		return err
	}

	wafm.Map = *linkedhashmap.New()

	for i := 0; i < arrLen; i++ {
		data := make([]interface{}, 3)

		if len(data) != 3 {
			return errors.New("Corrupted metadata")
		}

		err = decoder.Decode(&data)
		if err != nil {
			return err
		}

		path, ok := data[0].(string)
		if !ok {
			return errors.New("Corrupted metadata")
		}

		cidData, ok := data[1].([]byte)
		if !ok {
			return errors.New("Corrupted metadata")
		}

		contentType, ok := data[2].(string)
		if !ok {
			return errors.New("Corrupted metadata")
		}

		cid, err := encoding.CIDFromBytes(cidData)
		if err != nil {
			return err
		}

		wafm.Put(path, *NewWebAppMetadataFileReference(cid, contentType))
	}

	wafm.Sort()

	return nil
}

func (w *WebAppErrorPages) EncodeMsgpack(enc *msgpack.Encoder) error {
	if w == nil || *w == nil {
		return enc.EncodeMapLen(0)
	}

	err := enc.EncodeMapLen(len(*w))
	if err != nil {
		return err
	}

	for k, v := range *w {
		if err := enc.EncodeInt(int64(k)); err != nil {
			return err
		}
		if err := enc.EncodeString(v); err != nil {
			return err
		}
	}

	return nil
}

func (w *WebAppErrorPages) DecodeMsgpack(dec *msgpack.Decoder) error {
	if *w == nil {
		*w = make(map[int]string)
	}

	mapLen, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}

	*w = make(map[int]string, mapLen)

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt()
		if err != nil {
			return err
		}

		value, err := dec.DecodeString()
		if err != nil {
			return err
		}

		(*w)[key] = value
	}

	return nil
}
