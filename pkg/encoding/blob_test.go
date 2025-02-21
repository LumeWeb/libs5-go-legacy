package encoding

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"testing"
)

func TestNewMultihash(t *testing.T) {
	testBytes := []byte{0x01, 0x02, 0x03}
	blob := NewMultihash(testBytes)

	if !bytes.Equal(blob.fullBytes, testBytes) {
		t.Errorf("NewMultihash did not store bytes correctly, got %v, want %v", blob.fullBytes, testBytes)
	}
}

func TestBlob_FunctionType(t *testing.T) {
	tests := []struct {
		name     string
		bytes    []byte
		wantType crypto.HashType
	}{
		{
			name:     "Ed25519 hash type",
			bytes:    []byte{byte(crypto.HashTypeEd25519), 0x02, 0x03},
			wantType: crypto.HashTypeEd25519,
		},
		{
			name:     "zero hash type",
			bytes:    []byte{0x00, 0x02, 0x03},
			wantType: crypto.HashType(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob := NewMultihash(tt.bytes)
			got := blob.FunctionType()
			if got != tt.wantType {
				t.Errorf("FunctionType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestBlob_HashBytes(t *testing.T) {
	hashContent := []byte{0x02, 0x03, 0x04}
	fullBytes := append([]byte{byte(crypto.HashTypeEd25519)}, hashContent...)
	blob := NewMultihash(fullBytes)

	gotHash := blob.HashBytes()
	if !bytes.Equal(gotHash, hashContent) {
		t.Errorf("HashBytes() = %v, want %v", gotHash, hashContent)
	}
}

func TestMultihashFromBytes(t *testing.T) {
	content := []byte{0x01, 0x02, 0x03}
	hashType := crypto.HashTypeEd25519

	blob := MultihashFromBytes(content, hashType)

	expectedBytes := append([]byte{byte(hashType)}, content...)
	if !bytes.Equal(blob.fullBytes, expectedBytes) {
		t.Errorf("MultihashFromBytes produced incorrect bytes, got %v, want %v",
			blob.fullBytes, expectedBytes)
	}
}

func TestMultihashFromBase64Url(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03}
	encoded := base64.RawURLEncoding.EncodeToString(original)

	blob, err := MultihashFromBase64Url(encoded)
	if err != nil {
		t.Fatalf("MultihashFromBase64Url() error = %v", err)
	}

	if !bytes.Equal(blob.fullBytes, original) {
		t.Errorf("MultihashFromBase64Url produced incorrect bytes, got %v, want %v",
			blob.fullBytes, original)
	}

	// Test invalid base64
	_, err = MultihashFromBase64Url("invalid!base64")
	if err == nil {
		t.Error("MultihashFromBase64Url() should return error for invalid base64")
	}
}

func TestBlob_ToBase64Url(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03}
	blob := NewMultihash(original)

	got, err := blob.ToBase64Url()
	if err != nil {
		t.Fatalf("ToBase64Url() error = %v", err)
	}

	expected := base64.RawURLEncoding.EncodeToString(original)
	if got != expected {
		t.Errorf("ToBase64Url() = %v, want %v", got, expected)
	}
}

func TestBlob_ToBase32(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03}
	blob := NewMultihash(original)

	got, err := blob.ToBase32()
	if err != nil {
		t.Fatalf("ToBase32() error = %v", err)
	}

	decoded, err := base32.StdEncoding.DecodeString(got)
	if err != nil {
		t.Fatalf("Failed to decode Base32 string: %v", err)
	}

	if !bytes.Equal(decoded, original) {
		t.Errorf("ToBase32() roundtrip failed, got %v, want %v", decoded, original)
	}
}

func TestBlob_ToString(t *testing.T) {
	tests := []struct {
		name    string
		blob    *Blob
		want    string
		wantErr bool
	}{
		{
			name: "bridge type",
			blob: NewMultihash(append([]byte{byte(CIDTypeBridge)}, []byte("test123")...)),
			want: "test123",
		},
		{
			name: "non-bridge type",
			blob: NewMultihash([]byte{0x01, 0x02, 0x03}),
			want: base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x02, 0x03}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.blob.ToString()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlob_Equals(t *testing.T) {
	blob1 := NewMultihash([]byte{0x01, 0x02, 0x03})
	blob2 := NewMultihash([]byte{0x01, 0x02, 0x03})
	blob3 := NewMultihash([]byte{0x03, 0x02, 0x01})

	if !blob1.Equals(blob2) {
		t.Error("Equals() should return true for same content")
	}

	if blob1.Equals(blob3) {
		t.Error("Equals() should return false for different content")
	}
}

func TestBlob_JSON(t *testing.T) {
	// Use something that will produce valid base64url
	original := []byte{0x01, 0x02, 0x03}
	blob := NewMultihash(original)

	// Test marshaling
	data, err := json.Marshal(blob)
	if err != nil {
		t.Fatalf("Failed to marshal Blob: %v", err)
	}

	// Test that we get a valid JSON string
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		t.Fatalf("Unmarshaled data is not a valid JSON string: %v", err)
	}

	// Test unmarshaling back to a blob
	var decoded Blob
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Blob: %v", err)
	}

	if !blob.Equals(&decoded) {
		t.Errorf("JSON roundtrip failed, got %v, want %v", decoded.fullBytes, blob.fullBytes)
	}
}

func TestBlob_FullBytes(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03}
	blob := NewMultihash(original)

	got := blob.FullBytes()
	if !bytes.Equal(got, original) {
		t.Errorf("FullBytes() = %v, want %v", got, original)
	}

	// Modify returned bytes to ensure it doesn't affect original
	got[0] = 0xFF
	if bytes.Equal(got, blob.fullBytes) {
		t.Error("FullBytes() should return a copy, not the original slice")
	}
}
