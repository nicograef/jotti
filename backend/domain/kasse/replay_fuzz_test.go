//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
	"time"
	"unicode/utf8"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// FuzzApplyEvent prüft die Replay-Kante des Kassenjournals: ApplyEvent verarbeitet
// die persistierte Event-Data (JSONB) beim Wiederaufbau der tisch_sessions-Projektion.
// Der Fuzzer wirft beliebige (auch defekte) JSON-Payloads gegen jeden Event-Typ. Die
// zu haltende Eigenschaft ist Robustheit: kein Panic, egal wie kaputt das JSON ist —
// ein Panic beim Replay würde den Projektions-Rebuild (make rebuild-projections) und
// jede Sitzung mit diesem Event dauerhaft lahmlegen. Fachlich falsche, aber
// wohlgeformte Payloads dürfen einen Fehler liefern; das ist der erwartete Pfad.
//
// Der Seed-Korpus stammt aus den echten, eingefrorenen Event-JSON-Contracts
// (event_json_contract_test.go) je Event-Typ und läuft bei jedem `go test` mit.
func FuzzApplyEvent(f *testing.F) {
	seeds := []struct {
		typ  string
		data string
	}{
		{
			string(EventTypeBestellungAufgenommenV1),
			`{"bestellungId":"11111111-1111-4111-8111-111111111111","positionen":[` + fuzzPositionLiteral + `],"gesamtPreisCents":700,"kommentar":""}`,
		},
		{
			string(EventTypeZahlungKassiertV1),
			`{"zahlungId":"22222222-2222-4222-8222-222222222222","positionen":[` + fuzzPositionLiteral + `],"gesamtZahlungCents":700,"kommentar":""}`,
		},
		{
			string(EventTypeStornierungErteiltV1),
			`{"stornierungId":"33333333-3333-4333-8333-333333333333","zahlungId":"22222222-2222-4222-8222-222222222222","positionen":[` + fuzzPositionLiteral + `],"gesamtStornierungCents":700,"kommentar":""}`,
		},
		{
			string(EventTypeBestellungKorrigiertV1),
			`{"korrekturId":"44444444-4444-4444-8444-444444444444","positionen":[` + fuzzPositionLiteral + `],"gesamtCents":700,"kommentar":""}`,
		},
		{
			string(EventTypeAusgabeBestaetigtV1),
			`{"ausgabeId":"55555555-5555-4555-8555-555555555555","positionen":[` + fuzzPositionLiteral + `],"kommentar":""}`,
		},
		{
			string(EventTypeBestellungUmgebuchtV1),
			`{"umbuchungId":"66666666-6666-4666-8666-666666666666","quellTischId":1,"zielTischId":2,"positionen":[` + fuzzPositionLiteral + `],"gesamtCents":700,"kommentar":""}`,
		},
	}
	for _, s := range seeds {
		f.Add(s.typ, []byte(s.data))
	}
	// Zusätzliche Kanten: leeres Objekt, defektes JSON, unbekannter Typ.
	f.Add(string(EventTypeBestellungAufgenommenV1), []byte(`{}`))
	f.Add(string(EventTypeZahlungKassiertV1), []byte(`{"gesamtZahlungCents":-5}`))
	f.Add("unbekannt:v1", []byte(`{"foo":"bar"}`))

	f.Fuzz(func(t *testing.T, typ string, data []byte) {
		subject := TischSessionSubject(1, 1)

		// Ausgangszustand: eine gültige Bestellung stellt einen realistischen,
		// nicht-leeren Tisch her (Saldo 700, eine Position mit Menge 2). So testet
		// der Fuzzer das folgende Event auf einem echten Vorzustand — nicht nur auf
		// dem Nullwert — und die Invarianten haben einen Bezugspunkt.
		basis, err := ApplyEvent(TischSession{Subject: subject}, e.Event{
			ID: 1, UserID: 1, UserName: "fuzz", Version: 1,
			Type:    string(EventTypeBestellungAufgenommenV1),
			Time:    time.Unix(0, 0).UTC(),
			Subject: subject,
			Data:    json.RawMessage(`{"bestellungId":"11111111-1111-4111-8111-111111111111","positionen":[` + fuzzPositionLiteral + `],"gesamtPreisCents":700,"kommentar":""}`),
		})
		if err != nil {
			t.Fatalf("Basis-Bestellung muss anwendbar sein: %v", err)
		}

		evt := e.Event{
			ID:       2,
			UserID:   1,
			UserName: "fuzz",
			Type:     typ,
			Time:     time.Unix(0, 0).UTC(),
			Subject:  subject,
			Version:  1,
			Data:     json.RawMessage(data),
		}

		// Kein Panic: der Rückgabewert bei Fehler wird vom Aufrufer verworfen, daher
		// prüfen wir bei Fehlern nur die Panic-Freiheit der Replay-Kante. Ein
		// fachlich falsches, aber wohlgeformtes Event darf einen Fehler liefern.
		next, applyErr := ApplyEvent(basis, evt)
		if applyErr != nil {
			return
		}

		// Ab hier: das Event wurde erfolgreich angewendet. Die folgenden
		// semantischen Invarianten müssen für jeden erfolgreichen Replay gelten.

		// Invariante 1 — Positionsmengen konsistent: Reduzierungen und
		// Akkumulationen halten jede projizierte Position bei einer echten,
		// positiven Menge; eine Position mit Menge <= 0 wäre ein Projektionsfehler
		// (nicht-entfernter Nulleintrag oder Vorzeichenfehler). Die Prüfung greift
		// nur, wenn die Positionen des angewendeten Events selbst gültig sind
		// (Menge > 0, gesetzte PositionID) — eine Menge-0-Position in der Payload
		// liegt außerhalb des validierten Korpus und würde ihren Nulleintrag
		// erwartungsgemäß durchreichen.
		if eingabePositionenGueltig(evt.Data) {
			for _, pos := range next.UnbezahltePositionen {
				if pos.Menge <= 0 {
					t.Fatalf("unbezahlte Position mit nicht-positiver Menge %d (%s) nach %s", pos.Menge, pos.PositionID, typ)
				}
			}
			for _, pos := range next.AusstehendePositionen {
				if pos.Menge <= 0 {
					t.Fatalf("ausstehende Position mit nicht-positiver Menge %d (%s) nach %s", pos.Menge, pos.PositionID, typ)
				}
			}
		}

		// Invariante 2 — Saldo nie negativ: Der offene Betrag eines Tisches darf
		// nicht unter 0 fallen. Saldo-mindernde Events (Zahlung/Korrektur/Umbuchung
		// als Abgang) tragen einen validierten, nicht-negativen Betrag, der den
		// offenen Betrag nicht übersteigt. Der Fuzzer speist jedoch beliebige
		// Beträge ein: eine Minderung, die größer als der Vorzustands-Saldo ist,
		// liegt außerhalb des validierten Korpus (im Betrieb kann nie mehr kassiert
		// werden als offen ist) und wird von der Prüfung ausgenommen. Innerhalb des
		// realistischen Rahmens muss der Saldo aber nicht-negativ bleiben.
		if minderung := saldoMinderung(typ, evt.Data); minderung <= basis.SaldoCents && next.SaldoCents < 0 {
			t.Fatalf("negativer Saldo %d nach %s (Basis %d, Minderung %d)", next.SaldoCents, typ, basis.SaldoCents, minderung)
		}
	})
}

