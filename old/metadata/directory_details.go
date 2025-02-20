package metadata

import "github.com/vmihailenco/msgpack/v5"

type DirectoryMetadataDetails struct {
	Data map[int]interface{}
}

func NewDirectoryMetadataDetails(data map[int]interface{}) *DirectoryMetadataDetails {
	return &DirectoryMetadataDetails{
		Data: data,
	}
}

func (dmd *DirectoryMetadataDetails) IsShared() bool {
	_, exists := dmd.Data[3]
	return exists
}

func (dmd *DirectoryMetadataDetails) IsSharedReadOnly() bool {
	value, exists := dmd.Data[3].([]interface{})
	if !exists {
		return false
	}
	return len(value) > 1 && value[1] == true
}

func (dmd *DirectoryMetadataDetails) IsSharedReadWrite() bool {
	value, exists := dmd.Data[3].([]interface{})
	if !exists {
		return false
	}
	return len(value) > 2 && value[2] == true
}

func (dmd *DirectoryMetadataDetails) SetShared(value bool, write bool) {
	if dmd.Data == nil {
		dmd.Data = make(map[int]interface{})
	}
	sharedValue, exists := dmd.Data[3].([]interface{})
	if !exists {
		sharedValue = make([]interface{}, 3)
		dmd.Data[3] = sharedValue
	}
	if write {
		sharedValue[2] = value
	} else {
		sharedValue[1] = value
	}
}
func (dmd *DirectoryMetadataDetails) DecodeMsgpack(dec *msgpack.Decoder) error {
	mapLen, err := dec.DecodeMapLen()

	if err != nil {
		return err
	}

	for i := 0; i < mapLen; i++ {
		key, err := dec.DecodeInt8()
		if err != nil {
			return err
		}
		value, err := dec.DecodeInterface()
		if err != nil {
			return err
		}
		dmd.Data[int(key)] = value
	}

	return nil
}

func (dmd DirectoryMetadataDetails) EncodeMsgpack(enc *msgpack.Encoder) error {
	if dmd.Data == nil {
		dmd.Data = make(map[int]interface{})
	}
	return enc.Encode(dmd.Data)
}
