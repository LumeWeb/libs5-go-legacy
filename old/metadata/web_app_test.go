package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"testing"
)

func (wafr WebAppMetadataFileReference) Equal(other WebAppMetadataFileReference) bool {
	return wafr.Cid.Equals(other.Cid) && wafr.ContentType == other.ContentType
}

func getWebappMeta() *WebAppMetadata {
	data := getWebappContent()

	var webapp WebAppMetadata

	err := json.Unmarshal(data, &webapp)
	if err != nil {
		panic(err)
	}

	return &webapp
}

func getWebappContent() []byte {
	data := readFile("webapp.json")

	return data
}

func TestWebAppMetadata_DecodeJSON(t *testing.T) {
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
			jsonDm := getWebappMeta()
			dm := &WebAppMetadata{}

			if err := msgpack.Unmarshal(readFile("webapp.bin"), dm); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(jsonDm, dm) {
				fmt.Println(cmp.Diff(jsonDm, dm))
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", "msgpack does not match json", tt.wantErr)
			}
		})
	}
}

func TestWebAppMetadata_DecodeMsgpack(t *testing.T) {
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
			jsonDm := getWebappMeta()
			dm := &WebAppMetadata{}

			if err := msgpack.Unmarshal(readFile("webapp.bin"), dm); (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(jsonDm, dm) {
				t.Errorf("DecodeMsgpack() error = %v, wantErr %v", "msgpack does not match json", tt.wantErr)
			}
		})
	}
}

func TestWebAppMetadata_EncodeJSON(t *testing.T) {
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
			jsonDm := getWebappContent()
			dm := &WebAppMetadata{}

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
