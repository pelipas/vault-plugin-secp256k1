package backend

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPkcsProvider(t *testing.T) {

	p := NewRsaPkcsProvider()

	privateKey, publicKey := p.GenerateKeyPair(2048)
	privateKeyBytes := p.PrivateKeyToBytes(privateKey)
	publicKeyKeyBytes := p.PublicKeyToBytes(publicKey)

	privateKeyFromBytes := p.BytesToPrivateKey(privateKeyBytes)
	publicKeyKeyFromBytes := p.BytesToPublicKey(publicKeyKeyBytes)

	assert.True(t, privateKey.Equal(privateKeyFromBytes))
	assert.True(t, publicKey.Equal(publicKeyKeyFromBytes))

	msg := []byte("msg")
	encrypted := p.EncryptWithPublicKey(msg, publicKeyKeyBytes)
	decrypted := p.DecryptWithPrivateKey(encrypted, privateKeyBytes)
	assert.Equal(t, decrypted, msg)
}
