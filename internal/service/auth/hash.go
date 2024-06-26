package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func generateHash(password string) (string, error) {
	byts, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(byts), err
}

func compareHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
