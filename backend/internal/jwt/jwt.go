package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "jotti"

func GenerateJWTTokenForUser(userID int, userRole string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  issuer,
		"iat":  jwt.NewNumericDate(time.Now().UTC()),
		"exp":  jwt.NewNumericDate(time.Now().UTC().Add(12 * time.Hour)),
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

// Claims are the validated claims jotti reads from a token.
type Claims struct {
	UserID   int
	Role     string
	IssuedAt time.Time
}

func ParseAndValidateJWTToken(tokenString, secret string) (Claims, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}

	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer(issuer))
	if err != nil {
		return Claims{}, err
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok || userIDFloat < 0 {
		return Claims{}, errors.New("invalid sub claim")
	}

	userRole, ok := claims["role"].(string)
	if !ok {
		return Claims{}, errors.New("invalid role claim")
	}

	issuedAt, err := claims.GetIssuedAt()
	if err != nil || issuedAt == nil {
		return Claims{}, errors.New("invalid iat claim")
	}

	return Claims{UserID: int(userIDFloat), Role: userRole, IssuedAt: issuedAt.Time}, nil
}
