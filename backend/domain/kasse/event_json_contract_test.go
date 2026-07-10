//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
)

// The event payloads are stored as JSONB and the reporting layer reads individual
// keys straight from that JSON — see the kj_extract_* SQL functions in
// database/migrations/01_initial.up.sql, sqlc/queries/reporting.sql, and the
// position parsing in repository/reporting_repo. A struct-tag rename here would
// break those queries silently (no compile error), so these tests pin both the
// JSON keys and specific field values for every event type.
//
// Obligation: whenever a new EventType constant is added, (1) add it to
// allEventTypes below, (2) write a TestEventContract_* function for it, and
// (3) add it to contractedTypes in TestEventContract_AllTypesPinned.

// allEventTypes is the canonical list of all event types in the domain.
// Must be kept in sync with the EventType constants across *_events.go files.
var allEventTypes = []EventType{
	EventTypeBestellungAufgenommenV1,
	EventTypeZahlungKassiertV1,
	EventTypeStornierungErteiltV1,
	EventTypeBestellungKorrigiertV1,
	EventTypeBestellungUmgebuchtV1,
	EventTypeKassensitzungEroeffnetV1,
	EventTypeGeldtransitGebuchtV1,
	EventTypeKassensturzDurchgefuehrtV1,
	EventTypeDifferenzSollIstGebuchtV1,
	EventTypeTagesabschlussErstelltV1,
	EventTypeDirektverkaufGetaetigtV1,
	EventTypeDirektverkaufStorniertV1,
}

// TestEventContract_AllTypesPinned ensures every known event type has a frozen
// JSON contract test. Fails when a new EventType is added without an entry here.
func TestEventContract_AllTypesPinned(t *testing.T) {
	contractedTypes := map[EventType]bool{
		EventTypeBestellungAufgenommenV1:    true,
		EventTypeZahlungKassiertV1:          true,
		EventTypeStornierungErteiltV1:       true,
		EventTypeBestellungKorrigiertV1:     true,
		EventTypeBestellungUmgebuchtV1:      true,
		EventTypeKassensitzungEroeffnetV1:   true,
		EventTypeGeldtransitGebuchtV1:       true,
		EventTypeKassensturzDurchgefuehrtV1: true,
		EventTypeDifferenzSollIstGebuchtV1:  true,
		EventTypeTagesabschlussErstelltV1:   true,
		EventTypeDirektverkaufGetaetigtV1:   true,
		EventTypeDirektverkaufStorniertV1:   true,
	}
	for _, et := range allEventTypes {
		if !contractedTypes[et] {
			t.Errorf("event type %q not in contractedTypes — add a frozen contract test", et)
		}
	}
	if len(contractedTypes) != len(allEventTypes) {
		t.Errorf("contractedTypes has %d entries but allEventTypes has %d — keep them in sync",
			len(contractedTypes), len(allEventTypes))
	}
}

// TestEventContract_SQLLiterals verifies that the EventType Go constants match the
// SQL string literals used in the kj_extract_* functions in 01_initial.up.sql.
// A mismatch would make the SQL silently return NULL instead of the expected amount.
func TestEventContract_SQLLiterals(t *testing.T) {
	checks := []struct {
		constant EventType
		sqlLit   string
	}{
		{EventTypeBestellungAufgenommenV1, "bestellung-aufgenommen:v1"},
		{EventTypeZahlungKassiertV1, "zahlung-kassiert:v1"},
		{EventTypeStornierungErteiltV1, "stornierung-erteilt:v1"},
		{EventTypeBestellungKorrigiertV1, "bestellung-korrigiert:v1"},
		{EventTypeKassensitzungEroeffnetV1, "kassensitzung-eroeffnet:v1"},
		{EventTypeGeldtransitGebuchtV1, "geldtransit-gebucht:v1"},
		{EventTypeDifferenzSollIstGebuchtV1, "differenz-soll-ist-gebucht:v1"},
		{EventTypeDirektverkaufGetaetigtV1, "direktverkauf-getaetigt:v1"},
		{EventTypeDirektverkaufStorniertV1, "direktverkauf-storniert:v1"},
	}
	for _, c := range checks {
		if string(c.constant) != c.sqlLit {
			t.Errorf("EventType constant %q != SQL literal %q", c.constant, c.sqlLit)
		}
	}
}

// TestEventContract_GeldtransitRichtung pins the "einlage"/"entnahme" string
// literals used by kj_extract_geldtransit_cents in 01_initial.up.sql (D10).
// Prüft sowohl den JSON-Key-Namen als auch den exakten Wert gegenüber den
// Domain-Konstanten, damit Wert-Drift gegenüber den SQL-Literalen auffällt.
func TestEventContract_GeldtransitRichtung(t *testing.T) {
	for _, richtung := range []string{GeldtransitRichtungEinlage, GeldtransitRichtungEntnahme} {
		d := GeldtransitGebuchtV1Data{Richtung: richtung}
		raw, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("marshal GeldtransitGebuchtV1Data: %v", err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal to map: %v", err)
		}
		rawVal, ok := m["richtung"]
		if !ok {
			t.Errorf("richtung=%q: key \"richtung\" missing from JSON", richtung)
			continue
		}
		var got string
		if err := json.Unmarshal(rawVal, &got); err != nil {
			t.Fatalf("unmarshal richtung value: %v", err)
		}
		if got != richtung {
			t.Errorf("richtung JSON value: got %q, want %q", got, richtung)
		}
	}
}

