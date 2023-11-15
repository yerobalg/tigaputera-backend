package password

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type passwordLib struct{}

type Interface interface {
	Hash(string) (string, error)
	Compare(string, string) bool
}

func Init() Interface {
	return &passwordLib{}
}

func (p *passwordLib) Hash(password string) (string, error) {
	saltRound, err := strconv.Atoi(os.Getenv("BCRYPT_SALT_ROUND"))
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), saltRound)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *passwordLib) Compare(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
