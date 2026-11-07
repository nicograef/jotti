package kasse

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Every money sum of Kasse events and projections is >= 1 Cent, hence GTE(1).Required(): zog skips
// GTE on the zero value, only Required() rejects 0. Six fields go through the event constructors
// taking the sum; the other seven (summed from positions or copied from the event) through their schema.

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
			name:  "PfandrueckgabeStorniert.GesamtStornierungCents",
			field: "GesamtStornierungCents",
			run: func(summe int) error {
				rueckgabeID := uuid.New().String()
				_, err := NewPfandrueckgabeStorniertEvent(PfandrueckgabeSubject(1, rueckgabeID), uuid.New().String(), rueckgabeID, 0, 1, "TestUser", positionen, summe, "Test")
				return err
			},
		},
		{
			name:  "BestellungAufgenommen.GesamtPreisCents (event)",
			field: "GesamtPreisCents",
			run: func(summe int) error {
				data := BestellungAufgenommenV1Data{BestellungID: uuid.New().String(), Positionen: validEventPositionen(), GesamtPreisCents: summe}
				return validateEventData(bestellungAufgenommenV1DataSchema, &data, "test")
			},
		},
		{
			name:  "DirektverkaufGetaetigt.GesamtbetragCents (event)",
			field: "GesamtbetragCents",
			run: func(summe int) error {
				data := DirektverkaufGetaetigtV1Data{VerkaufID: uuid.New().String(), Positionen: validEventPositionen(), GesamtbetragCents: summe}
				return validateEventData(direktverkaufGetaetigtV1DataSchema, &data, "test")
			},
		},
		{
			name:  "PfandZurueckgegeben.GesamtbetragCents",
			field: "GesamtbetragCents",
			run: func(summe int) error {
				data := PfandZurueckgegebenV1Data{RueckgabeID: uuid.New().String(), Positionen: validEventPositionen(), GesamtbetragCents: summe}
				return validateEventData(pfandZurueckgegebenV1DataSchema, &data, "test")
			},
		},
		{
			name:  "Bestellung.GesamtPreisCents (projektion)",
			field: "GesamtPreisCents",
			run: func(summe int) error {
				b := Bestellung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtPreisCents: summe, AufgenommenAm: time.Now().UTC()}
				return validateEventData(bestellungSchema, &b, "test")
			},
		},
		{
			name:  "Zahlung.GesamtZahlungCents (projektion)",
			field: "GesamtZahlungCents",
			run: func(summe int) error {
				zahlung := Zahlung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtZahlungCents: summe, KassiertAm: time.Now().UTC()}
				return validateEventData(zahlungSchema, &zahlung, "test")
			},
		},
		{
			name:  "Stornierung.GesamtStornierungCents (projektion)",
			field: "GesamtStornierungCents",
			run: func(summe int) error {
				s := Stornierung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, Positionen: validProjektionsPositionen(), GesamtStornierungCents: summe, Kommentar: "Test", StorniertAm: time.Now().UTC()}
				return validateEventData(stornierungSchema, &s, "test")
			},
		},
		{
			name:  "Umbuchung.GesamtCents (projektion)",
			field: "GesamtCents",
			run: func(summe int) error {
				u := Umbuchung{ID: uuid.New().String(), UserID: 1, UserName: "TestUser", TischID: 1, QuellTischID: 1, ZielTischID: 2, Positionen: validProjektionsPositionen(), GesamtCents: summe}
				return validateEventData(umbuchungSchema, &u, "test")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/lehntNullAb", func(t *testing.T) {
			err := tc.run(0)
			if err == nil {
				t.Fatalf("expected validation error for 0-sum on %s, got nil", tc.field)
			}
			if !errors.Is(err, ErrEventDatenUngueltig) {
				t.Errorf("expected ErrEventDatenUngueltig for %s, got %v", tc.field, err)
			}
		})
		t.Run(tc.name+"/lehntNegativAb", func(t *testing.T) {
			err := tc.run(-1)
			if err == nil {
				t.Fatalf("expected validation error for negative sum on %s, got nil", tc.field)
			}
			if !errors.Is(err, ErrEventDatenUngueltig) {
				t.Errorf("expected ErrEventDatenUngueltig for %s, got %v", tc.field, err)
			}
		})
		t.Run(tc.name+"/erlaubtPositiv", func(t *testing.T) {
			if err := tc.run(1); err != nil {
				t.Errorf("expected positive sum to be valid for %s, got %v", tc.field, err)
			}
		})
	}
}
