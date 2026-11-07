package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestKassenbelegDrucken_Pfandrueckgabe_NegativeBetraege(t *testing.T) {
	cmd, auftragMock, rueckgabeID, _ := pfandBelegFixture(t)

	if _, err := cmd.KassenbelegDrucken(context.Background(), KassenbelegDruckenCommand{RueckgabeID: rueckgabeID}); err != nil {
		t.Fatalf("KassenbelegDrucken: %v", err)
	}

	got := pfandBelegPayload(t, auftragMock)
	if auftragMock.enqueued[0].Referenz != "pfand-zurueckgegeben:1" {
		t.Errorf("referenz = %q", auftragMock.enqueued[0].Referenz)
	}
	for _, check := range []string{"KASSENBELEG", "Bon-Nr: 1", "-2,00 x 2 = -4,00 EUR", "GESAMT: -4,00 EUR"} {
		if !strings.Contains(got, check) {
			t.Errorf("erwartet %q im Pfandbeleg, bekam:\n%q", check, got)
		}
	}
	for _, absent := range []string{"STORNOBELEG", "Storno zu Bon-Nr"} {
		if strings.Contains(got, absent) {
			t.Errorf("Pfandbeleg enthält %q:\n%q", absent, got)
		}
	}
}

func TestKassenbelegDrucken_PfandStorno_DruckbarAlsStornobeleg(t *testing.T) {
	cmd, auftragMock, rueckgabeID, stornierungID := pfandBelegFixture(t)

	if _, err := cmd.KassenbelegDrucken(context.Background(), KassenbelegDruckenCommand{RueckgabeID: rueckgabeID, StornierungID: stornierungID}); err != nil {
		t.Fatalf("KassenbelegDrucken: %v", err)
	}

	got := pfandBelegPayload(t, auftragMock)
	if auftragMock.enqueued[0].Referenz != "pfandrueckgabe-storniert:2" {
		t.Errorf("referenz = %q", auftragMock.enqueued[0].Referenz)
	}
	for _, check := range []string{"STORNOBELEG", "Bon-Nr: 2", "Storno zu Bon-Nr: 1", "2,00 x 1 = 2,00 EUR", "GESAMT: 2,00 EUR"} {
		if !strings.Contains(got, check) {
			t.Errorf("erwartet %q im Stornobeleg, bekam:\n%q", check, got)
		}
	}
	if strings.Contains(got, "-2,00") {
		t.Errorf("Stornobeleg zur Rückgabe enthält negative Beträge:\n%q", got)
	}
}

func TestKassenbelegDrucken_Pfand_NichtGefunden(t *testing.T) {
	cmd, _, rueckgabeID, _ := pfandBelegFixture(t)

	tests := []struct {
		name    string
		command KassenbelegDruckenCommand
		want    error
	}{
		{"unbekannte Rückgabe", KassenbelegDruckenCommand{RueckgabeID: uuid.New().String()}, ErrRueckgabeNichtGefunden},
		{"Stornierung zu unbekannter Rückgabe", KassenbelegDruckenCommand{RueckgabeID: uuid.New().String(), StornierungID: uuid.New().String()}, ErrRueckgabeNichtGefunden},
		{"unbekannte Stornierung", KassenbelegDruckenCommand{RueckgabeID: rueckgabeID, StornierungID: uuid.New().String()}, ErrStornierungNichtGefunden},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := cmd.KassenbelegDrucken(context.Background(), tc.command); !errors.Is(err, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, err)
			}
		})
	}
}
