package metadata

import (
	"encoding/base64"
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
)

type Base64UrlBinary []byte

func (b *Base64UrlBinary) UnmarshalJSON(data []byte) error {
	strData := string(data)
	if len(strData) >= 2 && strData[0] == '"' && strData[len(strData)-1] == '"' {
		strData = strData[1 : len(strData)-1]
	}

	if strData == "null" {
		return nil
	}

	decodedData, err := base64.RawURLEncoding.DecodeString(strData)
	if err != nil {
		return err
	}

	*b = Base64UrlBinary(decodedData)
	return nil
}
func (b Base64UrlBinary) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(base64.RawURLEncoding.EncodeToString(b)))
}

func decodeIntMap(dec *msgpack.Decoder) (map[int]interface{}, error) {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return nil, err
	}

	data := make(map[int]interface{}, mapLen)

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt()
		if err != nil {
			return nil, err
		}
		value, err := dec.DecodeInterface()
		if err != nil {
			return nil, err
		}
		data[key] = value
	}

	return data, nil
}
