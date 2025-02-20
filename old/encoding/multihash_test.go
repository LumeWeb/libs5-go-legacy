package encoding

import (
	"go.lumeweb.com/libs5-go/encoding/testdata"
	"go.lumeweb.com/libs5-go/types"
	"reflect"
	"strings"
	"testing"
)

func TestFromBase64Url(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name    string
		args    args
		want    *Multihash
		wantErr bool
	}{
		{
			name:    "Valid Base64 URL Encoded String",
			args:    args{hash: testdata.MediaBase64CID},
			want:    &Multihash{fullBytes: testdata.MediaCIDBytes},
			wantErr: false,
		},
		{
			name:    "Invalid Base64 URL String",
			args:    args{hash: "@@invalid@@"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty String",
			args:    args{hash: ""},
			want:    nil,
			wantErr: true, // or false
		},
		{
			name:    "Non-URL Base64 Encoded String",
			args:    args{hash: "aGVsbG8gd29ybGQ="},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "String Not Representing a Multihash",
			args:    args{hash: "cGxhaW50ZXh0"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Long String",
			args:    args{hash: "uYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFh"},
			want:    &Multihash{fullBytes: []byte(strings.Repeat("a", 750))},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MultihashFromBase64Url(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultihashFromBase64Url() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultihashFromBase64Url() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultihash_FunctionType(t *testing.T) {
	type fields struct {
		FullBytes []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   types.HashType
	}{
		{
			name: "Is Raw CID",
			fields: fields{
				FullBytes: testdata.RawCIDBytes[1:34],
			},
			want: types.HashTypeBlake3,
		}, {
			name: "Is Resolver CID",
			fields: fields{
				FullBytes: testdata.ResolverCIDBytes[1:34],
			},
			want: types.HashTypeEd25519,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Multihash{
				fullBytes: tt.fields.FullBytes,
			}
			if got := m.FunctionType(); got != tt.want {
				t.Errorf("FunctionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultihash_ToBase32(t *testing.T) {
	type fields struct {
		FullBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Is Raw CID",
			fields: fields{
				FullBytes: testdata.RawCIDBytes,
			},
			want: "beyprdl3g2it54ua2ian3a5wybrnva6kr7kzkpv6wbfvgjsb2aepw6yc6t4eq",
		}, {
			name: "Is Media CID",
			fields: fields{
				FullBytes: testdata.MediaCIDBytes,
			},
			want: "byupv6i7z5g6unmx2btc2ihsrrivnox6gqnwfgiwkpuw36d6q4dl35hi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Multihash{
				fullBytes: tt.fields.FullBytes,
			}
			got, err := m.ToBase32()
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

func TestMultihash_ToBase64Url(t *testing.T) {
	type fields struct {
		FullBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Is Raw CID",
			fields: fields{
				FullBytes: testdata.RawCIDBytes,
			},
			want: "uJh8Rr2bSJ95QGkAbsHbYDFtQeVH6sqfX1glqZMg6AR9vYF6fCQ",
		}, {
			name: "Is Media CID",
			fields: fields{
				FullBytes: testdata.MediaCIDBytes,
			},
			want: "uxR9fI_npvUay-gzFpB5RiirXX8aDbFMiyn0tvw_Q4Ne-nQ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Multihash{
				fullBytes: tt.fields.FullBytes,
			}
			got, err := m.ToBase64Url()
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

func TestNewMultihash(t *testing.T) {
	type args struct {
		fullBytes []byte
	}
	tests := []struct {
		name string
		args args
		want *Multihash
	}{
		{
			name: "Valid Base64 URL Encoded String",
			args: args{fullBytes: testdata.RawCIDBytes},
			want: &Multihash{fullBytes: testdata.RawCIDBytes},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMultihash(tt.args.fullBytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultihash() = %v, want %v", got, tt.want)
			}
		})
	}
}
