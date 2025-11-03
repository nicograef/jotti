//go:build unit

package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nicograef/jotti/backend/domain/user"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	service := Service{JWTSecret: "test_secret"}

	token, err := service.GenerateJWTTokenForUser(user.User{ID: 1, Role: user.AdminRole})
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Validate the token
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.JWTSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}

	if claims["sub"] != "1" {
		t.Errorf("Expected subject '1', got '%v'", claims["sub"])
	}
	if user.Role(claims["role"].(string)) != user.AdminRole {
		t.Errorf("Expected role '%s', got '%v'", user.AdminRole, claims["role"])
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	service := Service{JWTSecret: "test_secret"}
	user := user.User{ID: 2, Role: user.ServiceRole}

	token, err := service.GenerateJWTTokenForUser(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	payload, err := service.ParseAndValidateJWTToken(token)
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
