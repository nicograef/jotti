//go:build unit

package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTTokenForUser(t *testing.T) {
	token, err := GenerateJWTTokenForUser(1, "admin", "admin", "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Validate the token
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte("test_secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}

	if int(claims["sub"].(float64)) != 1 {
		t.Errorf("Expected subject '1', got '%v'", int(claims["sub"].(float64)))
	}
	if claims["username"].(string) != "admin" {
		t.Errorf("Expected username 'admin', got '%v'", claims["username"])
	}
	if claims["role"].(string) != "admin" {
		t.Errorf("Expected role '%s', got '%v'", "admin", claims["role"])
	}
	if _, ok := claims["alg"]; ok {
		t.Error("Expected no alg claim in token payload")
	}
}

func TestParseAndValidateJWTToken(t *testing.T) {
	token, err := GenerateJWTTokenForUser(2, "service", "service", "test_secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	userID, username, userRole, err := ParseAndValidateJWTToken(token, "test_secret")
	if err != nil {
		t.Fatalf("Failed to parse and validate JWT token: %v", err)
	}

	if userID != 2 {
		t.Errorf("Expected UserID '%d', got '%d'", 2, userID)
	}
	if username != "service" {
		t.Errorf("Expected username '%s', got '%s'", "service", username)
	}
	if userRole != "service" {
		t.Errorf("Expected Role '%s', got '%s'", "service", userRole)
	}
}

func makeTokenWithClaims(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("test_secret"))
	return s
}

func baseClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":      issuer,
		"iat":      jwt.NewNumericDate(time.Now().UTC()),
		"exp":      jwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour)),
		"sub":      float64(1),
		"username": "testuser",
		"role":     "admin",
	}
}

func TestParseAndValidateJWTToken_MalformedClaims(t *testing.T) {
	t.Run("sub missing", func(t *testing.T) {
		c := baseClaims()
		delete(c, "sub")
		_, _, _, err := ParseAndValidateJWTToken(makeTokenWithClaims(c), "test_secret")
		if err == nil {
			t.Error("expected error for missing sub, got nil")
		}
	})

	t.Run("sub wrong type", func(t *testing.T) {
		c := baseClaims()
		c["sub"] = "not-a-number"
		_, _, _, err := ParseAndValidateJWTToken(makeTokenWithClaims(c), "test_secret")
		if err == nil {
			t.Error("expected error for string sub, got nil")
		}
	})

	t.Run("sub negative", func(t *testing.T) {
		c := baseClaims()
		c["sub"] = float64(-1)
		_, _, _, err := ParseAndValidateJWTToken(makeTokenWithClaims(c), "test_secret")
		if err == nil {
			t.Error("expected error for negative sub, got nil")
		}
	})

	t.Run("role missing", func(t *testing.T) {
		c := baseClaims()
		delete(c, "role")
		_, _, _, err := ParseAndValidateJWTToken(makeTokenWithClaims(c), "test_secret")
		if err == nil {
			t.Error("expected error for missing role, got nil")
		}
	})

	t.Run("role wrong type", func(t *testing.T) {
		c := baseClaims()
		c["role"] = float64(42)
		_, _, _, err := ParseAndValidateJWTToken(makeTokenWithClaims(c), "test_secret")
		if err == nil {
			t.Error("expected error for numeric role, got nil")
		}
	})
}
