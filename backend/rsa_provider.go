package backend

import (
	"crypto/rsa"
	"golang.org/x/crypto/openpgp/packet"
)

type RsaProvider interface {
	GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey)
	PrivateKeyToBytes(priv *rsa.PrivateKey) []byte
	PublicKeyToBytes(pub *rsa.PublicKey) []byte
	BytesToPrivateKey(priv []byte) *rsa.PrivateKey
	BytesToPrivateKeyPacket(priv []byte) *packet.PrivateKey
	BytesToPublicKey(pub []byte) *rsa.PublicKey
	BytesToPublicKeyPacket(pub []byte) *packet.PublicKey
	EncryptWithPublicKey(msg []byte, pubKey []byte) []byte
	DecryptWithPrivateKey(ciphertext []byte, privKey []byte) []byte
}
