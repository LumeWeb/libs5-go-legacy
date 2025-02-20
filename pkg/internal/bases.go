package bases

import "github.com/multiformats/go-multibase"

func ToBase64Url(data []byte) (string, error) {
	return ToBase(data, "base64url")
}

func ToBase58BTC(data []byte) (string, error) {
	return ToBase(data, "base58btc")
}

func ToBase32(data []byte) (string, error) {
	return ToBase(data, "base32")
}

func ToHex(data []byte) (string, error) {
	return ToBase(data, "base16")
}

func ToBase(data []byte, base string) (string, error) {
	baseEncoder, _ := multibase.EncoderByName(base)

	ret, err := multibase.Encode(baseEncoder.Encoding(), data)
	if err != nil {
		return "", err
	}

	return ret, nil
}
