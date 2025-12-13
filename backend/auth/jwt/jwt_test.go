//go:build unit

package jwt

import (
	"testing"

	jwtpkg "github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	token, err := GenerateJWTTokenForUser(User{ID: 1, Role: "admin"}, "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Validate the token
	claims := jwtpkg.MapClaims{}
	_, err = jwtpkg.ParseWithClaims(token, claims, func(token *jwtpkg.Token) (interface{}, error) {
		return []byte("test_secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}

	if int(claims["sub"].(float64)) != 1 {
		t.Errorf("Expected subject '1', got '%v'", int(claims["sub"].(float64)))
	}
	if claims["role"].(string) != "admin" {
		t.Errorf("Expected role '%s', got '%v'", "admin", claims["role"])
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	user := User{ID: 2, Role: "service"}

	token, err := GenerateJWTTokenForUser(user, "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	payload, err := ParseAndValidateJWTToken(token, "test_secret")
	if err != nil {
		t.Fatalf("Failed to parse and validate JWT token: %v", err)
	}

	if payload.UserID != user.ID {
		t.Errorf("Expected UserID '%d', got '%d'", user.ID, payload.UserID)
	}
	if payload.Role != user.Role {
		t.Errorf("Expected Role '%s', got '%s'", user.Role, payload.Role)
	}
}
