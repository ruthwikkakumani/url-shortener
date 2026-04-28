package auth

import (
	"os"
)

func GetJWTKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}
	return []byte(secret)
}