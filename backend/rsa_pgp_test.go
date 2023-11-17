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

func TestWithToolGpg(t *testing.T) {

	t.Skip() // закомментировать при отладке

	// генерируем rsa ключ: gpg --full-generate-key
	// получаем список ключей: gpg --list-keys

	public := `` // вставить свой ключ из gpg (gpg --output public.pgp --armor --export USERNAME)

	p := NewRsaPgpProvider()

	msg := []byte("some_private_data") // данные для шифровки

	encrypted := p.EncryptWithPublicKey(msg, []byte(public))
	b64 := encodeBase64((encrypted))
	t.Log(b64) // положить в файл и декодировать с помощью: cat BASE64_ENCRYPTED_KEY_FILE | base64 --decode | gpg -d

}
