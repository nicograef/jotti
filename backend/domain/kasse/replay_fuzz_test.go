//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
	"time"
	"unicode/utf8"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// FuzzApplyEvent prüft die Replay-Kante des Kassenjournals: ApplyEvent verarbeitet die
// persistierte Event-Data (JSONB) beim Wiederaufbau der tisch_sessions-Projektion. Zu
// halten ist Panic-Freiheit, egal wie kaputt das JSON ist — ein Panic legte den
// Projektions-Rebuild (make rebuild-projections) und jede Sitzung mit diesem Event
// dauerhaft lahm. Fachlich falsche, aber wohlgeformte Payloads dürfen einen Fehler liefern.
// Der Seed-Korpus stammt aus den eingefrorenen Event-JSON-Contracts
// (event_json_contract_test.go).
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
			string(EventTypeBestellungUmgebuchtV1),
			`{"umbuchungId":"66666666-6666-4666-8666-666666666666","quellTischId":1,"zielTischId":2,"positionen":[` + fuzzPositionLiteral + `],"gesamtCents":700,"kommentar":""}`,
		},
	}
	for _, s := range seeds {
		f.Add(s.typ, []byte(s.data))
	}
	f.Add(string(EventTypeBestellungAufgenommenV1), []byte(`{}`))
	f.Add(string(EventTypeZahlungKassiertV1), []byte(`{"gesamtZahlungCents":-5}`))
	f.Add("unbekannt:v1", []byte(`{"foo":"bar"}`))

	f.Fuzz(func(t *testing.T, typ string, data []byte) {
		subject := TischSessionSubject(1, 1)

		// Ausgangszustand: eine gültige Bestellung (Saldo 700) — so trifft das Fuzz-Event
		// einen echten Vorzustand statt nur den Nullwert.
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

		// Bei Fehler zählt nur Panic-Freiheit; der Rückgabewert wird vom Aufrufer verworfen.
		next, applyErr := ApplyEvent(basis, evt)
		if applyErr != nil {
			return
		}

		// Invariante 1 — jede projizierte Position hält eine positive Menge; Menge <= 0 wäre
		// ein Projektionsfehler. Greift nur bei gültigen Eingabe-Positionen (Menge > 0,
		// PositionID gesetzt) — eine Menge-0-Position liegt außerhalb des validierten Korpus
		// und reicht ihren Nulleintrag erwartungsgemäß durch.
		if eingabePositionenGueltig(evt.Data) {
			for _, pos := range next.UnbezahltePositionen {
				if pos.Menge <= 0 {
					t.Fatalf("unbezahlte Position mit nicht-positiver Menge %d (%s) nach %s", pos.Menge, pos.PositionID, typ)
				}
			}
		}

		// Invariante 2 — der offene Betrag darf nicht unter 0 fallen. Eine Minderung, die
		// größer als der Vorzustands-Saldo ist, liegt außerhalb des validierten Korpus (im
		// Betrieb kann nie mehr kassiert werden als offen ist) und ist ausgenommen.
		if minderung := saldoMinderung(typ, evt.Data); minderung <= basis.SaldoCents && next.SaldoCents < 0 {
			t.Fatalf("negativer Saldo %d nach %s (Basis %d, Minderung %d)", next.SaldoCents, typ, basis.SaldoCents, minderung)
		}

		// Invariante 3 — SaldoCents ist stets Σ(EinzelpreisCents × Menge) über die
		// unbezahlten Positionen; diese Ableitung ist die einzige Quelle der Wahrheit.
		var erwarteterSaldo int
		for _, pos := range next.UnbezahltePositionen {
			erwarteterSaldo += pos.EinzelpreisCents * pos.Menge
		}
		if next.SaldoCents != erwarteterSaldo {
			t.Fatalf("SaldoCents %d weicht von Σ(EinzelpreisCents × Menge) %d ab nach %s", next.SaldoCents, erwarteterSaldo, typ)
		}
	})
}

// saldoMinderung liest den Saldo-mindernden Betrag aus der Payload, um die Saldo-Invariante
// auf den realistischen Rahmen einzugrenzen. Nicht-mindernde Events liefern 0.
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

// eingabePositionenGueltig meldet, ob jede Position der Payload eine PositionID und eine
// Menge > 0 hat; nur dann greift die Positions-Mengen-Invariante.
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

// FuzzPositionEventDataRoundtrip prüft: Eine PositionEventData muss serialisiert und wieder
// eingelesen feldgleich sein. Das Positions-JSONB speist DSFinV-K-Export und die
// SQL-Reporting-Extraktoren — ein stiller Feldverlust verfälscht Bon- und Steuersummen.
func FuzzPositionEventDataRoundtrip(f *testing.F) {
	f.Add("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 7, "Cola", "0,5l", "getraenk", "regel", 350, 2)
	f.Add("", 0, "", "", "", "", 0, 0)
	f.Add("x", -1, "Ünïcödé \t\n", "\"quote;\"", "essen", "ermaessigt", -999, 1000000)

	f.Fuzz(func(t *testing.T, posID string, varianteID int, produktName, varianteName, kategorie, steuersatz string, einzelpreis, menge int) {
		// Persistierte Event-Strings sind stets gültiges UTF-8. Ungültige Byte-Folgen ersetzt
		// json.Marshal durch U+FFFD — stdlib-Verhalten, keine jotti-Eigenschaft, außerhalb
		// des realistischen Korpus.
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
