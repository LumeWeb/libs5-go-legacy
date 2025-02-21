package encoding

import (
	"go.lumeweb.com/libs5-go/pkg/encoding/internal/testdata"
	"reflect"
	"testing"
)

type encoder struct {
	Multibase
	data []byte
}

func (e *encoder) ToBytes() []byte {
	return e.data
}

func newEncoder(data []byte) encoder {
	e := &encoder{data: data}
	m := NewMultibase(e)
	e.Multibase = m

	return *e
}

func TestDecodeString(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name      string
		args      args
		wantBytes []byte
		wantErr   bool
	}{
		{
			name:      "TestValidMultibase_z",
			args:      args{data: testdata.MediaBase58CID},
			wantBytes: testdata.MediaCIDBytes, // Adjust this based on the expected output of multibase.CIDFromString("zabc")
			wantErr:   false,
		},
		{
			name:      "TestValidMultibase_f",
			args:      args{data: testdata.MediaBase16CID},
			wantBytes: testdata.MediaCIDBytes, // Adjust this based on the expected output of multibase.CIDFromString("fxyz")
			wantErr:   false,
		},
		{
			name:      "TestValidMultibase_u",
			args:      args{data: testdata.MediaBase64CID},
			wantBytes: testdata.MediaCIDBytes, // Adjust this based on the expected output of multibase.CIDFromString("uhello")
			wantErr:   false,
		},
		{
			name:      "TestValidMultibase_b",
			args:      args{data: testdata.MediaBase32CID},
			wantBytes: testdata.MediaCIDBytes, // Adjust this based on the expected output of multibase.CIDFromString("bworld")
			wantErr:   false,
		},
		/*		{
				name:      "TestColonPrefix",
				args:      args{data: ":data"},
				wantBytes: []byte(":data"),
				wantErr:   false,
			},*/
		{
			name:      "TestUnsupportedPrefix",
			args:      args{data: "xunsupported"},
			wantBytes: nil,
			wantErr:   true,
		},
		{
			name:      "TestEmptyInput",
			args:      args{data: ""},
			wantBytes: nil,
			wantErr:   true,
		}, /*
			{
				name:      "TestColonOnlyInput",
				args:      args{data: ":"},
				wantBytes: []byte(":"),
				wantErr:   false,
			},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := MultibaseDecodeString(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultibaseDecodeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBytes, tt.wantBytes) {
				t.Errorf("MultibaseDecodeString() gotBytes = %v, want %v", gotBytes, tt.wantBytes)
			}
		})
	}
}

func TestMultibase_ToBase32(t *testing.T) {

	tests := []struct {
		name    string
		encoder encoder
		want    string
		wantErr bool
	}{
		{
			name:    "Is Raw CID",
			encoder: newEncoder(testdata.RawCIDBytes),
			want:    testdata.RawBase32CID,
			wantErr: false,
		}, {
			name:    "Is Media CID",
			encoder: newEncoder(testdata.MediaCIDBytes),
			want:    testdata.MediaBase32CID,
			wantErr: false,
		}, {
			name:    "Is Resolver CID",
			encoder: newEncoder(testdata.ResolverCIDBytes),
			want:    testdata.ResolverBase32CID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.encoder.ToBase32()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBase32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToBase32() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultibase_ToBase58(t *testing.T) {
	tests := []struct {
		name    string
		encoder encoder
		want    string
		wantErr bool
	}{
		{
			name:    "Is Raw CID",
			encoder: newEncoder(testdata.RawCIDBytes),
			want:    testdata.RawBase58CID,
			wantErr: false,
		}, {
			name:    "Is Media CID",
			encoder: newEncoder(testdata.MediaCIDBytes),
			want:    testdata.MediaBase58CID,
			wantErr: false,
		},
		{
			name:    "Is Resolver CID",
			encoder: newEncoder(testdata.ResolverCIDBytes),
			want:    testdata.ResolverBase58CID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.encoder.ToBase58()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBase58() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToBase58() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultibase_ToBase64Url(t *testing.T) {
	tests := []struct {
		name    string
		encoder encoder
		want    string
		wantErr bool
	}{
		{
			name:    "Is Raw CID",
			encoder: newEncoder(testdata.RawCIDBytes),
			want:    testdata.RawBase64CID,
			wantErr: false,
		}, {
			name:    "Is Media CID",
			encoder: newEncoder(testdata.MediaCIDBytes),
			want:    testdata.MediaBase64CID,
			wantErr: false,
		},
		{
			name:    "Is Resolver CID",
			encoder: newEncoder(testdata.ResolverCIDBytes),
			want:    testdata.ResolverBase64CID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.encoder.ToBase64Url()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBase64Url() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToBase64Url() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultibase_ToHex(t *testing.T) {
	tests := []struct {
		name    string
		encoder encoder
		want    string
		wantErr bool
	}{
		{
			name:    "Is Raw CID",
			encoder: newEncoder(testdata.RawCIDBytes),
			want:    testdata.RawBase16CID,
			wantErr: false,
		}, {
			name:    "Is Media CID",
			encoder: newEncoder(testdata.MediaCIDBytes),
			want:    testdata.MediaBase16CID,
			wantErr: false,
		},
		{
			name:    "Is Resolver CID",
			encoder: newEncoder(testdata.ResolverCIDBytes),
			want:    testdata.ResolverBase16CID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.encoder.ToHex()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToHex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultibase_ToString(t *testing.T) {
	TestMultibase_ToBase58(t)
}
