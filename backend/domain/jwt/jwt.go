package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "jotti"

func GenerateJWTTokenForUser(userID int, userRole string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"alg":  jwt.SigningMethodHS256.Alg(),
		"iss":  issuer,
		"iat":  jwt.NewNumericDate(time.Now()),
		"exp":  jwt.NewNumericDate(time.Now().Add(12 * time.Hour)), // 12 hours validity
		"sub":  userID,
		"role": userRole,
	})

	key := []byte(secret)
	stringToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func ParseAndValidateJWTToken(tokenString, secret string) (int, string, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer(issuer))
	if err != nil {
		return 0, "", err
	}

	userID := int(claims["sub"].(float64))
	userRole := claims["role"].(string)

	return userID, userRole, nil
}
