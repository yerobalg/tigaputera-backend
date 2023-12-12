package cryptolib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type cryptoLib struct {
	secretKey string
}

type Interface interface {
	Encrypt(plaintext string) string
	Decrypt(string) string
}

func Init(secretKey string) Interface {
	return &cryptoLib{
		secretKey: secretKey,
	}
}

func (c *cryptoLib) Encrypt(plaintext string) string {
	aes, err := aes.NewCipher([]byte(c.secretKey))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return string(ciphertext)
}

func (c *cryptoLib) Decrypt(ciphertext string) string {
	aes, err := aes.NewCipher([]byte(c.secretKey))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		panic(err)
	}

	return string(plaintext)
}
