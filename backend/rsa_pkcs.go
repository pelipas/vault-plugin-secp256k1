package backend

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/openpgp/packet"
	"log"
)

type rsaPkcsProvider struct{}

func (p rsaPkcsProvider) BytesToPrivateKeyPacket(priv []byte) *packet.PrivateKey {
	panic("not implemented")
}

func (p rsaPkcsProvider) BytesToPublicKeyPacket(pub []byte) *packet.PublicKey {
	panic("not implemented")
}

func NewRsaPkcsProvider() RsaProvider {
	return rsaPkcsProvider{}
}

// GenerateKeyPair generates a new key pair
func (p rsaPkcsProvider) GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	return privkey, &privkey.PublicKey
}

// PrivateKeyToBytes private key to bytes
func (p rsaPkcsProvider) PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// PublicKeyToBytes public key to bytes
func (p rsaPkcsProvider) PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		panic(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

// BytesToPrivateKey bytes to private key
func (p rsaPkcsProvider) BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			panic(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		panic(err)
	}
	return key
}

// BytesToPublicKey bytes to public key
func (p rsaPkcsProvider) BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			panic(err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		panic(err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil
	}
	return key
}

// EncryptWithPublicKey encrypts data with public key
func (p rsaPkcsProvider) EncryptWithPublicKey(msg []byte, pubKey []byte) []byte {
	pub := p.BytesToPublicKey(pubKey)
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func (p rsaPkcsProvider) DecryptWithPrivateKey(ciphertext []byte, privKey []byte) []byte {
	priv := p.BytesToPrivateKey(privKey)
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}
