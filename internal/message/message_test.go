package message

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageEncodedAndDecoded(t *testing.T) {
	msg := &Message{17, []byte("abracadabra")}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	CheckError(err, t)

	encryptedMsg, err := EncodeRSAGob(msg, privateKey.PublicKey)
	CheckError(err, t)

	msgDecrypted, err := DecodeRSAGob(encryptedMsg, *privateKey)
	CheckError(err, t)

	assert.Equal(t, msg, msgDecrypted)
}

func CheckError(e error, t *testing.T) {
	if e != nil {
		t.Fail()
	}
}
