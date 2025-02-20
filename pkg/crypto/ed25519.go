package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"go.lumeweb.com/libs5-go/old/types"
	"go.lumeweb.com/libs5-go/pkg/bytes"
)

type KeyPairEd25519 struct {
	Bytes []byte
}

// GenerateEd25519Key creates a new Ed25519 key pair using secure randomness
func GenerateEd25519Key() (*KeyPairEd25519, error) {
	_, privateKey, err := ed25519.GenerateKey(frand.Reader)
	if err != nil {
		return nil, err
	}

	// The private key contains both the private and public key
	return New(privateKey), nil
}

// GenerateSecureRandomBytes generates cryptographically secure random bytes
func GenerateSecureRandomBytes(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func New(bytes []byte) *KeyPairEd25519 {
	return &KeyPairEd25519{Bytes: bytes}
}

func (kp *KeyPairEd25519) PublicKey() []byte {
	return bytes.ConcatBytes([]byte{byte(types.HashTypeEd25519)}, kp.PublicKeyRaw())
}

func (kp *KeyPairEd25519) PublicKeyRaw() []byte {
	publicKey := ed25519.PrivateKey(kp.Bytes).Public()

	return publicKey.(ed25519.PublicKey)
}

func (kp *KeyPairEd25519) ExtractBytes() []byte {
	return kp.Bytes
}
