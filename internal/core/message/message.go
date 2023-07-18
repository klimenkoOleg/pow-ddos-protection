package message

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"pow-ddos-protection/internal/core/encryption"
)

// Header of TCP-message in protocol, means type of message
const (
	Step1ChallengeRequest     = iota // from client to server - request new challenge from server
	Step2TwoChallengeResponse        // from server to client - message with challenge for client
	Step3QuoteRequest                // from client to server - message with solved challenge
	Step4QuoteResponse               // from server to client - message with useful info is solution is correct, or with error if not
)

const (
	encryptionLabel = "OAEP Encrypted"
)

// Message - data wrapper for both client and server
type Message struct {
	Header  int    //	step number, see the constants above
	Payload []byte // 	binary payload, encrypted by encryption and serialized by gob
}

// Encode - RSA encode message to send it by tcp-connection
func EncodeRSAGob(m *Message, publicKey rsa.PublicKey) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff) // Will write to buffer.

	err := enc.Encode(&m)
	if err != nil {
		return nil, err
	}

	return rsa_oaep_encrypt(buff.Bytes(), publicKey)
}

// DecodeRSAGob - RSA decode message to get it by tcp-connection
func DecodeRSAGob(encrypted []byte, privateKey rsa.PrivateKey) (*Message, error) {
	decryptedBytes, err := rsa_oaep_decrypt(encrypted, privateKey)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	_, err = buff.Write(decryptedBytes)
	if err != nil {
		return nil, err
	}

	message := &Message{}
	enc := gob.NewDecoder(&buff) // Will read from buffer.

	err = enc.Decode(&message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func rsa_oaep_encrypt(secretMessage []byte, key rsa.PublicKey) ([]byte, error) {
	rng := rand.Reader
	//ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), []byte(encryptionLabel))
	ciphertext, err := encryption.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), []byte(encryptionLabel))
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func rsa_oaep_decrypt(cipherText []byte, privKey rsa.PrivateKey) ([]byte, error) {
	rng := rand.Reader
	//decryptedMessage, err := rsa.DecryptOAEP(sha256.New(), rng, &privKey, cipherText, []byte(encryptionLabel))
	decryptedMessage, err := encryption.DecryptOAEP(sha256.New(), rng, &privKey, cipherText, []byte(encryptionLabel))
	if err != nil {
		return nil, err
	}

	return decryptedMessage, nil
}
