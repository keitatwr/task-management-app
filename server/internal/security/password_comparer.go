package security

import "golang.org/x/crypto/bcrypt"

type PasswordComparer interface {
	ComparePassword(hashedPassword, password string) error
}

type BcryptPasswordComparer struct{}

func (c *BcryptPasswordComparer) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
