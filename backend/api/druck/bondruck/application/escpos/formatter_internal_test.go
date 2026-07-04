//go:build unit

package escpos

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncate_KeepsShortStringUnchanged(t *testing.T) {
	if got := truncate("Maria", 24); got != "Maria" {
		t.Errorf("kurzer Name sollte unveraendert bleiben; got %q", got)
	}
}

func TestTruncate_CountsRunesNotBytes(t *testing.T) {
	// 24 Umlaute = 48 Bytes, aber nur 24 Runen -> darf nicht gekuerzt werden.
	name := strings.Repeat("ä", 24)
	if got := truncate(name, 24); got != name {
		t.Errorf("24 Runen sollten nicht gekuerzt werden; got %d Runen", utf8.RuneCountInString(got))
	}
}

func TestTruncate_CutsOnRuneBoundary(t *testing.T) {
	name := strings.Repeat("ä", 30) // 30 Runen, 60 Bytes
	got := truncate(name, 24)

	if !utf8.ValidString(got) {
		t.Errorf("truncate erzeugte ungueltiges UTF-8: %q", got)
	}
	if utf8.RuneCountInString(got) != 24 {
		t.Errorf("Ergebnis sollte 24 Runen haben; got %d", utf8.RuneCountInString(got))
	}
}

func TestWrapLine_CountsRunesNotBytes(t *testing.T) {
	// "ääää öööö": 9 Runen, 17 Bytes. Bei Breite 10 runenbasiert kein Umbruch.
	if got := wrapLine("ääää öööö", 10); strings.Contains(got, "\n") {
		t.Errorf("runenbasiert sollte bei Breite 10 kein Umbruch erfolgen; got %q", got)
	}
}

func TestWrapLine_BreaksOnWordBoundary(t *testing.T) {
	// "aaaa bbbb cccc" bei Breite 9 -> Umbruch vor "cccc".
	if got := wrapLine("aaaa bbbb cccc", 9); !strings.Contains(got, "\n") {
		t.Errorf("erwartet Umbruch bei Breite 9; got %q", got)
	}
}
