package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/config"
)

func getJWTKey() []byte {
	return []byte(config.GetEnv("JWT_SECRET", ""))
}

func generateJWT(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(14 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"type": "auth",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	return token.SignedString(getJWTKey())
}

func generateResetToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"type": "reset",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	return token.SignedString(getJWTKey())
}

func verifyResetToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	
	if err != nil {
		return "", err
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["type"] != "reset" {
			return "", jwt.ErrTokenInvalidClaims
		}
		return claims["sub"].(string), nil
	}
	
	return "", jwt.ErrTokenInvalidClaims
}