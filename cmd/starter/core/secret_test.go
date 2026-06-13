package core

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSecretLength(t *testing.T) {
	s := GenerateSecret()
	if len(s) != 64 {
		t.Fatalf("Laenge: got %d, want 64", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		t.Fatalf("kein gueltiges Hex: %v", err)
	}
}

func TestGenerateSecretUnique(t *testing.T) {
	first := GenerateSecret()
	second := GenerateSecret()
	if first == second {
		t.Fatal("zwei Aufrufe lieferten denselben Wert")
	}
}
