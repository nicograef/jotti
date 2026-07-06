//go:build unit

package user

import (
	"regexp"
	"testing"
)

func TestGenerateOnetimePassword(t *testing.T) {
	// Genau 6 Ziffern.
	valid := regexp.MustCompile(`^\d{6}$`)

	for range 50 {
		password, err := generateOnetimePassword()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !valid.MatchString(password) {
			t.Fatalf("Expected exactly 6 digits, got %q", password)
		}
	}
}

func TestOnetimePasswordSchema(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 6 digits", "123456", false},
		{"too short (5 digits)", "12345", true},
		{"too long (7 digits)", "1234567", true},
		{"letters", "abcdef", true},
		{"mixed", "12ab56", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issue := OnetimePasswordSchema.Validate(&tc.input)
			if tc.wantErr && issue == nil {
				t.Errorf("expected validation error for %q, got none", tc.input)
			}
			if !tc.wantErr && issue != nil {
				t.Errorf("expected no error for %q, got %v", tc.input, issue)
			}
		})
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
