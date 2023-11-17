package backend

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPgpProvider(t *testing.T) {

	p := NewRsaPgpProvider()
	privateKey, publicKey := p.GenerateKeyPair(2048)

	privateKeyBytes := p.PrivateKeyToBytes(privateKey)
	publicKeyBytes := p.PublicKeyToBytes(publicKey)

	privateKeyFromBytes := p.BytesToPrivateKeyPacket(privateKeyBytes)
	publicKeyFromBytes := p.BytesToPublicKeyPacket(publicKeyBytes)

	assert.True(t, publicKey.Equal(publicKeyFromBytes.PublicKey))
	assert.False(t, privateKey.Equal(privateKeyFromBytes)) // check creation time different

	msg := []byte("msg")

	encrypted := p.EncryptWithPublicKey(msg, publicKeyBytes)
	t.Log(string(encrypted))
	decrypted := p.DecryptWithPrivateKey(encrypted, privateKeyBytes)
	t.Log(string(decrypted))

	assert.Equal(t, msg, decrypted)
}
