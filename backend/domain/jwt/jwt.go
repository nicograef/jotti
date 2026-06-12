package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "jotti"

func GenerateJWTTokenForUser(userID int, userName, userRole string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  issuer,
		"iat":  jwt.NewNumericDate(time.Now().UTC()),
		"exp":  jwt.NewNumericDate(time.Now().UTC().Add(12 * time.Hour)),
		"sub":  userID,
		"name": userName,
		"role": userRole,
	})

	key := []byte(secret)
	stringToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func ParseAndValidateJWTToken(tokenString, secret string) (int, string, string, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}

	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer(issuer))
	if err != nil {
		return 0, "", "", err
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok || userIDFloat < 0 {
		return 0, "", "", errors.New("invalid sub claim")
	}
	userID := int(userIDFloat)
	userName, _ := claims["name"].(string)
	userRole, ok := claims["role"].(string)
	if !ok {
		return 0, "", "", errors.New("invalid role claim")
	}

	return userID, userName, userRole, nil
}
