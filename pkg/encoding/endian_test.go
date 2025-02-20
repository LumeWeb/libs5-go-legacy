package encoding

import (
	"math"
	"testing"
)

func TestEncodeDecodeEndian(t *testing.T) {
	tests := []struct {
		name   string
		value  uint64
		length int
	}{
		{"zero_1byte", 0, 1},
		{"zero_8bytes", 0, 8},
		{"small_number", 42, 1},
		{"max_uint8", 255, 1},
		{"two_bytes", 256, 2},
		{"large_number", 123456789, 4},
		{"max_uint64", math.MaxUint64, 8},
		{"specific_pattern", 0xAABBCCDD, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeEndian(tt.value, tt.length)

			// Check encoded length
			if len(encoded) != tt.length {
				t.Errorf("EncodeEndian(%d, %d) length = %d; want %d",
					tt.value, tt.length, len(encoded), tt.length)
			}

			// Decode and verify roundtrip
			decoded := DecodeEndian(encoded)

			// For values that fit in the specified length, they should match exactly
			expectedValue := tt.value
			if tt.length < 8 {
				// Mask the expected value to the appropriate number of bytes
				mask := uint64((1 << (tt.length * 8)) - 1)
				expectedValue &= mask
			}

			if decoded != expectedValue {
				t.Errorf("Roundtrip failed: original=%d, got=%d, expected=%d",
					tt.value, decoded, expectedValue)
			}
		})
	}
}

func TestEncodeEndianExtremeValues(t *testing.T) {
	tests := []struct {
		name   string
		value  uint64
		length int
		want   []byte
	}{
		{
			name:   "single_byte_max",
			value:  0xFF,
			length: 1,
			want:   []byte{0xFF},
		},
		{
			name:   "two_bytes_pattern",
			value:  0xAABB,
			length: 2,
			want:   []byte{0xBB, 0xAA},
		},
		{
			name:   "truncate_value",
			value:  0xFFFFFFFF,
			length: 2,
			want:   []byte{0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeEndian(tt.value, tt.length)
			if len(got) != len(tt.want) {
				t.Errorf("EncodeEndian(%d, %d) length = %d; want %d",
					tt.value, tt.length, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("EncodeEndian(%d, %d) at position %d = %02x; want %02x",
						tt.value, tt.length, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDecodeEndianExtremeValues(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  uint64
	}{
		{
			name:  "single_byte_max",
			input: []byte{0xFF},
			want:  0xFF,
		},
		{
			name:  "two_bytes_pattern",
			input: []byte{0xBB, 0xAA},
			want:  0xAABB,
		},
		{
			name:  "four_bytes_pattern",
			input: []byte{0xDD, 0xCC, 0xBB, 0xAA},
			want:  0xAABBCCDD,
		},
		{
			name:  "empty_slice",
			input: []byte{},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeEndian(tt.input)
			if got != tt.want {
				t.Errorf("DecodeEndian(%v) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestEdgeCases tests some edge cases and potential error conditions
func TestEdgeCases(t *testing.T) {
	// Test encoding with length 0
	encoded := EncodeEndian(42, 0)
	if len(encoded) != 0 {
		t.Errorf("EncodeEndian with length 0 should return empty slice, got length %d", len(encoded))
	}

	// Test encoding with very large length
	encoded = EncodeEndian(42, 16)
	if len(encoded) != 16 {
		t.Errorf("EncodeEndian with length 16 should return slice of length 16, got length %d", len(encoded))
	}

	// Test decoding nil slice
	decoded := DecodeEndian(nil)
	if decoded != 0 {
		t.Errorf("DecodeEndian(nil) should return 0, got %d", decoded)
	}
}
