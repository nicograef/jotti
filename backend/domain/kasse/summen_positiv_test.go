//go:build unit

package kasse

import (
	"fmt"
	"strings"
	"testing"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
)

// Alle Geldsummen-Felder der Kasse-Events und -Projektionen müssen positiv sein: eine Summe
// wird über Positionen mit Preis >= 1 Cent gebildet, 0-Cent-Positionen sind nicht zulässig.
// Die Felder sind daher GTE(1).Required(): zog überspringt den GTE-Validator beim
// Zero-Value, erst Required() lehnt die 0 ab.
//
// Fünf Felder werden über die echten Event-Konstruktoren geprüft (die Summe ist dort ein
// Parameter); die übrigen sechs leiten die Konstruktoren aus den Positionen ab und werden
// direkt gegen ihr Schema validiert.

func validEventPositionen() []PositionEventData {
	return []PositionEventData{{
		PositionID:       uuid.New().String(),
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 500,
		Menge:            1,
	}}
}

func validProjektionsPositionen() []Position {
	return []Position{{
		PositionID:       uuid.New().String(),
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 500,
		Menge:            1,
	}}
}

// validateSchema spiegelt das Fehler-Wrapping der Konstruktoren (Fehlerform mit Feldname).
func validateSchema[T any](schema *z.StructSchema, value *T) error {
	if errs := schema.Validate(value); errs != nil {
		return fmt.Errorf("%v", z.Issues.FlattenAndCollect(errs))
	}
	return nil
}

func TestGeldsummen_MussPositiv(t *testing.T) {
	subject := TischSessionSubject(1, 1)
	// Die Storno-/Zahlungs-/Umbuchungs-Konstruktoren generieren keine PositionIDs.
	positionen := validProjektionsPositionen()
	zahlungID := uuid.New().String()

	cases := []struct {
		name  string
		field string
		run   func(summe int) error
	}{
		{
			name:  "ZahlungKassiert.GesamtZahlungCents",
			field: "GesamtZahlungCents",
			run: func(summe int) error {
				_, err := NewZahlungKassiertEvent(subject, 1, "TestUser", positionen, summe, "")
				return err
			},
		},
		{
			name:  "StornierungErteilt.GesamtStornierungCents",
			field: "GesamtStornierungCents",
			run: func(summe int) error {
				_, err := NewStornierungErteiltEvent(subject, 1, "TestUser", zahlungID, positionen, summe, "Test")
				return err
			},
		},
		{
			name:  "BestellungKorrigiert.GesamtCents",
			field: "GesamtCents",
			run: func(summe int) error {
				_, err := NewBestellungKorrigiertEvent(subject, 1, "TestUser", positionen, summe, "Test")
				return err
			},
		},
		{
			name:  "BestellungUmgebucht.GesamtCents",
			field: "GesamtCents",
			run: func(summe int) error {
				_, _, err := NewBestellungUmgebuchtEvents(1, 1, 2, 1, "TestUser", positionen, summe, "auf Tisch 2", "von Tisch 1", "")
				return err
			},
		},
		{
			name:  "DirektverkaufStorniert.GesamtStornierungCents",
			field: "GesamtStornierungCents",
			run: func(summe int) error {
				_, err := NewDirektverkaufStorniertEvent(DirektverkaufSubject(1, uuid.New().String()), uuid.New().String(), 1, "TestUser", positionen, summe, "Test")
				return err
			},
		},
		{
			name:  "BestellungAufgenommen.GesamtPreisCents (event)",
			field: "GesamtPreisCents",
			run: func(summe int) error {
				data := BestellungAufgenommenV1Data{BestellungID: uuid.New().String(), Positionen: validEventPositionen(), GesamtPreisCents: summe}
				return validateSchema(bestellungAufgenommenV1DataSchema, &data)
			},
		},
		{
			name:  "DirektverkaufGetaetigt.GesamtbetragCents (event)",
			field: "GesamtbetragCents",
			run: func(summe int) error {
				data := DirektverkaufGetaetigtV1Data{VerkaufID: uuid.New().String(), Positionen: validEventPositionen(), GesamtbetragCents: summe}
				return validateSchema(direktverkaufGetaetigtV1DataSchema, &data)
			},
		},
		{
			name:  "Bestellung.GesamtPreisCents (projektion)",
			field: "GesamtPreisCents",
			run: func(summe int) error {
				b := Bestellung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtPreisCents: summe, AufgenommenAm: time.Now().UTC()}
				return validateSchema(bestellungSchema, &b)
			},
		},
		{
			name:  "Zahlung.GesamtZahlungCents (projektion)",
			field: "GesamtZahlungCents",
			run: func(summe int) error {
				zahlung := Zahlung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtZahlungCents: summe, KassiertAm: time.Now().UTC()}
				return validateSchema(zahlungSchema, &zahlung)
			},
		},
		{
			name:  "Stornierung.GesamtStornierungCents (projektion)",
			field: "GesamtStornierungCents",
			run: func(summe int) error {
				s := Stornierung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtStornierungCents: summe, Kommentar: "Test", StorniertAm: time.Now().UTC()}
				return validateSchema(stornierungSchema, &s)
			},
		},
		{
			name:  "Umbuchung.GesamtCents (projektion)",
			field: "GesamtCents",
			run: func(summe int) error {
				u := Umbuchung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, QuellTischID: 1, ZielTischID: 2, Positionen: validProjektionsPositionen(), GesamtCents: summe}
				return validateSchema(umbuchungSchema, &u)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/lehntNullAb", func(t *testing.T) {
			err := tc.run(0)
			if err == nil {
				t.Fatalf("expected validation error for 0-sum on %s, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("expected %s validation error, got %v", tc.field, err)
			}
		})
		t.Run(tc.name+"/lehntNegativAb", func(t *testing.T) {
			err := tc.run(-1)
			if err == nil {
				t.Fatalf("expected validation error for negative sum on %s, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("expected %s validation error, got %v", tc.field, err)
			}
		})
		t.Run(tc.name+"/erlaubtPositiv", func(t *testing.T) {
			if err := tc.run(1); err != nil {
				t.Fatalf("expected positive sum to be valid for %s, got %v", tc.field, err)
			}
		})
	}
}
