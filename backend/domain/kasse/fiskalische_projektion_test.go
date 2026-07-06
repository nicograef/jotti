//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// projektionsEvent baut ein Event mit dem gegebenen Typ, Subject und JSON-serialisierten Daten.
func projektionsEvent(t *testing.T, typ EventType, subject string, data any) e.Event {
	t.Helper()
	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("event-daten serialisieren: %v", err)
	}
	return e.Event{Type: string(typ), Subject: subject, Version: 1, Data: payload}
}

// testPositionen ist die Standard-Position der Projektions-Tests:
// 2 × 5,00 € zum Regelsteuersatz — Brutto 10,00 €.
func testPositionen() []PositionEventData {
	return []PositionEventData{{
		PositionID:       "0f0e0d0c-0b0a-4908-8706-050403020100",
		ProduktName:      "Bratwurst",
		Steuersatz:       "regel",
		EinzelpreisCents: 500,
		Menge:            2,
	}}
}

// TestFiskalischeProjektion prueft tabellengetrieben je Event-Typ: signaturpflichtig
// ja/nein, processType und processData inklusive Vorzeichen-/Faktor-Faellen
// (Storno, Korrektur, Umbuchungs-Seiten, Differenz) und der datenabhaengigen
// Sitzungseroeffnung (mit/ohne Anfangsbestand).
func TestFiskalischeProjektion(t *testing.T) {
	tischSubject := "kassensitzung-1/tisch-7"

	tests := []struct {
		name            string
		event           e.Event
		pflichtig       bool
		wantProcessType string
		wantProcessData string
	}{
		{
			name: "bestellung aufgenommen: Bestellung-V1 mit positiven Mengen",
			event: projektionsEvent(t, EventTypeBestellungAufgenommenV1, tischSubject, BestellungAufgenommenV1Data{
				BestellungID: "b", Positionen: testPositionen(), GesamtPreisCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeBestellungV1,
			wantProcessData: `2;"Bratwurst";5.00`,
		},
		{
			name: "bestellung korrigiert: Bestellung-V1 mit negativen Mengen",
			event: projektionsEvent(t, EventTypeBestellungKorrigiertV1, tischSubject, BestellungKorrigiertV1Data{
				KorrekturID: "k", Positionen: testPositionen(), GesamtCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeBestellungV1,
			wantProcessData: `-2;"Bratwurst";5.00`,
		},
		{
			name: "umbuchung quelltisch: Abgang mit negativen Mengen",
			event: projektionsEvent(t, EventTypeBestellungUmgebuchtV1, "kassensitzung-1/tisch-7", BestellungUmgebuchtV1Data{
				UmbuchungID: "u", QuellTischID: 7, ZielTischID: 9, Positionen: testPositionen(), GesamtCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeBestellungV1,
			wantProcessData: `-2;"Bratwurst";5.00`,
		},
		{
			name: "umbuchung zieltisch: Zugang mit positiven Mengen",
			event: projektionsEvent(t, EventTypeBestellungUmgebuchtV1, "kassensitzung-1/tisch-9", BestellungUmgebuchtV1Data{
				UmbuchungID: "u", QuellTischID: 7, ZielTischID: 9, Positionen: testPositionen(), GesamtCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeBestellungV1,
			wantProcessData: `2;"Bratwurst";5.00`,
		},
		{
			name: "zahlung kassiert: Kassenbeleg-V1 mit Zahlungsteil",
			event: projektionsEvent(t, EventTypeZahlungKassiertV1, tischSubject, ZahlungKassiertV1Data{
				ZahlungID: "z", Positionen: testPositionen(), GesamtZahlungCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^10.00_0.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name: "stornierung erteilt: Kassenbeleg-V1 mit negierten Betraegen",
			event: projektionsEvent(t, EventTypeStornierungErteiltV1, tischSubject, StornierungErteiltV1Data{
				StornierungID: "s", ZahlungID: "z", Positionen: testPositionen(), GesamtStornierungCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^-10.00_0.00_0.00_0.00_0.00^-10.00:Bar",
		},
		{
			name: "direktverkauf getaetigt: Kassenbeleg-V1 mit Zahlungsteil",
			event: projektionsEvent(t, EventTypeDirektverkaufGetaetigtV1, "kassensitzung-1/direktverkauf-d1", DirektverkaufGetaetigtV1Data{
				VerkaufID: "d", Positionen: testPositionen(), GesamtbetragCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^10.00_0.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name: "direktverkauf storniert: Kassenbeleg-V1 mit negierten Betraegen",
			event: projektionsEvent(t, EventTypeDirektverkaufStorniertV1, "kassensitzung-1/direktverkauf-d1", DirektverkaufStorniertV1Data{
				StornierungID: "s", Positionen: testPositionen(), GesamtStornierungCents: 1000,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^-10.00_0.00_0.00_0.00_0.00^-10.00:Bar",
		},
		{
			name: "sitzungseroeffnung mit anfangsbestand: Eigenbeleg (Bareinlage)",
			event: projektionsEvent(t, EventTypeKassensitzungEroeffnetV1, "kassensitzung-1", KassensitzungEroeffnetV1Data{
				Datum: "2026-06-12", Bezeichnung: "Sommerfest", BetragCents: 15000, EroeffnetVon: 1,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^0.00_0.00_0.00_0.00_150.00^150.00:Bar",
		},
		{
			name: "sitzungseroeffnung ohne anfangsbestand: kein Geschaeftsvorfall",
			event: projektionsEvent(t, EventTypeKassensitzungEroeffnetV1, "kassensitzung-1", KassensitzungEroeffnetV1Data{
				Datum: "2026-06-12", Bezeichnung: "Sommerfest", BetragCents: 0, EroeffnetVon: 1,
			}),
			pflichtig: false,
		},
		{
			name: "geldtransit einlage: Eigenbeleg mit positivem Betrag",
			event: projektionsEvent(t, EventTypeGeldtransitGebuchtV1, "kassensitzung-1", GeldtransitGebuchtV1Data{
				BewegungID: "g", Richtung: "einlage", BetragCents: 5000, GebuchtVon: 1,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^0.00_0.00_0.00_0.00_50.00^50.00:Bar",
		},
		{
			name: "geldtransit entnahme: Eigenbeleg mit negativem Betrag",
			event: projektionsEvent(t, EventTypeGeldtransitGebuchtV1, "kassensitzung-1", GeldtransitGebuchtV1Data{
				BewegungID: "g", Richtung: "entnahme", BetragCents: 5000, GebuchtVon: 1,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^0.00_0.00_0.00_0.00_-50.00^-50.00:Bar",
		},
		{
			name: "differenz fehlbetrag: Eigenbeleg mit Ist minus Soll",
			event: projektionsEvent(t, EventTypeDifferenzSollIstGebuchtV1, "kassensitzung-1", DifferenzSollIstGebuchtV1Data{
				BetragCents: 300, GebuchtVon: 1, // Soll − Ist = +300 → Bargeldbewegung −3.00
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^0.00_0.00_0.00_0.00_-3.00^-3.00:Bar",
		},
		{
			name: "differenz ueberschuss: Eigenbeleg mit positivem Betrag",
			event: projektionsEvent(t, EventTypeDifferenzSollIstGebuchtV1, "kassensitzung-1", DifferenzSollIstGebuchtV1Data{
				BetragCents: -200, GebuchtVon: 1, // Ist > Soll → Bargeldbewegung +2.00
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeKassenbelegV1,
			wantProcessData: "Beleg^0.00_0.00_0.00_0.00_2.00^2.00:Bar",
		},
		{
			name: "tagesabschluss: SonstigerVorgang mit Z-Nummer und Zeitraum",
			event: projektionsEvent(t, EventTypeTagesabschlussErstelltV1, "kassensitzung-2", TagesabschlussErstelltV1Data{
				ZNr:         2,
				ZeitraumVon: time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC),
				ZeitraumBis: time.Date(2026, 6, 13, 22, 0, 0, 0, time.UTC),
				ErstelltVon: 1,
			}),
			pflichtig:       true,
			wantProcessType: tse.ProcessTypeSonstigerVorgang,
			wantProcessData: "Tagesabschluss^ZNr:2^Von:2026-06-13T10:00:00Z^Bis:2026-06-13T22:00:00Z",
		},
		{
			name:      "ausgabe bestaetigt: nicht signaturpflichtig",
			event:     projektionsEvent(t, EventTypeAusgabeBestaetigtV1, tischSubject, AusgabeBestaetigtV1Data{AusgabeID: "a", Positionen: testPositionen()}),
			pflichtig: false,
		},
		{
			name:      "kassensturz durchgefuehrt: nicht signaturpflichtig",
			event:     projektionsEvent(t, EventTypeKassensturzDurchgefuehrtV1, "kassensitzung-1", KassensturzDurchgefuehrtV1Data{IstBestandCents: 100, DurchgefuehrtVon: 1}),
			pflichtig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vorgang, pflichtig, err := FiskalischeProjektion(tt.event)
			if err != nil {
				t.Fatalf("FiskalischeProjektion: %v", err)
			}
			if pflichtig != tt.pflichtig {
				t.Fatalf("signaturpflichtig = %v, erwartet %v", pflichtig, tt.pflichtig)
			}
			if !tt.pflichtig {
				return
			}
			if vorgang.ProcessType != tt.wantProcessType {
				t.Errorf("processType = %q, erwartet %q", vorgang.ProcessType, tt.wantProcessType)
			}
			if vorgang.ProcessData != tt.wantProcessData {
				t.Errorf("processData = %q, erwartet %q", vorgang.ProcessData, tt.wantProcessData)
			}
		})
	}
}

// Ein unbekannter Event-Typ ist ein Fehler: Ein neuer Event-Typ ohne
// Projektions-Eintrag darf nicht still unsigniert bleiben.
func TestFiskalischeProjektion_UnbekannterTypIstFehler(t *testing.T) {
	evt := e.Event{Type: "unbekannt:v1", Subject: "kassensitzung-1", Version: 1, Data: []byte(`{}`)}
	if _, _, err := FiskalischeProjektion(evt); err == nil {
		t.Fatal("erwarteter Fehler fuer unbekannten Event-Typ blieb aus")
	}
}

// Nicht parsebare Event-Daten sind ein Fehler (kein stilles Ueberspringen).
func TestFiskalischeProjektion_KaputteDatenSindFehler(t *testing.T) {
	evt := e.Event{Type: string(EventTypeZahlungKassiertV1), Subject: "kassensitzung-1/tisch-1", Version: 1, Data: []byte(`{invalid`)}
	if _, _, err := FiskalischeProjektion(evt); err == nil {
		t.Fatal("erwarteter Fehler fuer kaputte Event-Daten blieb aus")
	}
}
