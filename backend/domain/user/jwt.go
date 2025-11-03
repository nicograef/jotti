package user

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenGeneration = errors.New("token generation error")

const issuer = "jotti"

func (s *Service) GenerateJWTTokenForUser(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"alg":  jwt.SigningMethodHS256.Alg(),
		"iss":  issuer,
		"iat":  jwt.NewNumericDate(time.Now()),
		"exp":  jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
		"sub":  strconv.Itoa(user.ID),
		"role": user.Role,
	})

	key := []byte(s.Cfg.JWTSecret)
	stringToken, err := token.SignedString(key)
	if err != nil {
		log.Printf("ERROR Failed to generate token for user %s: %v", user.Username, err)
		return "", ErrTokenGeneration
	}

	return stringToken, nil
}

type TokenPayload struct {
	UserID int
	Role   Role
}

func (s *Service) ParseAndValidateJWTToken(tokenString string) (*TokenPayload, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Cfg.JWTSecret), nil
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

	payload := &TokenPayload{
		UserID: userID,
		Role:   Role(claims["role"].(string)),
	}

	return payload, nil
}
