package crypto

import (
	"context"
	"crypto/ed25519"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/chacha20poly1305"
)

// DefaultCryptoImplementation provides production-ready crypto operations
type DefaultCryptoImplementation struct {
	mu sync.Mutex
}

func NewDefaultCrypto() *DefaultCryptoImplementation {
	return &DefaultCryptoImplementation{}
}

func (d *DefaultCryptoImplementation) GenerateSecureRandomBytes(length int) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return buf, nil
}

func (d *DefaultCryptoImplementation) HashBlake3(ctx context.Context, input []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		hash := blake3.New()
		_, err := hash.Write(input)
		if err != nil {
			return nil, fmt.Errorf("blake3 hash failed: %w", err)
		}
		return hash.Sum(nil), nil
	}
}

func (d *DefaultCryptoImplementation) HashBlake3Sync(input []byte) ([]byte, error) {
	hash := blake3.New()
	_, err := hash.Write(input)
	if err != nil {
		return nil, fmt.Errorf("blake3 hash failed: %w", err)
	}
	return hash.Sum(nil), nil
}

func (d *DefaultCryptoImplementation) HashBlake3File(ctx context.Context, size int64, openRead OpenReadFunction) ([]byte, error) {
	hash := blake3.New()
	
	for offset := int64(0); offset < size; offset += 1<<20 { // 1MB chunks
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			reader, err := openRead(offset, offset+1<<20)
			if err != nil {
				return nil, fmt.Errorf("failed to open read at offset %d: %w", offset, err)
			}

			_, err = io.CopyN(hash, reader, 1<<20)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return nil, fmt.Errorf("read failed at offset %d: %w", offset, err)
			}
		}
	}
	
	return hash.Sum(nil), nil
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
		if len(keyPair.PrivateKey) != ed25519.PrivateKeySize {
			return nil, errors.New("invalid private key length")
		}
		return ed25519.Sign(keyPair.PrivateKey, message), nil
	}
}

func (d *DefaultCryptoImplementation) NewKeyPairEd25519(ctx context.Context, seed []byte) (KeyPairEd25519, error) {
	select {
	case <-ctx.Done():
		return KeyPairEd25519{}, ctx.Err()
	default:
		if len(seed) != ed25519.SeedSize {
			return KeyPairEd25519{}, errors.New("invalid seed length")
		}
		
		privateKey := ed25519.NewKeyFromSeed(seed)
		return KeyPairEd25519{
			PublicKey:  privateKey[32:],
			PrivateKey: privateKey,
		}, nil
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
