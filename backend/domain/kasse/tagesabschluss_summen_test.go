package kasse

import (
	"encoding/json"
	"testing"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// makeAbschlussEvent baut ein minimales Event ohne Konstruktor-Validierung —
// ComputeAbschlussSummen liest nur die summen-relevanten JSON-Felder.
func makeAbschlussEvent(typ EventType, data any) e.Event {
	raw, err := json.Marshal(data)
	if err != nil {
		panic("makeAbschlussEvent: " + err.Error())
	}
	return e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(typ),
		Time:     time.Now().UTC(),
		Subject:  testSubject,
		Version:  1,
		Data:     raw,
	}
}

func TestComputeAbschlussSummen(t *testing.T) {
	zahlung := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeZahlungKassiertV1, map[string]int{"gesamtZahlungCents": cents})
	}
	warenruecknahme := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeStornierungErteiltV1, map[string]int{"gesamtStornierungCents": cents})
	}
	korrektur := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeBestellungKorrigiertV1, map[string]int{"gesamtCents": cents})
	}
	direktverkauf := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeDirektverkaufGetaetigtV1, map[string]int{"gesamtbetragCents": cents})
	}
	dvStorno := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeDirektverkaufStorniertV1, map[string]int{"gesamtStornierungCents": cents})
	}
	einlage := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]any{"richtung": "einlage", "betragCents": cents})
	}
	entnahme := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]any{"richtung": "entnahme", "betragCents": cents})
	}
	pfandrueckgabe := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypePfandZurueckgegebenV1, map[string]int{"gesamtbetragCents": cents})
	}
	pfandStorno := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypePfandrueckgabeStorniertV1, map[string]int{"gesamtStornierungCents": cents})
	}
	// Summen-neutrale Events
	bestellung := makeAbschlussEvent(EventTypeBestellungAufgenommenV1, map[string]int{"gesamtPreisCents": 500})
	umbuchung := makeAbschlussEvent(EventTypeBestellungUmgebuchtV1, map[string]int{"gesamtCents": 300})
	kassensturz := makeAbschlussEvent(EventTypeKassensturzDurchgefuehrtV1, map[string]int{"sollBestandCents": 1000})
	differenz := makeAbschlussEvent(EventTypeDifferenzSollIstGebuchtV1, map[string]int{"betragCents": -50})

	cases := []struct {
		name                                string
		events                              []e.Event
		wantUmsatz, wantStorno, wantTransit int
	}{
		{
			name: "leere Kassensitzung ergibt alle Nullen",
		},
		{
			name:   "Bestellungen ohne Zahlung sind summen-neutral",
			events: []e.Event{bestellung},
		},
		{
			name:       "Zahlung ueber mehrere Steuersaetze addiert Gesamtbetrag zum Umsatz",
			events:     []e.Event{zahlung(500), zahlung(300)},
			wantUmsatz: 800,
		},
		{
			name:       "Warenruecknahme reduziert Umsatz und erhoeht Stornierungen",
			events:     []e.Event{zahlung(500), warenruecknahme(300)},
			wantUmsatz: 200,
			wantStorno: 300,
		},
		{
			name:       "Korrektur (geldneutral) erhoeht Stornierungen ohne Umsatzwirkung",
			events:     []e.Event{zahlung(500), korrektur(200)},
			wantUmsatz: 500,
			wantStorno: 200,
		},
		{
			name:       "Umbuchung ist summen-neutral",
			events:     []e.Event{zahlung(500), umbuchung},
			wantUmsatz: 500,
		},
		{
			name:       "Direktverkauf erhoeht Umsatz",
			events:     []e.Event{direktverkauf(300)},
			wantUmsatz: 300,
		},
		{
			name:       "Direktverkauf-Storno reduziert Umsatz und erhoeht Stornierungen",
			events:     []e.Event{direktverkauf(500), dvStorno(250)},
			wantUmsatz: 250,
			wantStorno: 250,
		},
		{
			name:       "Pfandrueckgabe mindert Umsatz",
			events:     []e.Event{zahlung(500), pfandrueckgabe(200)},
			wantUmsatz: 300,
		},
		{
			name:       "Storno einer Pfandrueckgabe hebt die Umsatzminderung auf",
			events:     []e.Event{zahlung(500), pfandrueckgabe(200), pfandStorno(200)},
			wantUmsatz: 500,
		},
		{
			name:        "Geldtransit Einlage erhoeht Geldtransit",
			events:      []e.Event{einlage(1000)},
			wantTransit: 1000,
		},
		{
			name:        "Geldtransit Entnahme mindert Geldtransit",
			events:      []e.Event{einlage(1000), entnahme(400)},
			wantTransit: 600,
		},
		{
			name:       "Differenzbuchung ist summen-neutral",
			events:     []e.Event{zahlung(500), differenz},
			wantUmsatz: 500,
		},
		{
			name:       "Kassensturz ist summen-neutral",
			events:     []e.Event{zahlung(500), kassensturz},
			wantUmsatz: 500,
		},
		{
			name:       "Wiederanlauf: doppelter Kassensturz und Differenz bleiben summen-neutral",
			events:     []e.Event{zahlung(500), kassensturz, differenz, kassensturz, differenz},
			wantUmsatz: 500,
		},
		{
			name: "Mehrstufige Sitzung: alle sechs Typen korrekt aggregiert",
			events: []e.Event{
				zahlung(1200),
				warenruecknahme(400),
				korrektur(150),
				direktverkauf(800),
				dvStorno(300),
				einlage(500),
				entnahme(200),
				// neutral
				bestellung, umbuchung, kassensturz, differenz,
			},
			// erwartet: Umsatz 1200 - 400 + 800 - 300 ergibt 1300,
			// Storno 400 + 150 + 300 ergibt 850, Transit 500 - 200 ergibt 300.
			wantUmsatz:  1300,
			wantStorno:  850,
			wantTransit: 300,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ComputeAbschlussSummen(tc.events)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got.UmsatzCents != tc.wantUmsatz {
				t.Errorf("UmsatzCents = %d, want %d", got.UmsatzCents, tc.wantUmsatz)
			}
			if got.StornierungCents != tc.wantStorno {
				t.Errorf("StornierungCents = %d, want %d", got.StornierungCents, tc.wantStorno)
			}
			if got.GeldtransitCents != tc.wantTransit {
				t.Errorf("GeldtransitCents = %d, want %d", got.GeldtransitCents, tc.wantTransit)
			}
		})
	}
}