// --- helpers ---

func unmarshalJSON(t *testing.T, src string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(src), dst); err != nil {
		t.Fatalf("unmarshal %T: %v", dst, err)
	}
}

func assertField(t *testing.T, label string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", label, got, want)
	}
}

// assertJSONKeyPresent marshals v and checks that jsonKey is present in the output.
// Schützt gegen stille Tag-Umbenennung bei Feldern, die im Literal mit Zero-Value
// gepinnt sind (z. B. kommentar: "") — ein fehlendes JSON-Tag wäre sonst unsichtbar.
func assertJSONKeyPresent(t *testing.T, v any, jsonKey string) {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, ok := m[jsonKey]; !ok {
		t.Errorf("JSON key %q missing from %T output — struct tag renamed?", jsonKey, v)
	}
}

// positionLiteral is the frozen position JSON used by each event contract test.
// Changing any key here means the corresponding SQL or frontend schema must also change.
const positionLiteral = `{
	"positionId":       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	"varianteId":       7,
	"produktName":      "Cola",
	"varianteName":     "0,5l",
	"kategorie":        "getraenk",
	"steuersatz":       "regel",
	"einzelpreisCents": 350,
	"menge":            2
}`

func assertPosition(t *testing.T, p PositionEventData) {
	t.Helper()
	assertField(t, "positionId", p.PositionID, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	assertField(t, "varianteId", p.VarianteID, 7)
	assertField(t, "einzelpreisCents", p.EinzelpreisCents, 350)
	assertField(t, "menge", p.Menge, 2)
	assertField(t, "steuersatz", p.Steuersatz, "regel")
	assertField(t, "produktName", p.ProduktName, "Cola")
	assertField(t, "varianteName", p.VarianteName, "0,5l")
}

// --- frozen contract tests per event type ---

func TestEventContract_BestellungAufgenommenV1(t *testing.T) {
	const lit = `{
		"bestellungId":    "11111111-1111-4111-8111-111111111111",
		"positionen":      [` + positionLiteral + `],
		"gesamtPreisCents": 700,
		"kommentar":       ""
	}`
	var data BestellungAufgenommenV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "bestellungId", data.BestellungID, "11111111-1111-4111-8111-111111111111")
	assertField(t, "gesamtPreisCents", data.GesamtPreisCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_ZahlungKassiertV1(t *testing.T) {
	const lit = `{
		"zahlungId":          "22222222-2222-4222-8222-222222222222",
		"positionen":         [` + positionLiteral + `],
		"gesamtZahlungCents": 700,
		"kommentar":          ""
	}`
	var data ZahlungKassiertV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "zahlungId", data.ZahlungID, "22222222-2222-4222-8222-222222222222")
	assertField(t, "gesamtZahlungCents", data.GesamtZahlungCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_StornierungErteiltV1(t *testing.T) {
	const lit = `{
		"stornierungId":          "33333333-3333-4333-8333-333333333333",
		"zahlungId":              "22222222-2222-4222-8222-222222222222",
		"positionen":             [` + positionLiteral + `],
		"gesamtStornierungCents": 700,
		"kommentar":              "Gast hat reklamiert"
	}`
	var data StornierungErteiltV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "stornierungId", data.StornierungID, "33333333-3333-4333-8333-333333333333")
	assertField(t, "zahlungId", data.ZahlungID, "22222222-2222-4222-8222-222222222222")
	assertField(t, "gesamtStornierungCents", data.GesamtStornierungCents, 700)
	assertField(t, "kommentar", data.Kommentar, "Gast hat reklamiert")
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
}

func TestEventContract_BestellungKorrigiertV1(t *testing.T) {
	const lit = `{
		"korrekturId": "44444444-4444-4444-8444-444444444444",
		"positionen":  [` + positionLiteral + `],
		"gesamtCents": 700,
		"kommentar":   ""
	}`
	var data BestellungKorrigiertV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "korrekturId", data.KorrekturID, "44444444-4444-4444-8444-444444444444")
	assertField(t, "gesamtCents", data.GesamtCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_BestellungUmgebuchtV1(t *testing.T) {
	const lit = `{
		"umbuchungId":  "55555555-5555-4555-8555-555555555555",
		"quellTischId": 3,
		"zielTischId":  7,
		"positionen":   [` + positionLiteral + `],
		"gesamtCents":  700,
		"kommentar":    ""
	}`
	var data BestellungUmgebuchtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "umbuchungId", data.UmbuchungID, "55555555-5555-4555-8555-555555555555")
	assertField(t, "quellTischId", data.QuellTischID, 3)
	assertField(t, "zielTischId", data.ZielTischID, 7)
	assertField(t, "gesamtCents", data.GesamtCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_KassensitzungEroeffnetV1(t *testing.T) {
	const lit = `{
		"datum":        "2026-07-07",
		"bezeichnung":  "Maihock 2026",
		"betragCents":  5000,
		"eroeffnetVon": 1
	}`
	var data KassensitzungEroeffnetV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "datum", data.Datum, "2026-07-07")
	assertField(t, "bezeichnung", data.Bezeichnung, "Maihock 2026")
	assertField(t, "betragCents", data.BetragCents, 5000)
	assertField(t, "eroeffnetVon", data.EroeffnetVon, 1)
}

func TestEventContract_GeldtransitGebuchtV1(t *testing.T) {
	const lit = `{
		"geldtransitId": "77777777-7777-4777-8777-777777777777",
		"richtung":      "einlage",
		"betragCents":   10000,
		"kommentar":     "Wechselgeld nachgefüllt",
		"gebuchtVon":    2
	}`
	var data GeldtransitGebuchtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "geldtransitId", data.GeldtransitID, "77777777-7777-4777-8777-777777777777")
	assertField(t, "richtung", data.Richtung, "einlage")
	assertField(t, "betragCents", data.BetragCents, 10000)
	assertField(t, "gebuchtVon", data.GebuchtVon, 2)
}

func TestEventContract_KassensturzDurchgefuehrtV1(t *testing.T) {
	const lit = `{
		"sollBestandCents": 15000,
		"istBestandCents":  14500,
		"differenzCents":   -500,
		"durchgefuehrtVon": 1
	}`
	var data KassensturzDurchgefuehrtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "sollBestandCents", data.SollBestandCents, 15000)
	assertField(t, "istBestandCents", data.IstBestandCents, 14500)
	assertField(t, "differenzCents", data.DifferenzCents, -500)
	assertField(t, "durchgefuehrtVon", data.DurchgefuehrtVon, 1)
}

func TestEventContract_DifferenzSollIstGebuchtV1(t *testing.T) {
	const lit = `{
		"betragCents": -500,
		"gebuchtVon":  1
	}`
	var data DifferenzSollIstGebuchtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "betragCents", data.BetragCents, -500)
	assertField(t, "gebuchtVon", data.GebuchtVon, 1)
}

