package backend

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"io"
	"strings"
	"time"
)

type rsaPgpProvider struct {
}

func NewRsaPgpProvider() RsaProvider {
	return rsaPgpProvider{}
}

func (r rsaPgpProvider) GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}

	return privkey, &privkey.PublicKey
}

func (r rsaPgpProvider) PrivateKeyToBytes(key *rsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, openpgp.PrivateKeyType, make(map[string]string))
	if err != nil {
		panic(err)
	}
	pgpKey := packet.NewRSAPrivateKey(time.Now(), key)
	err = pgpKey.Serialize(w)
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (r rsaPgpProvider) PublicKeyToBytes(key *rsa.PublicKey) []byte {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, openpgp.PublicKeyType, make(map[string]string))
	if err != nil {
		panic(err)
	}
	pgpKey := packet.NewRSAPublicKey(time.Now(), key)
	err = pgpKey.Serialize(w)
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (r rsaPgpProvider) BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	panic("not implemented")
}

func (r rsaPgpProvider) BytesToPrivateKeyPacket(priv []byte) *packet.PrivateKey {
	in := bytes.NewReader(priv)
	block, err := armor.Decode(in)
	if err != nil {
		panic(err)
	}

	if block.Type != openpgp.PrivateKeyType {
		panic("Invalid private key type")
	}

	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		panic(err)
	}

	key, ok := pkt.(*packet.PrivateKey)
	if !ok {
		panic("Invalid private key")
	}
	return key
}

func (r rsaPgpProvider) BytesToPublicKey(pub []byte) *rsa.PublicKey {
	panic("not implemented")
}

func (r rsaPgpProvider) BytesToPublicKeyPacket(pub []byte) *packet.PublicKey {

	in := bytes.NewReader(pub)

	block, err := armor.Decode(in)
	if err != nil {
		panic(err)
	}

	if block.Type != openpgp.PublicKeyType {
		panic("Invalid public key type")
	}

	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		panic(err)
	}

	key, ok := pkt.(*packet.PublicKey)
	if !ok {
		panic("Invalid public key")
	}
	return key
}

func (r rsaPgpProvider) EncryptWithPublicKey(msg []byte, pubKey []byte) []byte {

	buf := new(bytes.Buffer)

	entity := r.createEntityFromKeys(r.BytesToPublicKeyPacket(pubKey), nil)

	w, err := armor.Encode(buf, "ENCRYPTED", make(map[string]string))
	if err != nil {
		panic(err)
	}

	plain, err := openpgp.Encrypt(w, []*openpgp.Entity{entity}, nil, nil, nil)
	if err != nil {
		panic(err)
	}

	_, err = plain.Write(msg)
	if err != nil {
		panic(err)
	}

	err = plain.Close()
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (r rsaPgpProvider) DecryptWithPrivateKey(ciphertext []byte, privKey []byte) []byte {

	privPacket := r.BytesToPrivateKeyPacket(privKey)

	entity := r.createEntityFromKeys(&privPacket.PublicKey, privPacket)

	block, err := armor.Decode(bytes.NewReader(ciphertext))
	if err != nil {
		panic(err)
	}

	if block.Type != "ENCRYPTED" {
		panic("Invalid message type")
	}

	var entityList openpgp.EntityList
	entityList = append(entityList, entity)

	md, err := openpgp.ReadMessage(block.Body, entityList, nil, nil)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, md.UnverifiedBody)
	if err != nil {
		panic(err)
	}

	return []byte(buf.String())
}

func (r rsaPgpProvider) createEntityFromKeys(pubKey *packet.PublicKey, privKey *packet.PrivateKey) *openpgp.Entity {

	bits := 0
	if pubKey != nil {
		bits = pubKey.PublicKey.(*rsa.PublicKey).N.BitLen()
	}

	if privKey != nil {
		bits = privKey.PublicKey.PublicKey.(*rsa.PublicKey).N.BitLen()
	}

	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: 9,
		},
		RSABits: bits,
	}
	currentTime := config.Now()
	uid := packet.NewUserId("", "", "")

	e := openpgp.Entity{
		PrimaryKey: pubKey,
		PrivateKey: privKey,
		Identities: make(map[string]*openpgp.Identity),
	}
	isPrimaryId := false

	e.Identities[uid.Id] = &openpgp.Identity{
		Name:   uid.Name,
		UserId: uid,
		SelfSignature: &packet.Signature{
			CreationTime: currentTime,
			SigType:      packet.SigTypePositiveCert,
			PubKeyAlgo:   packet.PubKeyAlgoRSA,
			Hash:         config.Hash(),
			IsPrimaryId:  &isPrimaryId,
			FlagsValid:   true,
			FlagSign:     true,
			FlagCertify:  true,
			IssuerKeyId:  &e.PrimaryKey.KeyId,
		},
	}

	keyLifetimeSecs := uint32(86400 * 365)

	e.Subkeys = make([]openpgp.Subkey, 1)
	e.Subkeys[0] = openpgp.Subkey{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Sig: &packet.Signature{
			CreationTime:              currentTime,
			SigType:                   packet.SigTypeSubkeyBinding,
			PubKeyAlgo:                packet.PubKeyAlgoRSA,
			Hash:                      config.Hash(),
			PreferredHash:             []uint8{8}, // SHA-256
			FlagsValid:                true,
			FlagEncryptStorage:        true,
			FlagEncryptCommunications: true,
			IssuerKeyId:               &e.PrimaryKey.KeyId,
			KeyLifetimeSecs:           &keyLifetimeSecs,
		},
	}
	return &e
}