// Ein summen-wirksames Event mit korrupten JSON-Daten muss einen Fehler liefern statt still
// übersprungen zu werden.
func TestComputeAbschlussSummen_UnparsebaresEventGibtFehler(t *testing.T) {
	corruptData := json.RawMessage(`{"fehlerhaft": true`) // ungültiges JSON

	summenWirksam := []EventType{
		EventTypeZahlungKassiertV1,
		EventTypeZahlungKassiertV2,
		EventTypeStornierungErteiltV1,
		EventTypeBestellungKorrigiertV1,
		EventTypeDirektverkaufGetaetigtV1,
		EventTypeDirektverkaufGetaetigtV2,
		EventTypeDirektverkaufStorniertV1,
		EventTypePfandZurueckgegebenV1,
		EventTypePfandrueckgabeStorniertV1,
		EventTypeGeldtransitGebuchtV1,
	}

	for _, typ := range summenWirksam {
		t.Run(string(typ), func(t *testing.T) {
			evt := e.Event{
				ID:      1,
				UserID:  1,
				Type:    string(typ),
				Time:    time.Now().UTC(),
				Subject: testSubject,
				Version: 1,
				Data:    corruptData,
			}
			_, err := ComputeAbschlussSummen([]e.Event{evt})
			if err == nil {
				t.Errorf("expected error for unparseable %s event, got nil", typ)
			}
		})
	}

	t.Run("summen-neutrale Events werden übersprungen", func(t *testing.T) {
		evt := e.Event{
			ID:      1,
			UserID:  1,
			Type:    string(EventTypeBestellungAufgenommenV1),
			Time:    time.Now().UTC(),
			Subject: testSubject,
			Version: 1,
			Data:    corruptData,
		}
		got, err := ComputeAbschlussSummen([]e.Event{evt})
		if err != nil {
			t.Fatalf("expected no error for neutral event, got %v", err)
		}
		if got != (AbschlussSummen{}) {
			t.Errorf("expected zero summen for neutral event, got %+v", got)
		}
	})
}
