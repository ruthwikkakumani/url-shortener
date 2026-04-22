package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/config"
)

func getJWTKey() []byte {
	return []byte(config.GetEnv("JWT_SECRET", ""))
}

func generateJWT(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(14 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	return token.SignedString(getJWTKey())
}