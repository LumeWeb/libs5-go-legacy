package crypto

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"go.lumeweb.com/libs5-go/pkg/crypto/internal/test"
	"testing"
)

func TestEd25519Operations(t *testing.T) {
	crypto := NewDefaultCrypto()
	ctx := context.Background()

	// Use a deterministic test seed
	seed := test.MustDecodeHex("008a2a3ca07c46de4bf3f5e9eb79c8b3f31279f3ad0ab3ab40051a561f976c90")

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
				signature: test.MustDecodeHex("9cd1b3f61928156a5365102269f0dd1aaa164162218467ad988619452d65e2696e9cc0906ca7961237390096281856bbbc5b8fc39cfef26ff6e661f370900338"),
			},
			{
				name:      "short message",
				message:   []byte("Hello, Ed25519!"),
				signature: test.MustDecodeHex("56976cbfbad2b1fe8b5b88e983c445ce66b2c33d290f527bd55ea96d1750c2d97b8028c70eec4693a0a53dda0e5fa691ee140aaa7bc3453e479dc366cf9bb35a"),
			},
			{
				name:      "longer message",
				message:   []byte("A longer message that should be signed with Ed25519"),
				signature: test.MustDecodeHex("f95dd2be564789acb2450a886ed34812677f5d686e49015333cd6450297a23ba4e6495d6b4f83019f3de3dbc552444dd6927e8ddbf9fae5c9f954b7cf25f8b17"),
			},
			{
				name:      "1KB message",
				message:   bytes.Repeat([]byte("A"), 1024),
				signature: test.MustDecodeHex("501cf7e377f4bafaeeed633e20d6a9e4c3f11f0c00a74ad126eed16a3f020c1124c33bdba5c4210ab1f636acfa477127d204be2fffff72b0886045b8e6523a21"),
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
		invalidKey := KeyPairEd25519{Bytes: make([]byte, 10)}
		_, err = crypto.SignEd25519(ctx, invalidKey, []byte("test"))
		if err == nil {
			t.Error("Expected error for invalid private key size")
		}
	})
}
