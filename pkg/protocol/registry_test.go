package protocol

import (
	"bytes"
	"go.lumeweb.com/libs5-go/pkg/crypto"
	"testing"
)

func TestSignedRegistryEntry_Basic(t *testing.T) {
	// Create a new keypair for testing
	kp, err := crypto.GenerateEd25519Key()
	if err != nil {
		t.Fatalf("Failed to create keypair: %v", err)
	}

	data := []byte("test data")
	revision := uint64(1)

	// Create and sign a registry entry
	entry := NewRegistryEntry(kp, data, revision)
	signed := entry.Sign()

	// Verify the signed entry contains correct data
	if !bytes.Equal(signed.Data(), data) {
		t.Errorf("Data mismatch. Got %v, want %v", signed.Data(), data)
	}
	if signed.Revision() != revision {
		t.Errorf("Revision mismatch. Got %v, want %v", signed.Revision(), revision)
	}
	if !bytes.Equal(signed.PK(), kp.PublicKey()) {
		t.Errorf("Public key mismatch. Got %v, want %v", signed.PK(), kp.PublicKey())
	}
}

func TestSignedRegistryEntry_Verification(t *testing.T) {
	kp, _ := crypto.GenerateEd25519Key()
	data := []byte("test data")
	revision := uint64(1)

	// Create and sign a registry entry
	entry := NewRegistryEntry(kp, data, revision)
	signed := entry.Sign()

	// Verify the signature
	if !signed.Verify() {
		t.Error("Signature verification failed")
	}

	// Modify the data and verify signature fails
	signed.SetData([]byte("modified data"))
	if signed.Verify() {
		t.Error("Signature verification should fail with modified data")
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	kp, _ := crypto.GenerateEd25519Key()
	data := []byte("test data")
	revision := uint64(1)

	// Create and sign a registry entry
	entry := NewRegistryEntry(kp, data, revision)
	signed := entry.Sign()

	// Marshal the signed entry
	marshaled := MarshalSignedRegistryEntry(signed)

	// Unmarshal and verify
	unmarshaled, err := UnmarshalSignedRegistryEntry(marshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify all fields match
	if !bytes.Equal(unmarshaled.Data(), signed.Data()) {
		t.Error("Data mismatch after marshal/unmarshal")
	}
	if unmarshaled.Revision() != signed.Revision() {
		t.Error("Revision mismatch after marshal/unmarshal")
	}
	if !bytes.Equal(unmarshaled.PK(), signed.PK()) {
		t.Error("Public key mismatch after marshal/unmarshal")
	}
	if !bytes.Equal(unmarshaled.Signature(), signed.Signature()) {
		t.Error("Signature mismatch after marshal/unmarshal")
	}
}

func TestMarshalRegistryEntry(t *testing.T) {
	pk := []byte("test public key")
	data := []byte("test data")
	revision := uint64(1)

	marshaled := MarshalRegistryEntry(pk, data, revision)

	// Verify the structure
	if marshaled[0] != byte(RecordTypeRegistryEntry) {
		t.Error("Invalid record type in marshaled data")
	}

	// Test without public key
	marshaledNoPK := MarshalRegistryEntry(nil, data, revision)
	if len(marshaledNoPK) >= len(marshaled) {
		t.Error("Marshaled data without PK should be shorter")
	}
}

func TestUnmarshalSignedRegistryEntry_InvalidInput(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "Too short input",
			input:   make([]byte, 42),
			wantErr: true,
		},
		{
			name: "Invalid data length",
			// Create a buffer that claims to have more data than actually present
			// First 43 bytes (including 1 byte for type, 33 for pk, 8 for revision, 1 for length)
			// then claim data length of 255 but don't provide that much data
			input:   append(append(make([]byte, 42), byte(255)), make([]byte, 1)...),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := UnmarshalSignedRegistryEntry(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("UnmarshalSignedRegistryEntry() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSignedRegistryEntryImpl_Setters(t *testing.T) {
	sre := &SignedRegistryEntryImpl{}

	pk := []byte("test pk")
	sre.SetPK(pk)
	if !bytes.Equal(sre.PK(), pk) {
		t.Error("SetPK failed")
	}

	revision := uint64(123)
	sre.SetRevision(revision)
	if sre.Revision() != revision {
		t.Error("SetRevision failed")
	}

	data := []byte("test data")
	sre.SetData(data)
	if !bytes.Equal(sre.Data(), data) {
		t.Error("SetData failed")
	}

	sig := []byte("test signature")
	sre.SetSignature(sig)
	if !bytes.Equal(sre.Signature(), sig) {
		t.Error("SetSignature failed")
	}
}
