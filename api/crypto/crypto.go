package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func (*Service) Hash(data string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (*Service) CompareHashAndData(hash string, data string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(data))
}
