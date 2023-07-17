package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {

	secretMessage := "PoW sucks, use PoS"
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	CheckError(t, err)

	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	hash := sha256.New()
	encrypted, err := EncryptOAEP(hash, rng, &privateKey.PublicKey, []byte(secretMessage), label)

	decrypted, err := DecryptOAEP(hash, rng, privateKey, encrypted, label)

	assert.True(t, string(decrypted) == secretMessage)
}

func CheckError(t *testing.T, e error) {
	if e != nil {
		t.Fail()
	}
}
