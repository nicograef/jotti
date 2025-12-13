package jwt

import (
	"errors"
	"time"

	jwtpkg "github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var errTokenGeneration = errors.New("token generation error")

const issuer = "jotti"

type User struct {
	ID       int
	Username string
	Role     string
}

func GenerateJWTTokenForUser(user User, secret string) (string, error) {
	token := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, jwtpkg.MapClaims{
		"alg":  jwtpkg.SigningMethodHS256.Alg(),
		"iss":  issuer,
		"iat":  jwtpkg.NewNumericDate(time.Now()),
		"exp":  jwtpkg.NewNumericDate(time.Now().Add(12 * time.Hour)), // 12 hours validity
		"sub":  user.ID,
		"role": user.Role,
	})

	key := []byte(secret)
	stringToken, err := token.SignedString(key)
	if err != nil {
		log.Error().
			Err(err).
			Str("username", user.Username).
			Msg("Failed to generate token")
		return "", errTokenGeneration
	}

	return stringToken, nil
}

type tokenPayload struct {
	UserID int
	Role   string
}

func ParseAndValidateJWTToken(tokenString, secret string) (*tokenPayload, error) {
	claims := jwtpkg.MapClaims{}
	keyFunc := func(token *jwtpkg.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	_, err := jwtpkg.ParseWithClaims(tokenString, claims, keyFunc, jwtpkg.WithValidMethods([]string{jwtpkg.SigningMethodHS256.Alg()}), jwtpkg.WithExpirationRequired(), jwtpkg.WithIssuer(issuer))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse token")
		return nil, err
	}

	payload := &tokenPayload{
		UserID: int(claims["sub"].(float64)),
		Role:   claims["role"].(string),
	}

	return payload, nil
}
