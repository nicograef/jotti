package user

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenGeneration = errors.New("token generation error")

func (s *Service) GenerateJWTTokenForUser(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  "jotti",
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
