package metadata

import (
	"encoding/json"
	"fmt"
	"github.com/LumeWeb/libs5-go/encoding"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/vmihailenco/msgpack/v5"
)

type jsonData = map[string]interface{}

var names = map[types.MetadataExtension]string{
	types.MetadataExtensionLicenses:           "licenses",
	types.MetadataExtensionDonationKeys:       "donationKeys",
	types.MetadataExtensionWikidataClaims:     "wikidataClaims",
	types.MetadataExtensionLanguages:          "languages",
	types.MetadataExtensionSourceUris:         "sourceUris",
	types.MetadataExtensionPreviousVersions:   "previousVersions",
	types.MetadataExtensionTimestamp:          "timestamp",
	types.MetadataExtensionOriginalTimestamp:  "originalTimestamp",
	types.MetadataExtensionTags:               "tags",
	types.MetadataExtensionCategories:         "categories",
	types.MetadataExtensionBasicMediaMetadata: "basicMediaMetadata",
	types.MetadataExtensionViewTypes:          "viewTypes",
	types.MetadataExtensionBridge:             "bridge",
	types.MetadataExtensionRoutingHints:       "routingHints",
}

var namesReverse = map[string]types.MetadataExtension{
	"licenses":           types.MetadataExtensionLicenses,
	"donationKeys":       types.MetadataExtensionDonationKeys,
	"wikidataClaims":     types.MetadataExtensionWikidataClaims,
	"languages":          types.MetadataExtensionLanguages,
	"sourceUris":         types.MetadataExtensionSourceUris,
	"previousVersions":   types.MetadataExtensionPreviousVersions,
	"timestamp":          types.MetadataExtensionTimestamp,
	"originalTimestamp":  types.MetadataExtensionOriginalTimestamp,
	"tags":               types.MetadataExtensionTags,
	"categories":         types.MetadataExtensionCategories,
	"basicMediaMetadata": types.MetadataExtensionBasicMediaMetadata,
	"viewTypes":          types.MetadataExtensionViewTypes,
	"bridge":             types.MetadataExtensionBridge,
	"routingHints":       types.MetadataExtensionRoutingHints,
}

type ExtraMetadata struct {
	Data map[int]interface{}
}
type keyValue struct {
	Key   interface{}
	Value interface{}
}

var _ SerializableMetadata = (*ExtraMetadata)(nil)

func NewExtraMetadata(data map[int]interface{}) *ExtraMetadata {
	return &ExtraMetadata{
		Data: data,
	}
}

func (em ExtraMetadata) MarshalJSON() ([]byte, error) {
	data, err := em.encode()
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}

func (em *ExtraMetadata) UnmarshalJSON(data []byte) error {

	em.Data = make(map[int]interface{})
	jsonObject := make(map[int]interface{})
	if err := json.Unmarshal(data, &jsonObject); err != nil {
		return err
	}

	for name, value := range jsonObject {
		err := em.decodeItem(keyValue{Key: name, Value: value})
		if err != nil {
			return err
		}
	}
	return nil
}

func (em *ExtraMetadata) decodeItem(pair keyValue) error {
	var metadataKey int

	// Determine the type of the key and convert it if necessary
	switch k := pair.Key.(type) {
	case string:
		if val, ok := namesReverse[k]; ok {
			metadataKey = int(val)
		} else {
			return fmt.Errorf("unknown key in JSON: %s", k)
		}
	case int8:
		metadataKey = int(k)
	default:
		return fmt.Errorf("unsupported key type")
	}

	if metadataKey == int(types.MetadataExtensionUpdateCID) {
		cid, err := encoding.CIDFromBytes([]byte(pair.Value.(string)))
		if err != nil {
			return err
		}
		em.Data[metadataKey] = cid
	} else {
		em.Data[metadataKey] = pair.Value
	}

	return nil
}

func (em *ExtraMetadata) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return err
	}

	em.Data = make(map[int]interface{}, mapLen)

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt8()
		if err != nil {
			return err
		}

		var value interface{}
		if key == int8(types.MetadataExtensionUpdateCID) {
			value, err = dec.DecodeString()
		} else {
			value, err = dec.DecodeInterface()
		}
		if err != nil {
			return err
		}

		err = em.decodeItem(keyValue{Key: key, Value: value})
		if err != nil {
			return err
		}
	}

	if mapLen == 0 {
		em.Data = make(map[int]interface{})
	}

	return nil
}

func (em ExtraMetadata) EncodeMsgpack(enc *msgpack.Encoder) error {
	data, err := em.encode()
	if err != nil {
		return err
	}

	return enc.Encode(data)
}

func (em ExtraMetadata) encode() (jsonData, error) {
	jsonObject := make(map[string]interface{})
	for key, value := range em.Data {
		name, ok := names[types.MetadataExtension(key)]
		if ok {
			if types.MetadataExtension(key) == types.MetadataExtensionUpdateCID {
				cid, err := encoding.CIDFromBytes(value.([]byte))
				var cidString string
				if err == nil {
					cidString, err = cid.ToString()
				}
				if err == nil {
					jsonObject["updateCID"] = cidString
				} else {
					jsonObject["updateCID"] = ""
				}
			} else {
				jsonObject[name] = value
			}
		}
	}

	return jsonObject, nil
}
