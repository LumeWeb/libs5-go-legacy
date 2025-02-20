package encoding

import "encoding/base64"

func UnmarshalBase64UrlJSON(data []byte) ([]byte, error) {
	strData := string(data)
	if len(strData) >= 2 && strData[0] == '"' && strData[len(strData)-1] == '"' {
		strData = strData[1 : len(strData)-1]
	}

	if strData == "null" {
		return nil, nil
	}

	decodedData, err := MultibaseDecodeString(strData)
	if err != nil {
		if err != ErrMultibaseEncodingNotSupported {
			return nil, err
		}
	} else {
		return decodedData, nil
	}

	decodedData, err = base64.RawURLEncoding.DecodeString(strData)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}
