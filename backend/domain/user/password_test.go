//go:build unit

package user

import (
	"regexp"
	"testing"
)

func TestGenerateOnetimePassword(t *testing.T) {
	// Alphanumerisch (ohne verwechselbare o, i, l, 1), 8 Zeichen.
	valid := regexp.MustCompile(`^[a-hj-km-np-z023-9]{8}$`)

	for range 50 {
		password, err := generateOnetimePassword()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !valid.MatchString(password) {
			t.Fatalf("Expected 8 unambiguous alphanumeric chars, got %q", password)
		}
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
