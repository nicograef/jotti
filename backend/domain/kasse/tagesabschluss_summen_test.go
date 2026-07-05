//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// makeAbschlussEvent baut ein minimales Event mit dem angegebenen Typ und Data.
// Bypasses die Konstruktor-Validatoren, da ComputeAbschlussSummen keine Schema-
// Validierung durchführt — nur die summen-relevanten JSON-Felder zählen.
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

// sqlReferenzSummen berechnet dieselben drei Summen wie reporting.sql:10-43,
// indem es die JSON-Feldnamen direkt wie die kj_extract_*-SQL-Funktionen verwendet.
// Damit ist der Test unabhängig von den Go-Struct-Tags in den Event-Data-Typen.
func sqlReferenzSummen(events []e.Event) AbschlussSummen {
	var s AbschlussSummen
	for _, evt := range events {
		var d map[string]json.RawMessage
		if err := json.Unmarshal(evt.Data, &d); err != nil {
			continue
		}

		parseInt := func(key string) int {
			v, ok := d[key]
			if !ok {
				return 0
			}
			var i int
			if err := json.Unmarshal(v, &i); err != nil {
				return 0
			}
			return i
		}
		parseStr := func(key string) string {
			v, ok := d[key]
			if !ok {
				return ""
			}
			var str string
			if err := json.Unmarshal(v, &str); err != nil {
				return ""
			}
			return str
		}

		switch evt.Type {
		case string(EventTypeZahlungKassiertV1):
			// kj_extract_zahlung_cents: data->>'gesamtZahlungCents'
			s.UmsatzCents += parseInt("gesamtZahlungCents")
		case string(EventTypeStornierungErteiltV1):
			// kj_extract_stornierung_cents: data->>'gesamtStornierungCents'
			cents := parseInt("gesamtStornierungCents")
			s.UmsatzCents -= cents
			s.StornierungCents += cents
		case string(EventTypeBestellungKorrigiertV1):
			// kj_extract_korrektur_cents: data->>'gesamtCents'
			s.StornierungCents += parseInt("gesamtCents")
		case string(EventTypeDirektverkaufGetaetigtV1):
			// kj_extract_direktverkauf_cents: data->>'gesamtbetragCents'
			s.UmsatzCents += parseInt("gesamtbetragCents")
		case string(EventTypeDirektverkaufStorniertV1):
			// kj_extract_direktverkauf_storno_cents: data->>'gesamtStornierungCents'
			cents := parseInt("gesamtStornierungCents")
			s.UmsatzCents -= cents
			s.StornierungCents += cents
		case string(EventTypeGeldtransitGebuchtV1):
			// kj_extract_geldtransit_cents: einlage → +betragCents, entnahme → -betragCents
			betrag := parseInt("betragCents")
			if parseStr("richtung") == "einlage" {
				s.GeldtransitCents += betrag
			} else {
				s.GeldtransitCents -= betrag
			}
		}
	}
	return s
}

