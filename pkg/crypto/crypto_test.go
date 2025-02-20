package crypto_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	libcrypto "go.lumeweb.com/libs5-go/pkg/crypto"
	"golang.org/x/crypto/ed25519"
	"io"
	"math"
	"sync"
	"testing"
)

func TestGenerateSecureRandomBytes(t *testing.T) {
	crypto := libcrypto.NewDefaultCrypto()

	t.Run("valid lengths", func(t *testing.T) {
		testSizes := []int{1, 16, 32, 64, 1024, 1024 * 1024} // Test up to 1MB
		for _, size := range testSizes {
			t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
				b, err := crypto.GenerateSecureRandomBytes(size)
				if err != nil {
					t.Fatalf("Unexpected error for size %d: %v", size, err)
				}
				if len(b) != size {
					t.Errorf("Expected %d bytes, got %d", size, len(b))
				}
			})
		}
	})

	t.Run("invalid lengths", func(t *testing.T) {
		invalidSizes := []int{
			-1,               // Negative
			0,                // Zero
			-1 * (1 << 31),   // MinInt32
			math.MaxInt64,    // MaxInt64
			10*1024*1024 + 1, // Just over max allowed
		}

		for _, size := range invalidSizes {
			t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
				_, err := crypto.GenerateSecureRandomBytes(size)
				if err == nil {
					t.Errorf("Expected error for invalid size %d", size)
				}
			})
		}
	})

	t.Run("concurrency safety", func(t *testing.T) {
		const (
			numGoroutines   = 100
			bytesPerRoutine = 32
		)

		var wg sync.WaitGroup
		results := make(chan []byte, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				b, err := crypto.GenerateSecureRandomBytes(bytesPerRoutine)
				if err != nil {
					errors <- err
					return
				}
				results <- b
			}()
		}

		// Wait for all goroutines and close channels
		go func() {
			wg.Wait()
			close(results)
			close(errors)
		}()

		// Check for errors
		for err := range errors {
			t.Errorf("Goroutine error: %v", err)
		}

		// Check for duplicates
		seen := make(map[string]bool)
		for result := range results {
			if len(result) != bytesPerRoutine {
				t.Errorf("Expected %d bytes, got %d", bytesPerRoutine, len(result))
			}

			key := string(result)
			if seen[key] {
				t.Error("Duplicate random bytes generated")
			}
			seen[key] = true
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
			name:     "empty string",
			input:    []byte(""),
			expected: mustDecodeHex("af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"),
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: mustDecodeHex("d74981efa70a0c880b8d8c1985d075dbcbf679b99a5f9914e5aaf96b831a9e24"),
		},
		{
			name:     "1KB string",
			input:    bytes.Repeat([]byte("A"), 1024),
			expected: mustDecodeHex("f7314bcd4f08b945da46890d4abcbe9bd78905369461379ed5ab893eaccff236"),
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
				t.Errorf("Hash mismatch for %s:\nwant: %x\ngot:  %x",
					tc.name, tc.expected, hash)
			}

			// Test sync version
			syncHash, err := crypto.HashBlake3Sync(tc.input)
			if err != nil {
				t.Fatalf("HashBlake3Sync failed: %v", err)
			}
			if !bytes.Equal(syncHash, tc.expected) {
				t.Errorf("HashBlake3Sync mismatch for %s:\nwant: %x\ngot:  %x",
					tc.name, tc.expected, syncHash)
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

	// Use deterministic test vectors
	key := mustDecodeHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	nonce := mustDecodeHex("6465666768696a6b6c6d6e6f707172737475767778797a7b")

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "empty message",
			plaintext: []byte(""),
		},
		{
			name:      "short message",
			plaintext: []byte("Hello, XChaCha20!"),
		},
		{
			name:      "32-byte message",
			plaintext: bytes.Repeat([]byte("A"), 32),
		},
		{
			name:      "64-byte message",
			plaintext: bytes.Repeat([]byte("B"), 64),
		},
		{
			name:      "json message",
			plaintext: []byte(`{"user":"alice","timestamp":1234567890}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test encryption
			ciphertext, err := crypto.EncryptXChaCha20Poly1305(ctx, key, nonce, tc.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Test decryption
			decrypted, err := crypto.DecryptXChaCha20Poly1305(ctx, key, nonce, ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify roundtrip
			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Errorf("Decrypted text doesn't match original:\nwant: %x\ngot:  %x",
					tc.plaintext, decrypted)
			}

			// Test tampering detection
			if len(ciphertext) > 0 {
				tampered := make([]byte, len(ciphertext))
				copy(tampered, ciphertext)
				tampered[0] ^= 0x01 // Flip one bit

				_, err = crypto.DecryptXChaCha20Poly1305(ctx, key, nonce, tampered)
				if err == nil {
					t.Error("Expected error for tampered ciphertext")
				}
			}
		})
	}

	t.Run("invalid key size", func(t *testing.T) {
		invalidKey := make([]byte, 31) // Too short
		_, err := crypto.EncryptXChaCha20Poly1305(ctx, invalidKey, nonce, []byte("test"))
		if err == nil {
			t.Error("Expected error for invalid key size")
		}
	})

	t.Run("invalid nonce size", func(t *testing.T) {
		invalidNonce := make([]byte, 23) // Too short
		_, err := crypto.EncryptXChaCha20Poly1305(ctx, key, invalidNonce, []byte("test"))
		if err == nil {
			t.Error("Expected error for invalid nonce size")
		}
	})

	t.Run("concurrent encryption", func(t *testing.T) {
		const concurrency = 10
		var wg sync.WaitGroup
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func(i int) {
				defer wg.Done()

				// Use different nonces for each goroutine
				localNonce := make([]byte, len(nonce))
				copy(localNonce, nonce)
				localNonce[0] = byte(i) // Make nonce unique per goroutine

				msg := []byte(fmt.Sprintf("concurrent test %d", i))
				ciphertext, err := crypto.EncryptXChaCha20Poly1305(ctx, key, localNonce, msg)
				if err != nil {
					t.Errorf("Concurrent encryption %d failed: %v", i, err)
					return
				}

				decrypted, err := crypto.DecryptXChaCha20Poly1305(ctx, key, localNonce, ciphertext)
				if err != nil {
					t.Errorf("Concurrent decryption %d failed: %v", i, err)
					return
				}

				if !bytes.Equal(decrypted, msg) {
					t.Errorf("Concurrent test %d: decrypted text doesn't match original", i)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := crypto.EncryptXChaCha20Poly1305(ctx, key, nonce, []byte("test"))
		if err != context.Canceled {
			t.Errorf("Expected context canceled error, got %v", err)
		}

		_, err = crypto.DecryptXChaCha20Poly1305(ctx, key, nonce, []byte("test"))
		if err != context.Canceled {
			t.Errorf("Expected context canceled error, got %v", err)
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
