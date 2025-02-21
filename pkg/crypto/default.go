package crypto

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"go.lumeweb.com/libs5-go/pkg/fs"
	"io"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
	"lukechampine.com/blake3"
	"lukechampine.com/frand"
)

// DefaultCryptoImplementation provides production-ready crypto operations
type DefaultCryptoImplementation struct {
	mu sync.Mutex
}

const mbChunk = 1 << 20
const maxLength = 10 * 1024 * 1024

var _ CryptoImplementation = (*DefaultCryptoImplementation)(nil)

func NewDefaultCrypto() *DefaultCryptoImplementation {
	return &DefaultCryptoImplementation{}
}

func (d *DefaultCryptoImplementation) GenerateSecureRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("invalid length: must be positive")
	}

	// Prevent integer overflow and unreasonable allocations
	// 10MB is a reasonable upper limit for random bytes
	if length > maxLength {
		return nil, fmt.Errorf("length %d exceeds maximum allowed size of %d bytes", length, maxLength)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	buf := make([]byte, length)
	_, err := frand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return buf, nil
}

// HashBlake3 computes a BLAKE3 hash and truncates to 256 bits
func (d *DefaultCryptoImplementation) HashBlake3(ctx context.Context, input []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		hash := blake3.New(64, nil)
		_, err := hash.Write(input)
		if err != nil {
			return nil, fmt.Errorf("blake3 hash failed: %w", err)
		}
		fullHash := hash.Sum(nil)
		// Truncate to 256 bits (32 bytes) following reference implementation pattern
		return fullHash[:32], nil
	}
}

// HashBlake3Sync calls the async version with a background context
func (d *DefaultCryptoImplementation) HashBlake3Sync(input []byte) ([]byte, error) {
	// Create a background context that will never be canceled
	ctx := context.Background()

	// Call the async method which handles the truncation
	return d.HashBlake3(ctx, input)
}

func (d *DefaultCryptoImplementation) HashBlake3File(ctx context.Context, size int64, openRead fs.OpenReadFunction) ([]byte, error) {
	hash := blake3.New(64, nil)

	for offset := int64(0); offset < size; offset += mbChunk { // 1MB chunks
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			reader, err := openRead(int(offset), mbChunk)
			if err != nil {
				return nil, fmt.Errorf("failed to open read at offset %d: %w", offset, err)
			}

			_, err = io.CopyN(hash, reader, mbChunk)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return nil, fmt.Errorf("read failed at offset %d: %w", offset, err)
			}
		}
	}

	fullHash := hash.Sum(nil)
	// Truncate to 256 bits (32 bytes) following reference implementation pattern
	return fullHash[:32], nil
}

func (d *DefaultCryptoImplementation) VerifyEd25519(ctx context.Context, publicKey, message, signature []byte) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		if len(publicKey) != ed25519.PublicKeySize {
			return false, errors.New("invalid public key length")
		}
		return ed25519.Verify(publicKey, message, signature), nil
	}
}

func (d *DefaultCryptoImplementation) SignEd25519(ctx context.Context, keyPair KeyPairEd25519, message []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if len(keyPair.ExtractBytes()) != ed25519.PrivateKeySize {
			return nil, errors.New("invalid private key length")
		}
		return ed25519.Sign(keyPair.ExtractBytes(), message), nil
	}
}

func (d *DefaultCryptoImplementation) NewKeyPairEd25519(ctx context.Context, seed []byte) (*KeyPairEd25519, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if len(seed) != ed25519.SeedSize {
			return nil, errors.New("invalid seed length")
		}

		return NewKeyFromSeed(seed), nil
	}
}

func (d *DefaultCryptoImplementation) EncryptXChaCha20Poly1305(ctx context.Context, key, nonce, plaintext []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if len(nonce) != chacha20poly1305.NonceSizeX {
			return nil, errors.New("invalid nonce size")
		}

		aead, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create AEAD: %w", err)
		}

		return aead.Seal(nil, nonce, plaintext, nil), nil
	}
}

func (d *DefaultCryptoImplementation) DecryptXChaCha20Poly1305(ctx context.Context, key, nonce, ciphertext []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if len(nonce) != chacha20poly1305.NonceSizeX {
			return nil, errors.New("invalid nonce size")
		}

		aead, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create AEAD: %w", err)
		}

		return aead.Open(nil, nonce, ciphertext, nil)
	}
}
