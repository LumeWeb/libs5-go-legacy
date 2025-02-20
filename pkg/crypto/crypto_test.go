package crypto_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	libcrypto "go.lumeweb.com/libs5-go/pkg/crypto"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"io"
	"testing"
)

func TestGenerateSecureRandomBytes(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()

	// Test vectors for specific sizes
	t.Run("specific sizes", func(t *testing.T) {
		sizes := []int{16, 32, 64, 128}
		for _, size := range sizes {
			b, err := crypto.GenerateSecureRandomBytes(size)
			if err != nil {
				t.Fatalf("Unexpected error for size %d: %v", size, err)
			}
			if len(b) != size {
				t.Errorf("Expected %d bytes, got %d", size, len(b))
			}
		}
	})

	// Test for uniqueness using real random values
	t.Run("uniqueness test", func(t *testing.T) {
		knownRandoms := []string{
			"5f11b3f99fffa91841cfc4a5d955b672",                                 // 16 bytes
			"c45edc738e3ddd9e8d82ea14cafaae3af5967df91ec47b6475165a28993d4992", // 32 bytes
		}

		// Generate new random and ensure it's not in our known set
		b, err := crypto.GenerateSecureRandomBytes(16)
		if err != nil {
			t.Fatal(err)
		}

		hexResult := hex.EncodeToString(b)
		for _, known := range knownRandoms {
			if hexResult == known {
				t.Error("Generated random bytes matched a known value - this should be statistically impossible")
			}
		}
	})

	t.Run("invalid sizes", func(t *testing.T) {
		invalidSizes := []int{0, -1, -32}
		for _, size := range invalidSizes {
			_, err := crypto.GenerateSecureRandomBytes(size)
			if err == nil {
				t.Errorf("Expected error for invalid size %d", size)
			}
		}
	})
}

func TestHashBlake3(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	testCases := []struct {
		name  string
		input []byte
		// Note: These are SHA-256 hashes since we couldn't generate BLAKE3 in JS,
		// but they demonstrate the structure. Replace with actual BLAKE3 hashes.
		expectedHash string
	}{
		{
			name:         "empty string",
			input:        []byte(""),
			expectedHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:         "hello world",
			input:        []byte("hello world"),
			expectedHash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:         "1KB string",
			input:        bytes.Repeat([]byte("A"), 1024),
			expectedHash: "c8cd55c4c6374e72d3a52197809e8410fd0573635444b2e489e83cdc924ed384",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := crypto.HashBlake3(ctx, tc.input)
			if err != nil {
				t.Fatalf("HashBlake3 failed: %v", err)
			}

			if hexHash := hex.EncodeToString(hash); hexHash != tc.expectedHash {
				t.Errorf("Hash mismatch for %s:\nwant: %s\ngot:  %s",
					tc.name, tc.expectedHash, hexHash)
			}
		})
	}
}

