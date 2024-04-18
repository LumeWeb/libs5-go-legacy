package metadata

import (
	"errors"
	"github.com/LumeWeb/libs5-go/serialize"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/vmihailenco/msgpack/v5"
)

type DirectoryMetadata struct {
	Details       DirectoryMetadataDetails `json:"details"`
	Directories   directoryReferenceMap    `json:"directories"`
	Files         fileReferenceMap         `json:"files"`
	ExtraMetadata ExtraMetadata            `json:"extraMetadata"`
	BaseMetadata
}

var _ SerializableMetadata = (*DirectoryMetadata)(nil)

func NewEmptyDirectoryMetadata() *DirectoryMetadata {
	return &DirectoryMetadata{}
}

func NewDirectoryMetadata(details DirectoryMetadataDetails, directories directoryReferenceMap, files fileReferenceMap, extraMetadata ExtraMetadata) *DirectoryMetadata {
	dirMetadata := &DirectoryMetadata{
		Details:       details,
		Directories:   directories,
		Files:         files,
		ExtraMetadata: extraMetadata,
	}

	dirMetadata.Type = "directory"
	return dirMetadata
}
func (dm *DirectoryMetadata) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := serialize.InitMarshaller(enc, types.MetadataTypeDirectory)
	if err != nil {
		return err
	}

	items := make([]interface{}, 4)

	items[0] = dm.Details
	items[1] = dm.Directories
	items[2] = dm.Files
	items[3] = dm.ExtraMetadata.Data

	return enc.Encode(items)
}

func (dm *DirectoryMetadata) DecodeMsgpack(dec *msgpack.Decoder) error {
	_, err := serialize.InitUnmarshaller(dec, types.MetadataTypeDirectory)
	if err != nil {
		return err
	}
	val, err := dec.DecodeArrayLen()

	if err != nil {
		return err
	}

	if val != 4 {
		return errors.New(" Corrupted metadata")
	}

	for i := 0; i < val; i++ {
		switch i {
		case 0:
			err = dec.Decode(&dm.Details)
			if err != nil {
				return err
			}
		case 1:
			err = dec.Decode(&dm.Directories)
			if err != nil {
				return err
			}

		case 2:
			err = dec.Decode(&dm.Files)
			if err != nil {
				return err
			}
		case 3:
			intMap, err := decodeIntMap(dec)
			if err != nil {
				return err
			}
			dm.ExtraMetadata.Data = intMap
		}
	}

	dm.Type = "directory"

	return nil
}
