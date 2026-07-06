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

func TestQRVersionForLengthM_KnownCapacities(t *testing.T) {
	cases := []struct {
		payloadLen  int
		wantVersion int
	}{
		{1, 1},    // V1 haelt 16 Byte
		{16, 1},   // V1 haelt genau 16 Byte
		{17, 2},   // V2 ab 17 Byte
		{507, 17}, // V17 haelt genau 507 Byte
		{508, 18}, // V18 ab 508 Byte
	}
	for _, tc := range cases {
		got := qrVersionForLengthM(tc.payloadLen)
		if got != tc.wantVersion {
			t.Errorf("qrVersionForLengthM(%d) = %d, want %d", tc.payloadLen, got, tc.wantVersion)
		}
	}
}

func TestQRModuleSizeByte_500BytePayload_UsesSize6(t *testing.T) {
	// V17 (507-Byte-Kapazitaet): Matrix 85 Module + 8 Ruhezone = 93 Module.
	// 93 * 6 = 558 Dots <= 576 Dots -> Modulgroesse 6.
	if got := qrModuleSizeByte(500); got != 6 {
		t.Errorf("qrModuleSizeByte(500) = %d, want 6 (93 Module * 6 = 558 <= 576 Dots)", got)
	}
}

func TestQRModuleSizeByte_508BytePayload_UsesSize5(t *testing.T) {
	// V18 (563-Byte-Kapazitaet): Matrix 89 Module + 8 Ruhezone = 97 Module.
	// 97 * 6 = 582 Dots > 576 -> Modulgroesse 5: 97 * 5 = 485 <= 576 Dots.
	if got := qrModuleSizeByte(508); got != 5 {
		t.Errorf("qrModuleSizeByte(508) = %d, want 5 (97 Module * 6 = 582 > 576; * 5 = 485 <= 576)", got)
	}
}

func TestQRModuleSizeByte_AllLengthsUpTo600_FitWithin576Dots(t *testing.T) {
	for payloadLen := 1; payloadLen <= 600; payloadLen++ {
		v := qrVersionForLengthM(payloadLen)
		size := qrModuleSizeByte(payloadLen)
		totalModules := 4*v + 17 + 8
		dotsWide := totalModules * int(size)
		if dotsWide > 576 {
			t.Errorf("payload %d Byte: Version %d, Modulgroesse %d -> %d Dots > 576",
				payloadLen, v, size, dotsWide)
		}
	}
}
