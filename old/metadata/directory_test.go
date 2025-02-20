package metadata

import (
	"bytes"
	"encoding/json"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	cmp "github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"path/filepath"
	"testing"
)

func isEqual(sizeFunc1, sizeFunc2 func() int, iteratorFunc1, iteratorFunc2 func() linkedhashmap.Iterator) bool {
	if sizeFunc1() != sizeFunc2() {
		return false
	}

	iter1 := iteratorFunc1()
	iter2 := iteratorFunc2()

	for iter1.Next() {
		iter2.Next()
		if iter1.Key() != iter2.Key() {
			return false
		}
		if !cmp.Equal(iter1.Value(), iter2.Value()) {
			return false
		}
	}

	return true
}

func (frm fileReferenceMap) Equal(other fileReferenceMap) bool {
	return isEqual(frm.Size, other.Size, frm.Iterator, other.Iterator)
}

func (frm FileHistoryMap) Equal(other FileHistoryMap) bool {
	return isEqual(frm.Size, other.Size, frm.Iterator, other.Iterator)
}

func (drm directoryReferenceMap) Equal(other directoryReferenceMap) bool {
	return isEqual(drm.Size, other.Size, drm.Iterator, other.Iterator)
}
func (ext ExtMap) Equal(other ExtMap) bool {
	return isEqual(ext.Size, other.Size, ext.Iterator, other.Iterator)
}
func (fr FileReference) Equal(other FileReference) bool {
	return fr.File.CID().Equals(other.File.CID())
}

func readFile(filename string) []byte {
	filePath := filepath.Join("testdata", filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return data
}

func getDirectoryMeta() *DirectoryMetadata {
	data := getDirectoryMetaContent()

	var dir DirectoryMetadata

	err := json.Unmarshal(data, &dir)
	if err != nil {
		panic(err)
	}

	return &dir
}

func getDirectoryMetaContent() []byte {
	data := readFile("directory.json")

	return data
}

func TestDirectoryMetadata_DecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Decode",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonDm := getDirectoryMeta()
			dm := &DirectoryMetadata{}

			if err := msgpack.Unmarshal(readFile("directory.bin"), dm); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(jsonDm, dm) {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", "msgpack does not match json", tt.wantErr)
			}
		})
	}
}

func TestDirectoryMetadata_DecodeMsgpack(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Decode",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonDm := getDirectoryMeta()
			dm := &DirectoryMetadata{}

			if err := msgpack.Unmarshal(readFile("directory.bin"), dm); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(jsonDm, dm) {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", "msgpack does not match json", tt.wantErr)
			}
		})
	}
}

func TestDirectoryMetadata_EncodeMsgpack(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Encode",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &DirectoryMetadata{}

			good := readFile("directory.bin")

			if err := msgpack.Unmarshal(good, dm); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			out, err := msgpack.Marshal(dm)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(good, out) {
				t.Errorf("EncodeMsgpack() error = %v, wantErr %v", "msgpack does not match sample", tt.wantErr)
			}

			dm2 := &DirectoryMetadata{}

			if err := msgpack.Unmarshal(out, dm2); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(dm, dm2) {
				t.Errorf("EncodeMsgpack() error = %v, wantErr %v", "msgpack deser verification does not match", tt.wantErr)
			}
		})
	}
}
func TestDirectoryMetadata_EncodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Encode",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonDm := getDirectoryMetaContent()
			dm := &DirectoryMetadata{}

			if err := json.Unmarshal(jsonDm, dm); (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			jsonData, err := json.MarshalIndent(dm, "", "\t")

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			buf := bytes.NewBuffer(nil)

			err = json.Indent(buf, jsonData, "", "\t")
			if err != nil {
				t.Errorf("EncodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, buf.Bytes(), jsonData)
		})
	}
}
