package crypto

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/fs"
)

// CryptoImplementation defines cryptographic operations
type CryptoImplementation interface {
	// GenerateSecureRandomBytes generates secure random bytes
	GenerateSecureRandomBytes(length int) ([]byte, error)

	// HashBlake3 hashes input using BLAKE3
	HashBlake3(ctx context.Context, input []byte) ([]byte, error)

	// HashBlake3Sync synchronously hashes input using BLAKE3
	HashBlake3Sync(input []byte) ([]byte, error)

	// HashBlake3File hashes a file using BLAKE3
	HashBlake3File(ctx context.Context, size int64, openRead fs.OpenReadFunction) ([]byte, error)

	// VerifyEd25519 verifies an Ed25519 signature
	VerifyEd25519(ctx context.Context, publicKey, message, signature []byte) (bool, error)

	// SignEd25519 creates an Ed25519 signature
	SignEd25519(ctx context.Context, keyPair KeyPairEd25519, message []byte) ([]byte, error)

	// NewKeyPairEd25519 creates a new Ed25519 key pair from a seed
	NewKeyPairEd25519(ctx context.Context, seed []byte) (*KeyPairEd25519, error)

	// EncryptXChaCha20Poly1305 encrypts data using XChaCha20-Poly1305
	EncryptXChaCha20Poly1305(ctx context.Context, key, nonce, plaintext []byte) ([]byte, error)

	// DecryptXChaCha20Poly1305 decrypts data using XChaCha20-Poly1305
	DecryptXChaCha20Poly1305(ctx context.Context, key, nonce, ciphertext []byte) ([]byte, error)
}
