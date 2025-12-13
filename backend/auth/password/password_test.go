//go:build unit

package password

import (
	"strconv"
	"testing"
)

func TestGenerateOnetimePassword(t *testing.T) {
	password, err := GenerateOnetimePassword()
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
