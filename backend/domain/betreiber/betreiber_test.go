//go:build unit

package betreiber

import (
	"strings"
	"testing"
)

func TestNewBetreiber_LehntZuLangenVereinsnamenAb(t *testing.T) {
	zuLang := strings.Repeat("a", 61)

	if _, err := NewBetreiber(zuLang, "Musterstraße 1", "12345", "Musterstadt", nil, nil); err == nil {
		t.Errorf("NewBetreiber nahm einen Vereinsnamen mit %d Zeichen an; die amtliche MaxLength ist 60", len(zuLang))
	}
}

func TestNewBetreiber_TrimmtJedesFeld(t *testing.T) {
	steuernummer := "  12/345/67890  "
	ustID := "  DE123456789  "

	b, err := NewBetreiber("  Sportverein  ", "  Musterstraße 1  ", "  12345  ", "  Musterstadt  ", &steuernummer, &ustID)
	if err != nil {
		t.Fatalf("NewBetreiber: %v", err)
	}

	if b.Vereinsname != "Sportverein" {
		t.Errorf("Vereinsname ist %q, erwartet %q", b.Vereinsname, "Sportverein")
	}
	if b.Strasse != "Musterstraße 1" {
		t.Errorf("Strasse ist %q, erwartet %q", b.Strasse, "Musterstraße 1")
	}
	if b.Plz != "12345" {
		t.Errorf("Plz ist %q, erwartet %q", b.Plz, "12345")
	}
	if b.Ort != "Musterstadt" {
		t.Errorf("Ort ist %q, erwartet %q", b.Ort, "Musterstadt")
	}
	if *b.Steuernummer != "12/345/67890" {
		t.Errorf("Steuernummer ist %q, erwartet %q", *b.Steuernummer, "12/345/67890")
	}
	if *b.UstID != "DE123456789" {
		t.Errorf("UstID ist %q, erwartet %q", *b.UstID, "DE123456789")
	}
}

func TestNewBetreiber_LehntFeldAusLeerzeichenAb(t *testing.T) {
	if _, err := NewBetreiber("   ", "Musterstraße 1", "12345", "Musterstadt", nil, nil); err == nil {
		t.Error("NewBetreiber nahm einen Vereinsnamen aus Leerzeichen an")
	}
}

func TestNewBetreiber_NimmtOptionaleFelderOhneWert(t *testing.T) {
	b, err := NewBetreiber("Sportverein", "Musterstraße 1", "12345", "Musterstadt", nil, nil)
	if err != nil {
		t.Fatalf("NewBetreiber: %v", err)
	}

	if b.Steuernummer != nil || b.UstID != nil {
		t.Errorf("Steuernummer/UstID sind %v/%v, erwartet nil/nil", b.Steuernummer, b.UstID)
	}
}
