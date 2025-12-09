//go:build unit

package user

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	token, err := generateJWTTokenForUser(User{ID: 1, Role: AdminRole}, "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Validate the token
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test_secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}

	if int(claims["sub"].(float64)) != 1 {
		t.Errorf("Expected subject '1', got '%v'", int(claims["sub"].(float64)))
	}
	if Role(claims["role"].(string)) != AdminRole {
		t.Errorf("Expected role '%s', got '%v'", AdminRole, claims["role"])
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	user := User{ID: 2, Role: ServiceRole}

	token, err := generateJWTTokenForUser(user, "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	payload, err := parseAndValidateJWTToken(token, "test_secret")
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
