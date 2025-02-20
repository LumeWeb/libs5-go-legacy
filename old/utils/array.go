package utils

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"net/url"
)

func EncodeMsgpackArray(enc *msgpack.Encoder, array interface{}) error {
	switch v := array.(type) {
	case []*url.URL:
		// Handle []*url.URL slice
		err := enc.EncodeInt(int64(len(v)))
		if err != nil {
			return err
		}
		for _, item := range v {
			err = enc.EncodeString(item.String())
			if err != nil {
				return err
			}
		}
		return nil
	default:
		// Handle generic case
		arr, ok := array.([]interface{})
		if !ok {
			return errors.New("unsupported type for EncodeMsgpackArray")
		}
		err := enc.EncodeInt(int64(len(arr)))
		if err != nil {
			return err
		}
		for _, item := range arr {
			err = enc.Encode(item)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func DecodeMsgpackArray(dec *msgpack.Decoder) ([]interface{}, error) {
	arrayLen, err := dec.DecodeInt()
	if err != nil {
		return nil, err
	}

	array := make([]interface{}, arrayLen)

	for i := 0; i < int(arrayLen); i++ {
		item, err := dec.DecodeInterface()
		if err != nil {
			return nil, err
		}

		array[i] = item
	}

	return array, nil
}
func DecodeMsgpackURLArray(dec *msgpack.Decoder) ([]*url.URL, error) {
	arrayLen, err := dec.DecodeInt()
	if err != nil {
		return nil, err
	}

	urlArray := make([]*url.URL, arrayLen)

	for i := 0; i < int(arrayLen); i++ {
		item, err := dec.DecodeInterface()
		if err != nil {
			return nil, err
		}

		// Type assert each item to *url.URL
		urlItem, ok := item.(string)
		if !ok {
			// Handle the case where the item is not a *url.URL
			return nil, fmt.Errorf("expected string, got %T", item)
		}

		urlArray[i], err = url.Parse(urlItem)
		if err != nil {
			return nil, err
		}
	}

	return urlArray, nil
}
