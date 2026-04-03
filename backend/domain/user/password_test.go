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
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"too short (2 chars)", "ab", true},
		{"too short (3 chars)", "abc", true},
		{"valid (8+ chars)", "secure123", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issue := PasswordSchema.Validate(&tc.input)
			if tc.wantErr && issue == nil {
				t.Errorf("expected validation error for %q, got none", tc.input)
			}
			if !tc.wantErr && issue != nil {
				t.Errorf("expected no error for %q, got %v", tc.input, issue)
			}
		})
	}
}
