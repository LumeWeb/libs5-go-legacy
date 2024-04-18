package ed25519

import (
	"crypto/ed25519"
	"github.com/LumeWeb/libs5-go/types"
	"github.com/LumeWeb/libs5-go/utils"
)

type KeyPairEd25519 struct {
	Bytes []byte
}

func New(bytes []byte) *KeyPairEd25519 {
	return &KeyPairEd25519{Bytes: bytes}
}

func (kp *KeyPairEd25519) PublicKey() []byte {
	return utils.ConcatBytes([]byte{byte(types.HashTypeEd25519)}, kp.PublicKeyRaw())
}

func (kp *KeyPairEd25519) PublicKeyRaw() []byte {
	publicKey := ed25519.PrivateKey(kp.Bytes).Public()

	return publicKey.(ed25519.PublicKey)
}

func (kp *KeyPairEd25519) ExtractBytes() []byte {
	return kp.Bytes
}
