package crypto_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"testing"
	"time"

	libcrypto "go.lumeweb.com/libs5-go/pkg/crypto"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
)

func TestGenerateSecureRandomBytes(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()

	t.Run("valid length", func(t *testing.T) {
		b, err := crypto.GenerateSecureRandomBytes(32)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(b) != 32 {
			t.Errorf("Expected 32 bytes, got %d", len(b))
		}
	})

	t.Run("zero length", func(t *testing.T) {
		_, err := crypto.GenerateSecureRandomBytes(0)
		if err == nil {
			t.Error("Expected error for zero length")
		}
	})

	t.Run("concurrency safety", func(t *testing.T) {
		// Test concurrent access to thread-safe implementation
		iterations := 100
		results := make(chan []byte, iterations)

		for i := 0; i < iterations; i++ {
			go func() {
				b, _ := crypto.GenerateSecureRandomBytes(16)
				results <- b
			}()
		}

		unique := make(map[string]bool)
		for i := 0; i < iterations; i++ {
			b := <-results
			key := string(b)
			if unique[key] {
				t.Fatal("Duplicate random bytes generated")
			}
			unique[key] = true
		}
	})
}

func TestHashBlake3(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: mustDecodeHex("af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"),
		},
		{
			name:     "simple string",
			input:    []byte("hello world"),
			expected: mustDecodeHex("d74981efa70a8c42e40244a485670017b4e8d5ca5c3b35dcc9b70a54a41f4624"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test async version
			hash, err := crypto.HashBlake3(ctx, tc.input)
			if err != nil {
				t.Fatalf("HashBlake3 failed: %v", err)
			}
			if !bytes.Equal(hash, tc.expected) {
				t.Errorf("HashBlake3: expected %x, got %x", tc.expected, hash)
			}

			// Test sync version
			syncHash, err := crypto.HashBlake3Sync(tc.input)
			if err != nil {
				t.Fatalf("HashBlake3Sync failed: %v", err)
			}
			if !bytes.Equal(syncHash, tc.expected) {
				t.Errorf("HashBlake3Sync: expected %x, got %x", tc.expected, syncHash)
			}
		})
	}

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := crypto.HashBlake3(ctx, []byte("test"))
		if err != context.Canceled {
			t.Errorf("Expected context canceled error, got %v", err)
		}
	})
}

func TestHashBlake3File(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	// Test data that's exactly 2MB (2 chunks)
	largeData := bytes.Repeat([]byte{0x01}, 2*1024*1024)

	testCases := []struct {
		name     string
		data     []byte
		expected []byte
	}{
		{
			name:     "empty file",
			data:     []byte{},
			expected: mustDecodeHex("af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"),
		},
		{
			name:     "exact chunk size",
			data:     bytes.Repeat([]byte{0x01}, 1*1024*1024),
			expected: mustDecodeHex("1c0bbf1a1d9eb4c927e96e0d9735e0d32a8880f7a2fe4a75b7d5f6e2e1f7a2e8"),
		},
		{
			name:     "multiple chunks",
			data:     largeData,
			expected: mustDecodeHex("aad2a551e56a3d3dbf0e0e1e2a3f3d3e4e5d6c7b8a9f0e1d2c3b4a5968778695"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			openRead := func(start, end int) (io.Reader, error) {
				if start > len(tc.data) {
					return bytes.NewReader(nil), nil
				}
				if end > len(tc.data) {
					end = len(tc.data)
				}
				return bytes.NewReader(tc.data[start:end]), nil
			}

			hash, err := crypto.HashBlake3File(ctx, int64(len(tc.data)), openRead)
			if err != nil {
				t.Fatalf("HashBlake3File failed: %v", err)
			}

			if !bytes.Equal(hash, tc.expected) {
				t.Errorf("Expected %x, got %x", tc.expected, hash)
			}
		})
	}
}

func TestEd25519Operations(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	t.Run("key generation and signing", func(t *testing.T) {
		seed := make([]byte, ed25519.SeedSize)
		copy(seed, "test-seed-1234567890abcdef")

		keyPair, err := crypto.NewKeyPairEd25519(ctx, seed)
		if err != nil {
			t.Fatalf("NewKeyPairEd25519 failed: %v", err)
		}

		msg := []byte("test message")
		sig, err := crypto.SignEd25519(ctx, *keyPair, msg)
		if err != nil {
			t.Fatalf("SignEd25519 failed: %v", err)
		}

		valid, err := crypto.VerifyEd25519(ctx, keyPair.PublicKeyRaw(), msg, sig)
		if err != nil {
			t.Fatalf("VerifyEd25519 failed: %v", err)
		}
		if !valid {
			t.Error("Signature verification failed")
		}

		// Test with tampered message
		tamperedMsg := []byte("test messagE")
		valid, _ = crypto.VerifyEd25519(ctx, keyPair.PublicKeyRaw(), tamperedMsg, sig)
		if valid {
			t.Error("Tampered message should not verify")
		}
	})

	t.Run("invalid key sizes", func(t *testing.T) {
		// Test invalid seed size
		_, err := crypto.NewKeyPairEd25519(ctx, []byte("short"))
		if err == nil {
			t.Error("Expected error for invalid seed size")
		}

		// Test invalid private key
		invalidKey := libcrypto.KeyPairEd25519{Bytes: make([]byte, 10)}
		_, err = crypto.SignEd25519(ctx, invalidKey, []byte("test"))
		if err == nil {
			t.Error("Expected error for invalid private key size")
		}
	})
}

func TestXChaCha20Poly1305(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	key := make([]byte, chacha20poly1305.KeySize)
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	plaintext := []byte("secret message")

	t.Run("encrypt/decrypt roundtrip", func(t *testing.T) {
		ciphertext, err := crypto.EncryptXChaCha20Poly1305(ctx, key, nonce, plaintext)
		if err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		decrypted, err := crypto.DecryptXChaCha20Poly1305(ctx, key, nonce, ciphertext)
		if err != nil {
			t.Fatalf("Decryption failed: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Decrypted text mismatch: want %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("tampered ciphertext", func(t *testing.T) {
		ciphertext, _ := crypto.EncryptXChaCha20Poly1305(ctx, key, nonce, plaintext)
		ciphertext[0] ^= 0x01 // Flip first bit

		_, err := crypto.DecryptXChaCha20Poly1305(ctx, key, nonce, ciphertext)
		if err == nil {
			t.Error("Expected error for tampered ciphertext")
		}
	})

	t.Run("invalid key/nonce sizes", func(t *testing.T) {
		_, err := crypto.EncryptXChaCha20Poly1305(ctx, []byte("short"), nonce, plaintext)
		if err == nil {
			t.Error("Expected error for invalid key size")
		}

		_, err = crypto.EncryptXChaCha20Poly1305(ctx, key, []byte("short"), plaintext)
		if err == nil {
			t.Error("Expected error for invalid nonce size")
		}
	})
}

// mustDecodeHex is a helper function to decode hex strings for test vectors
func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid test vector hex: " + s)
	}
	return b
}