// saldoMinderung liest den Saldo-mindernden Betrag eines Events aus der Payload,
// um die Saldo-Invariante auf den realistischen Rahmen (Minderung <= offener
// Betrag) einzugrenzen. Nicht-mindernde Events liefern 0.
func saldoMinderung(typ string, data json.RawMessage) int {
	switch typ {
	case string(EventTypeZahlungKassiertV1):
		var d ZahlungKassiertV1Data
		if json.Unmarshal(data, &d) == nil {
			return d.GesamtZahlungCents
		}
	case string(EventTypeBestellungKorrigiertV1):
		var d BestellungKorrigiertV1Data
		if json.Unmarshal(data, &d) == nil {
			return d.GesamtCents
		}
	case string(EventTypeBestellungUmgebuchtV1):
		var d BestellungUmgebuchtV1Data
		if json.Unmarshal(data, &d) == nil && d.QuellTischID == 1 {
			// Nur der Abgang (Quelltisch == Subjekt-Tisch 1) mindert den Saldo.
			return d.GesamtCents
		}
	}
	return 0
}

// eingabePositionenGueltig prüft, ob die Positionen in der Event-Payload einem
// gültigen Vorgang entsprechen (jede Position hat eine PositionID und eine Menge
// > 0). Nur dann greift die Positions-Mengen-Invariante; Menge-0- oder
// PositionID-lose Positionen liegen außerhalb des validierten Schreibpfads.
func eingabePositionenGueltig(data json.RawMessage) bool {
	var payload struct {
		Positionen []PositionEventData `json:"positionen"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	for _, pos := range payload.Positionen {
		if pos.PositionID == "" || pos.Menge <= 0 {
			return false
		}
	}
	return true
}

// FuzzPositionEventDataRoundtrip prüft die Persistenz-Roundtrip-Eigenschaft der
// Positions-Payload: Eine PositionEventData, die serialisiert und wieder eingelesen
// wird, muss feldgleich sein. Das JSONB der Positionen ist die Quelle sowohl für den
// DSFinV-K-Export als auch für die SQL-Reporting-Extraktoren — ein stiller
// Feldverlust hier verfälscht Bon-Summen und Steueraufteilung.
func FuzzPositionEventDataRoundtrip(f *testing.F) {
	f.Add("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 7, "Cola", "0,5l", "getraenk", "regel", 350, 2)
	f.Add("", 0, "", "", "", "", 0, 0)
	f.Add("x", -1, "Ünïcödé \t\n", "\"quote;\"", "essen", "ermaessigt", -999, 1000000)

	f.Fuzz(func(t *testing.T, posID string, varianteID int, produktName, varianteName, kategorie, steuersatz string, einzelpreis, menge int) {
		// Persistierte Event-Strings sind stets gültiges UTF-8 (validierte Eingaben,
		// UTF-8-JSONB in Postgres). Ungültige Byte-Folgen ersetzt json.Marshal durch
		// U+FFFD — das ist stdlib-Verhalten, keine jotti-Eigenschaft, und außerhalb
		// des realistischen Korpus. Solche Eingaben überspringen wir.
		for _, s := range []string{posID, produktName, varianteName, kategorie, steuersatz} {
			if !utf8.ValidString(s) {
				t.Skip()
			}
		}
		orig := PositionEventData{
			PositionID:       posID,
			VarianteID:       varianteID,
			ProduktName:      produktName,
			VarianteName:     varianteName,
			Kategorie:        kategorie,
			Steuersatz:       steuersatz,
			EinzelpreisCents: einzelpreis,
			Menge:            menge,
		}
		raw, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal PositionEventData: %v", err)
		}
		var back PositionEventData
		if err := json.Unmarshal(raw, &back); err != nil {
			t.Fatalf("unmarshal PositionEventData %q: %v", raw, err)
		}
		if back != orig {
			t.Fatalf("Roundtrip-Verlust:\n orig = %+v\n back = %+v\n json = %s", orig, back, raw)
		}
	})
}

// fuzzPositionLiteral spiegelt die eingefrorene Position aus event_json_contract_test.go.
const fuzzPositionLiteral = `{"positionId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","varianteId":7,"produktName":"Cola","varianteName":"0,5l","kategorie":"getraenk","steuersatz":"regel","einzelpreisCents":350,"menge":2}`