func TestComputeAbschlussSummen(t *testing.T) {
	// Shorthands für die sum-relevanten Event-Typen als Inline-Daten.
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
		return makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]interface{}{"richtung": "einlage", "betragCents": cents})
	}
	entnahme := func(cents int) e.Event {
		return makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]interface{}{"richtung": "entnahme", "betragCents": cents})
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
			name:       "Warenruecknahme reduziert Umsatz und erhoet Stornierungen",
			events:     []e.Event{zahlung(500), warenruecknahme(300)},
			wantUmsatz: 200,
			wantStorno: 300,
		},
		{
			name:       "Korrektur (geldneutral) erhoet Stornierungen ohne Umsatzwirkung",
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
			name:       "Direktverkauf erhoet Umsatz",
			events:     []e.Event{direktverkauf(300)},
			wantUmsatz: 300,
		},
		{
			name:       "Direktverkauf-Storno reduziert Umsatz und erhoet Stornierungen",
			events:     []e.Event{direktverkauf(500), dvStorno(250)},
			wantUmsatz: 250,
			wantStorno: 250,
		},
		{
			name:        "Geldtransit Einlage erhoet Geldtransit",
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
			name: "Wiederanlauf: doppelter Kassensturz und Differenz bleiben summen-neutral",
			// Erster Abschlussversuch (gescheitert): Kassensturz + Differenz committed,
			// zweiter Durchlauf liest alle Events einschliesslich der ersten Abschluss-Events.
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
			// umsatz:   1200 - 400 + 800 - 300 = 1300
			// storno:    400 + 150 + 300 = 850
			// transit:   500 - 200 = 300
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

// TestComputeAbschlussSummen_AequivalenzMitSQLReporting belegt, dass
// ComputeAbschlussSummen für dieselbe Kassensitzung exakt dieselben drei
// Summen liefert wie die kj_extract_*-Funktionen in reporting.sql:10-43.
// sqlReferenzSummen liest die JSON-Felder direkt über die SQL-Feldnamen
// (als raw JSON keys), unabhängig von den Go-Struct-Tags der Event-Data-Typen.
func TestComputeAbschlussSummen_AequivalenzMitSQLReporting(t *testing.T) {
	// Szenario mit allen sechs summen-wirksamen Typen und neutralen Events.
	events := []e.Event{
		// zahlung-kassiert: kj_extract_zahlung_cents → gesamtZahlungCents
		makeAbschlussEvent(EventTypeZahlungKassiertV1, map[string]int{"gesamtZahlungCents": 2238}),
		makeAbschlussEvent(EventTypeZahlungKassiertV1, map[string]int{"gesamtZahlungCents": 1470}),
		// stornierung-erteilt: kj_extract_stornierung_cents → gesamtStornierungCents
		makeAbschlussEvent(EventTypeStornierungErteiltV1, map[string]int{"gesamtStornierungCents": 1455}),
		// bestellung-korrigiert: kj_extract_korrektur_cents → gesamtCents
		makeAbschlussEvent(EventTypeBestellungKorrigiertV1, map[string]int{"gesamtCents": 200}),
		// direktverkauf-getaetigt: kj_extract_direktverkauf_cents → gesamtbetragCents
		makeAbschlussEvent(EventTypeDirektverkaufGetaetigtV1, map[string]int{"gesamtbetragCents": 880}),
		// direktverkauf-storniert: kj_extract_direktverkauf_storno_cents → gesamtStornierungCents
		makeAbschlussEvent(EventTypeDirektverkaufStorniertV1, map[string]int{"gesamtStornierungCents": 335}),
		// geldtransit-gebucht: kj_extract_geldtransit_cents → richtung + betragCents
		makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]interface{}{"richtung": "einlage", "betragCents": 500}),
		makeAbschlussEvent(EventTypeGeldtransitGebuchtV1, map[string]interface{}{"richtung": "entnahme", "betragCents": 150}),
		// summen-neutrale Events
		makeAbschlussEvent(EventTypeBestellungAufgenommenV1, map[string]int{"gesamtPreisCents": 999}),
		makeAbschlussEvent(EventTypeKassensturzDurchgefuehrtV1, map[string]int{"sollBestandCents": 5000}),
		makeAbschlussEvent(EventTypeDifferenzSollIstGebuchtV1, map[string]int{"betragCents": -42}),
	}

	got, err := ComputeAbschlussSummen(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ref := sqlReferenzSummen(events)

	if got != ref {
		t.Errorf("ComputeAbschlussSummen %+v != sqlReferenzSummen %+v", got, ref)
	}

	// Erwartete Werte aus der reporting.sql-Formel (manuell gerechnet):
	// Umsatz:      2238 + 1470 (zahlungen) - 1455 (storno) + 880 (dv) - 335 (dvStorno) = 2798
	// Storno:      1455 (storno) + 200 (korrektur) + 335 (dvStorno) = 1990
	// Geldtransit: 500 (einlage) - 150 (entnahme) = 350
	const wantUmsatz, wantStorno, wantTransit = 2798, 1990, 350

	if got.UmsatzCents != wantUmsatz {
		t.Errorf("UmsatzCents = %d, want %d", got.UmsatzCents, wantUmsatz)
	}
	if got.StornierungCents != wantStorno {
		t.Errorf("StornierungCents = %d, want %d", got.StornierungCents, wantStorno)
	}
	if got.GeldtransitCents != wantTransit {
		t.Errorf("GeldtransitCents = %d, want %d", got.GeldtransitCents, wantTransit)
	}
}

// TestComputeAbschlussSummen_UnparsebaresEventGibtFehler prüft, dass ein
// summen-wirksames Event mit korrupten JSON-Daten einen Fehler liefert statt
// stillschweigend übersprungen zu werden. Jeder der sechs summen-wirksamen
// Typen wird einzeln getestet.
func TestComputeAbschlussSummen_UnparsebaresEventGibtFehler(t *testing.T) {
	corruptData := json.RawMessage(`{"fehlerhaft": true`) // ungültiges JSON

	summenWirksam := []EventType{
		EventTypeZahlungKassiertV1,
		EventTypeStornierungErteiltV1,
		EventTypeBestellungKorrigiertV1,
		EventTypeDirektverkaufGetaetigtV1,
		EventTypeDirektverkaufStorniertV1,
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
				t.Fatalf("expected error for unparseable %s event, got nil", typ)
			}
		})
	}

	// Summen-neutrale Events mit korrupten Daten werden weiterhin übersprungen.
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
