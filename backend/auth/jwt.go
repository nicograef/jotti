package auth

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nicograef/jotti/backend/user"
)

var errTokenGeneration = errors.New("token generation error")

const issuer = "jotti"

func generateJWTTokenForUser(user user.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"alg":  jwt.SigningMethodHS256.Alg(),
		"iss":  issuer,
		"iat":  jwt.NewNumericDate(time.Now()),
		"exp":  jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
		"sub":  strconv.Itoa(user.ID),
		"role": user.Role,
	})

	key := []byte(secret)
	stringToken, err := token.SignedString(key)
	if err != nil {
		log.Printf("ERROR Failed to generate token for user %s: %v", user.Username, err)
		return "", errTokenGeneration
	}

	return stringToken, nil
}

type tokenPayload struct {
	UserID int
	Role   user.Role
}

func parseAndValidateJWTToken(tokenString, secret string) (*tokenPayload, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer(issuer))
	if err != nil {
		log.Printf("ERROR Failed to parse token: %v", err)
		return nil, err
	}

	userID, err := strconv.Atoi((claims["sub"].(string)))
	if err != nil {
		log.Printf("ERROR Failed to convert UserID to int: %v", err)
		return nil, err
	}

	payload := &tokenPayload{
		UserID: userID,
		Role:   user.Role(claims["role"].(string)),
	}

	return payload, nil
}
