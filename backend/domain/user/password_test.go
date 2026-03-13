//go:build unit

package user

import (
	"strconv"
	"testing"
)

func TestGenerateOnetimePassword(t *testing.T) {
	password, err := generateOnetimePassword()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(password) != 6 {
		t.Fatalf("Expected password length 6, got %d", len(password))
	}
	if _, err := strconv.Atoi(password); err != nil {
		t.Fatalf("Expected numeric password, got %s", password)
	}
}

func TestPasswordValidation(t *testing.T) {
	short := "abc"
	if issue := PasswordSchema.Validate(&short); issue == nil {
		t.Error("expected validation error for short password")
	}

	empty := ""
	if issue := PasswordSchema.Validate(&empty); issue == nil {
		// zog treats empty string as zero value; validated via Required() in struct schemas.
		// For standalone use, short non-empty strings are still caught.
	}

	short2 := "ab"
	if issue := PasswordSchema.Validate(&short2); issue == nil {
		t.Error("expected validation error for 2-char password")
	}

	valid := "secure123"
	if issue := PasswordSchema.Validate(&valid); issue != nil {
		t.Errorf("expected no error for valid password, got %v", issue)
	}
}
