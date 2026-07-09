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
		evt := e.Event{
			ID:       1,
			UserID:   1,
			UserName: "fuzz",
			Type:     typ,
			Time:     time.Unix(0, 0).UTC(),
			Subject:  TischSessionSubject(1, 1),
			Version:  1,
			Data:     json.RawMessage(data),
		}

		// Kein Panic: der Rückgabewert bei Fehler wird vom Aufrufer verworfen, daher
		// prüfen wir hier nur die Panic-Freiheit der Replay-Kante.
		_, _ = ApplyEvent(TischSession{Subject: evt.Subject}, evt)
	})
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