func TestEventContract_TagesabschlussErstelltV1(t *testing.T) {
	const lit = `{
		"zNr":               3,
		"zeitraumVon":       "2026-07-07T08:00:00Z",
		"zeitraumBis":       "2026-07-07T22:00:00Z",
		"umsatzGesamtCents": 250000,
		"stornierungCents":  500,
		"geldtransitCents":  10000,
		"erstelltVon":       1
	}`
	var data TagesabschlussErstelltV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "zNr", data.ZNr, 3)
	assertField(t, "umsatzGesamtCents", data.UmsatzGesamtCents, 250000)
	assertField(t, "stornierungCents", data.StornierungCents, 500)
	assertField(t, "geldtransitCents", data.GeldtransitCents, 10000)
	assertField(t, "erstelltVon", data.ErstelltVon, 1)
	if data.ZeitraumVon.IsZero() {
		t.Error("zeitraumVon must not be zero after unmarshal")
	}
	if data.ZeitraumBis.IsZero() {
		t.Error("zeitraumBis must not be zero after unmarshal")
	}
}

func TestEventContract_DirektverkaufGetaetigtV1(t *testing.T) {
	const lit = `{
		"verkaufId":         "88888888-8888-4888-8888-888888888888",
		"positionen":        [` + positionLiteral + `],
		"gesamtbetragCents": 700,
		"kommentar":         ""
	}`
	var data DirektverkaufGetaetigtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "verkaufId", data.VerkaufID, "88888888-8888-4888-8888-888888888888")
	assertField(t, "gesamtbetragCents", data.GesamtbetragCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
}

func TestEventContract_DirektverkaufStorniertV1(t *testing.T) {
	const lit = `{
		"stornierungId":          "99999999-9999-4999-8999-999999999999",
		"verkaufId":              "88888888-8888-4888-8888-888888888888",
		"positionen":             [` + positionLiteral + `],
		"gesamtStornierungCents": 700,
		"kommentar":              "Bestellung falsch"
	}`
	var data DirektverkaufStorniertV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "stornierungId", data.StornierungID, "99999999-9999-4999-8999-999999999999")
	assertField(t, "verkaufId", data.VerkaufID, "88888888-8888-4888-8888-888888888888")
	assertField(t, "gesamtStornierungCents", data.GesamtStornierungCents, 700)
	assertField(t, "kommentar", data.Kommentar, "Bestellung falsch")
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
}
