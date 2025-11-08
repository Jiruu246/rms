package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(JWTKey []byte, userID string, role string, exp time.Time) (string, error) {
	// Define token claims
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString(JWTKey)
}

// ValidateJWT validates a JWT token and returns the claims if valid
func ValidateJWT(JWTKey []byte, tokenString string) (JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWTKey, nil
	})

	if err != nil {
		return JWTClaims{}, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return *claims, nil
	}

	return JWTClaims{}, errors.New("invalid token")
}
