//go:build unit

package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	token, err := GenerateJWTTokenForUser(1, "admin", "test_secret")
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
	if claims["role"].(string) != "admin" {
		t.Errorf("Expected role '%s', got '%v'", "admin", claims["role"])
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	token, err := GenerateJWTTokenForUser(2, "service", "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	userID, userRole, err := ParseAndValidateJWTToken(token, "test_secret")
	if err != nil {
		t.Fatalf("Failed to parse and validate JWT token: %v", err)
	}

	if userID != 2 {
		t.Errorf("Expected UserID '%d', got '%d'", 2, userID)
	}
	if userRole != "service" {
		t.Errorf("Expected Role '%s', got '%s'", "service", userRole)
	}
}