func TestHashBlake3File(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	testCases := []struct {
		name     string
		size     int64
		data     []byte
		expected []byte // Replace with actual BLAKE3 hashes
	}{
		{
			name:     "empty file",
			size:     0,
			data:     []byte{},
			expected: mustDecodeHex("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
		},
		{
			name:     "1KB file",
			size:     1024,
			data:     bytes.Repeat([]byte("test-pattern-123"), 1024/len("test-pattern-123")+1)[:1024],
			expected: mustDecodeHex("f400267ceb96cbbc0aacab1302f3646f9263c7c346532edb4ad259485fd2a75c"),
		},
		{
			name:     "1MB file",
			size:     1024 * 1024,
			data:     bytes.Repeat([]byte("test-pattern-456"), 1024*1024/len("test-pattern-456")+1)[:1024*1024],
			expected: mustDecodeHex("e431a24d35079ad68c26c793de2769aada4f6e096cc76d2b9e11b3a898a6c04c"),
		},
		{
			name:     "2MB file",
			size:     2 * 1024 * 1024,
			data:     bytes.Repeat([]byte("test-pattern-789"), 2*1024*1024/len("test-pattern-789")+1)[:2*1024*1024],
			expected: mustDecodeHex("f63560e1ae5256836977fcc835c32a6914b4ae71ae9c91ff59fd285746b0013c"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a reader function that returns chunks of the test data
			openRead := func(start, end int) (io.Reader, error) {
				if start >= len(tc.data) {
					return bytes.NewReader(nil), nil
				}
				if end > len(tc.data) {
					end = len(tc.data)
				}
				return bytes.NewReader(tc.data[start:end]), nil
			}

			hash, err := crypto.HashBlake3File(ctx, tc.size, openRead)
			if err != nil {
				t.Fatalf("HashBlake3File failed: %v", err)
			}

			if !bytes.Equal(hash, tc.expected) {
				t.Errorf("Hash mismatch for %s:\nwant: %x\ngot:  %x",
					tc.name, tc.expected, hash)
			}
		})
	}

	t.Run("invalid reader", func(t *testing.T) {
		openRead := func(start, end int) (io.Reader, error) {
			return nil, fmt.Errorf("simulated read error")
		}

		_, err := crypto.HashBlake3File(ctx, 1024, openRead)
		if err == nil {
			t.Error("Expected error from invalid reader")
		}
	})
}

func TestEd25519Operations(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()
	ctx := context.Background()

	// Use a deterministic test seed
	seed := mustDecodeHex("008a2a3ca07c46de4bf3f5e9eb79c8b3f31279f3ad0ab3ab40051a561f976c90")

	t.Run("key generation and signing", func(t *testing.T) {
		keyPair, err := crypto.NewKeyPairEd25519(ctx, seed)
		if err != nil {
			t.Fatalf("NewKeyPairEd25519 failed: %v", err)
		}

		testCases := []struct {
			name    string
			message []byte
			// These signatures are from ECDSA P-256 (since Ed25519 wasn't available in browser)
			// Replace with actual Ed25519 signatures in production
			signature []byte
		}{
			{
				name:      "empty message",
				message:   []byte(""),
				signature: mustDecodeHex("9cd1b3f61928156a5365102269f0dd1aaa164162218467ad988619452d65e2696e9cc0906ca7961237390096281856bbbc5b8fc39cfef26ff6e661f370900338"),
			},
			{
				name:      "short message",
				message:   []byte("Hello, Ed25519!"),
				signature: mustDecodeHex("56976cbfbad2b1fe8b5b88e983c445ce66b2c33d290f527bd55ea96d1750c2d97b8028c70eec4693a0a53dda0e5fa691ee140aaa7bc3453e479dc366cf9bb35a"),
			},
			{
				name:      "longer message",
				message:   []byte("A longer message that should be signed with Ed25519"),
				signature: mustDecodeHex("f95dd2be564789acb2450a886ed34812677f5d686e49015333cd6450297a23ba4e6495d6b4f83019f3de3dbc552444dd6927e8ddbf9fae5c9f954b7cf25f8b17"),
			},
			{
				name:      "1KB message",
				message:   bytes.Repeat([]byte("A"), 1024),
				signature: mustDecodeHex("501cf7e377f4bafaeeed633e20d6a9e4c3f11f0c00a74ad126eed16a3f020c1124c33bdba5c4210ab1f636acfa477127d204be2fffff72b0886045b8e6523a21"),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test signing
				sig, err := crypto.SignEd25519(ctx, *keyPair, tc.message)
				if err != nil {
					t.Fatalf("SignEd25519 failed: %v", err)
				}

				// Test verification
				valid, err := crypto.VerifyEd25519(ctx, keyPair.PublicKeyRaw(), tc.message, sig)
				if err != nil {
					t.Fatalf("VerifyEd25519 failed: %v", err)
				}
				if !valid {
					t.Error("Signature verification failed")
				}

				// Test verification with tampered message
				if len(tc.message) > 0 {
					tamperedMsg := make([]byte, len(tc.message))
					copy(tamperedMsg, tc.message)
					tamperedMsg[0] ^= 0x01 // Flip one bit

					valid, err = crypto.VerifyEd25519(ctx, keyPair.PublicKeyRaw(), tamperedMsg, sig)
					if err != nil {
						t.Fatalf("VerifyEd25519 with tampered message failed: %v", err)
					}
					if valid {
						t.Error("Signature verified for tampered message")
					}
				}
			})
		}
	})

	t.Run("invalid key sizes", func(t *testing.T) {
		// Test with invalid seed size
		invalidSeed := make([]byte, ed25519.SeedSize-1)
		_, err := crypto.NewKeyPairEd25519(ctx, invalidSeed)
		if err == nil {
			t.Error("Expected error for invalid seed size")
		}

		// Test with invalid private key
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
