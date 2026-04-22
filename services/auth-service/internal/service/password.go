package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	
	return salt, err
}

func HashPassword(password string) (string, error){
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}
	
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		1,
		64 * 1024,
		4,
		32,
	)
	
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	
	final := b64Salt + "$" + b64Hash
	
	return final, nil
}

func VerifyPassword(stored, password string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		1,
		64*1024,
		4,
		32,
	)

	return subtle.ConstantTimeCompare(hash, newHash) == 1
}