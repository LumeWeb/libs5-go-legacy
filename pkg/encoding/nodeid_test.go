package encoding

import (
	"bytes"
	"encoding/hex"
	"github.com/multiformats/go-multibase"
	"testing"
)

func TestNewNodeId(t *testing.T) {
	testBytes := []byte("test123")
	nodeId := NewNodeId(testBytes)

	if nodeId == nil {
		t.Error("NewNodeId returned nil")
	}

	if !bytes.Equal(nodeId.bytes, testBytes) {
		t.Errorf("NewNodeId did not store bytes correctly, got %v, want %v", nodeId.bytes, testBytes)
	}

	// Verify defensive copy
	testBytes[0] = 'x'
	if bytes.Equal(nodeId.bytes, testBytes) {
		t.Error("NewNodeId should make a copy of input bytes")
	}
}

func TestNodeId_Bytes(t *testing.T) {
	testBytes := []byte("test123")
	nodeId := NewNodeId(testBytes)

	result := nodeId.Bytes()
	if !bytes.Equal(result, testBytes) {
		t.Errorf("Bytes() returned incorrect value, got %v, want %v", result, testBytes)
	}

	// Ensure returned bytes are a copy and not the original reference
	result[0] = 'x'
	if bytes.Equal(result, nodeId.bytes) {
		t.Error("Bytes() should return a copy of the bytes, not the original reference")
	}
}

func generateTestBase58(t *testing.T, input string) string {
	encoded, err := multibase.Encode(multibase.Base58BTC, []byte(input))
	if err != nil {
		t.Fatalf("Failed to generate test base58: %v", err)
	}
	return encoded
}

func TestDecodeNodeId(t *testing.T) {
	testStr := "test123"
	validBase58 := generateTestBase58(t, testStr)

	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr error
	}{
		{
			name:    "valid base58btc",
			input:   validBase58,
			want:    []byte(testStr),
			wantErr: nil,
		},
		{
			name:    "invalid base58",
			input:   "invalid!!!",
			wantErr: multibase.ErrUnsupportedEncoding,
		},
		{
			name:    "wrong encoding",
			input:   "b64", // base64
			wantErr: errorNotBase58BTC,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: multibase.ErrUnsupportedEncoding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeNodeId(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("DecodeNodeId() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("DecodeNodeId() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("DecodeNodeId() unexpected error = %v", err)
				return
			}

			if !bytes.Equal(got.bytes, tt.want) {
				t.Errorf("DecodeNodeId() got = %v, want %v",
					hex.EncodeToString(got.bytes),
					hex.EncodeToString(tt.want))
			}
		})
	}
}

func TestNodeId_Equals(t *testing.T) {
	tests := []struct {
		name     string
		nodeId   *NodeId
		other    interface{}
		expected bool
	}{
		{
			name:     "same content",
			nodeId:   NewNodeId([]byte("test123")),
			other:    NewNodeId([]byte("test123")),
			expected: true,
		},
		{
			name:     "different content",
			nodeId:   NewNodeId([]byte("test123")),
			other:    NewNodeId([]byte("test456")),
			expected: false,
		},
		{
			name:     "nil comparison",
			nodeId:   NewNodeId([]byte("test123")),
			other:    nil,
			expected: false,
		},
		{
			name:     "wrong type",
			nodeId:   NewNodeId([]byte("test123")),
			other:    "test123",
			expected: false,
		},
		{
			name:     "empty bytes",
			nodeId:   NewNodeId([]byte{}),
			other:    NewNodeId([]byte{}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nodeId.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNodeId_ToBase58(t *testing.T) {
	testStr := "test123"
	validBase58 := generateTestBase58(t, testStr)

	tests := []struct {
		name    string
		bytes   []byte
		want    string
		wantErr bool
	}{
		{
			name:    "valid bytes",
			bytes:   []byte(testStr),
			want:    validBase58,
			wantErr: false,
		},
		{
			name:    "empty bytes",
			bytes:   []byte{},
			want:    "z",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeId := NewNodeId(tt.bytes)
			got, err := nodeId.ToBase58()

			if (err != nil) != tt.wantErr {
				t.Errorf("ToBase58() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ToBase58() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeId_ToString(t *testing.T) {
	nodeId := NewNodeId([]byte("test123"))
	base58Str, err := nodeId.ToBase58()
	if err != nil {
		t.Fatalf("ToBase58() failed: %v", err)
	}

	toString, err := nodeId.ToString()
	if err != nil {
		t.Errorf("ToString() unexpected error = %v", err)
	}

	if toString != base58Str {
		t.Errorf("ToString() = %v, want %v", toString, base58Str)
	}
}

func TestNodeId_Raw(t *testing.T) {
	testBytes := []byte("test123")
	nodeId := NewNodeId(testBytes)

	result := nodeId.Raw()
	if !bytes.Equal(result, testBytes) {
		t.Errorf("Raw() returned incorrect value, got %v, want %v", result, testBytes)
	}

	// Ensure returned bytes are a copy and not the original reference
	result[0] = 'x'
	if bytes.Equal(result, nodeId.bytes) {
		t.Error("Raw() should return a copy of the bytes, not the original reference")
	}
}

func TestNodeId_RoundTrip(t *testing.T) {
	originalBytes := []byte("test123")
	nodeId := NewNodeId(originalBytes)

	// Convert to string
	str, err := nodeId.ToString()
	if err != nil {
		t.Fatalf("ToString() failed: %v", err)
	}

	// Decode back to NodeId
	decoded, err := DecodeNodeId(str)
	if err != nil {
		t.Fatalf("DecodeNodeId() failed: %v", err)
	}

	// Compare
	if !nodeId.Equals(decoded) {
		t.Error("Round trip conversion failed, NodeIds are not equal")
	}

	if !bytes.Equal(originalBytes, decoded.Bytes()) {
		t.Error("Round trip conversion failed, bytes are not equal")
	}
}
