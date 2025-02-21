package encoding

import (
	"bytes"
	"github.com/vmihailenco/msgpack/v5"
	"net/url"
	"reflect"
	"testing"
)

func TestEncodeMsgpackArray(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		want         []byte
		generateWant func() []byte // Function to generate expected bytes
	}{
		{
			name:  "empty_interface_array",
			input: []interface{}{},
			generateWant: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(0)
				return buf.Bytes()
			},
		},
		{
			name:  "mixed_array",
			input: []interface{}{42, "hello", true},
			generateWant: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(3)
				enc.Encode(42)
				enc.Encode("hello")
				enc.Encode(true)
				return buf.Bytes()
			},
		},
		{
			name: "url_array",
			input: []*url.URL{
				mustParseURL("http://example.com"),
				mustParseURL("https://test.org"),
			},
			generateWant: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(2)
				enc.EncodeString("http://example.com")
				enc.EncodeString("https://test.org")
				return buf.Bytes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := msgpack.NewEncoder(&buf)

			err := EncodeMsgpackArray(enc, tt.input)
			if err != nil {
				t.Fatalf("EncodeMsgpackArray() error = %v", err)
			}

			want := tt.generateWant()
			got := buf.Bytes()

			if !bytes.Equal(got, want) {
				t.Errorf("EncodeMsgpackArray() = %v, want %v", got, want)
			}
		})
	}
}

func TestDecodeMsgpackArray(t *testing.T) {
	tests := []struct {
		name          string
		generateInput func() []byte
		want          []interface{}
	}{
		{
			name: "empty_array",
			generateInput: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(0)
				return buf.Bytes()
			},
			want: []interface{}{},
		},
		{
			name: "mixed_array",
			generateInput: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(3)
				enc.Encode(42)
				enc.Encode("hello")
				enc.Encode(true)
				return buf.Bytes()
			},
			want: []interface{}{42, "hello", true}, // Changed from int64(42) to just 42
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.generateInput()
			dec := msgpack.NewDecoder(bytes.NewReader(input))

			got, err := DecodeMsgpackArray(dec)
			if err != nil {
				t.Fatalf("DecodeMsgpackArray() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeMsgpackArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeMsgpackURLArray(t *testing.T) {
	tests := []struct {
		name          string
		generateInput func() []byte
		want          []*url.URL
		wantErr       bool
	}{
		{
			name: "valid_urls",
			generateInput: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(2)
				enc.Encode("http://example.com")
				enc.Encode("https://test.org")
				return buf.Bytes()
			},
			want: []*url.URL{
				mustParseURL("http://example.com"),
				mustParseURL("https://test.org"),
			},
		},
		{
			name: "invalid_url",
			generateInput: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(1)
				enc.Encode("not-a-url")
				return buf.Bytes()
			},
			wantErr: true,
		},
		{
			name: "wrong_type",
			generateInput: func() []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.EncodeInt(1)
				enc.Encode(42) // Not a string
				return buf.Bytes()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.generateInput()
			dec := msgpack.NewDecoder(bytes.NewReader(input))

			got, err := DecodeMsgpackURLArray(dec)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeMsgpackURLArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeMsgpackURLArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to parse URLs in tests
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// TestRoundTrip tests encoding and decoding in sequence
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "empty_array",
			input: []interface{}{},
		},
		{
			name:  "mixed_values",
			input: []interface{}{42, "hello", true},
		},
		{
			name: "url_array",
			input: []*url.URL{
				mustParseURL("http://example.com"),
				mustParseURL("https://test.org"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			var buf bytes.Buffer
			enc := msgpack.NewEncoder(&buf)

			err := EncodeMsgpackArray(enc, tt.input)
			if err != nil {
				t.Fatalf("EncodeMsgpackArray() error = %v", err)
			}

			// Decode
			dec := msgpack.NewDecoder(bytes.NewReader(buf.Bytes()))

			var got interface{}
			var decodeErr error

			switch tt.input.(type) {
			case []*url.URL:
				got, decodeErr = DecodeMsgpackURLArray(dec)
			default:
				got, decodeErr = DecodeMsgpackArray(dec)
			}

			if decodeErr != nil {
				t.Fatalf("Decode error = %v", decodeErr)
			}

			if !reflect.DeepEqual(got, tt.input) {
				t.Errorf("Round trip failed: got %v, want %v", got, tt.input)
			}
		})
	}
}
