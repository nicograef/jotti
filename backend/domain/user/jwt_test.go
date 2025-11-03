package user

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nicograef/jotti/backend/config"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	service := Service{Cfg: config.Config{JWTSecret: "test_secret"}}
	user := User{ID: 1, Role: AdminRole}

	token, err := service.GenerateJWTTokenForUser(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Validate the token
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.Cfg.JWTSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}

	if claims["sub"] != "1" {
		t.Errorf("Expected subject '1', got '%v'", claims["sub"])
	}
	if Role(claims["role"].(string)) != AdminRole {
		t.Errorf("Expected role '%s', got '%v'", AdminRole, claims["role"])
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	service := Service{Cfg: config.Config{JWTSecret: "test_secret"}}
	user := User{ID: 2, Role: ServiceRole}

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
