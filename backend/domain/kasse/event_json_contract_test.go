package kasse

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Reporting SQL reads event JSON keys directly (kj_extract_* in 01_initial.up.sql, sqlc/queries/
// reporting.sql, repository/reporting_repo), so these tests pin the keys a tag rename would break.
// A new EventType needs an entry in allEventTypes, a TestEventContract_* func and a byte-identity row.

// allEventTypes must list every EventType constant of the package; TestEventContract_AllTypesPinned checks it.
var allEventTypes = []EventType{
	EventTypeBestellungAufgenommenV1,
	EventTypeZahlungKassiertV1,
	EventTypeZahlungKassiertV2,
	EventTypeStornierungErteiltV1,
	EventTypeBestellungKorrigiertV1,
	EventTypeBestellungUmgebuchtV1,
	EventTypeKassensitzungEroeffnetV1,
	EventTypeGeldtransitGebuchtV1,
	EventTypeKassensturzDurchgefuehrtV1,
	EventTypeDifferenzSollIstGebuchtV1,
	EventTypeTagesabschlussErstelltV1,
	EventTypeDirektverkaufGetaetigtV1,
	EventTypeDirektverkaufGetaetigtV2,
	EventTypeDirektverkaufStorniertV1,
	EventTypePfandZurueckgegebenV1,
	EventTypePfandrueckgabeStorniertV1,
}

// Fails when an EventType constant in a *_events.go file is missing from allEventTypes (or vice
// versa): the constants are read from the package source, so a new event type cannot slip past
// the contract tests below.
func TestEventContract_AllTypesPinned(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	var deklariert []string
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)
				typ, ok := vs.Type.(*ast.Ident)
				if !ok || typ.Name != "EventType" {
					continue
				}
				for _, v := range vs.Values {
					lit, ok := v.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						t.Fatalf("%s: EventType constant is not a string literal", fset.Position(v.Pos()))
					}
					s, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatalf("unquote %s: %v", lit.Value, err)
					}
					deklariert = append(deklariert, s)
				}
			}
		}
	}
	gelistet := make([]string, 0, len(allEventTypes))
	for _, et := range allEventTypes {
		gelistet = append(gelistet, string(et))
	}
	slices.Sort(deklariert)
	slices.Sort(gelistet)
	if !slices.Equal(deklariert, gelistet) {
		t.Errorf("EventType constants in source != allEventTypes — add a frozen contract test\n source: %v\n listed: %v",
			deklariert, gelistet)
	}
}

// The EventType constants must match the SQL literals in the kj_extract_* functions
// (01_initial.up.sql, 11_pfandrueckgabe.up.sql, 13_pfand_kennzeichen.up.sql); a mismatch makes the SQL silently return NULL.
func TestEventContract_SQLLiterals(t *testing.T) {
	checks := []struct {
		constant EventType
		sqlLit   string
	}{
		{EventTypeBestellungAufgenommenV1, "bestellung-aufgenommen:v1"},
		{EventTypeZahlungKassiertV1, "zahlung-kassiert:v1"},
		{EventTypeZahlungKassiertV2, "zahlung-kassiert:v2"},
		{EventTypeStornierungErteiltV1, "stornierung-erteilt:v1"},
		{EventTypeBestellungKorrigiertV1, "bestellung-korrigiert:v1"},
		{EventTypeBestellungUmgebuchtV1, "bestellung-umgebucht:v1"},
		{EventTypeKassensitzungEroeffnetV1, "kassensitzung-eroeffnet:v1"},
		{EventTypeGeldtransitGebuchtV1, "geldtransit-gebucht:v1"},
		{EventTypeKassensturzDurchgefuehrtV1, "kassensturz-durchgefuehrt:v1"},
		{EventTypeDifferenzSollIstGebuchtV1, "differenz-soll-ist-gebucht:v1"},
		{EventTypeTagesabschlussErstelltV1, "tagesabschluss-erstellt:v1"},
		{EventTypeDirektverkaufGetaetigtV1, "direktverkauf-getaetigt:v1"},
		{EventTypeDirektverkaufGetaetigtV2, "direktverkauf-getaetigt:v2"},
		{EventTypeDirektverkaufStorniertV1, "direktverkauf-storniert:v1"},
		{EventTypePfandZurueckgegebenV1, "pfand-zurueckgegeben:v1"},
		{EventTypePfandrueckgabeStorniertV1, "pfandrueckgabe-storniert:v1"},
	}
	for _, c := range checks {
		if string(c.constant) != c.sqlLit {
			t.Errorf("EventType constant %q != SQL literal %q", c.constant, c.sqlLit)
		}
	}
}

// Pins the "einlage"/"entnahme" literals used by kj_extract_geldtransit_cents in
// 01_initial.up.sql — both the JSON key and the exact value, so value drift against the SQL
// literals shows up.
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

// assertJSONKeyPresent guards against a silent tag rename on fields pinned with a zero value
// (e.g. kommentar: "") — a missing json tag would otherwise be invisible.
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

// assertJSONKeyAbsent secures the byte identity of the additive omitempty field: an event
// without benutzerKommentar must not carry the key.
func assertJSONKeyAbsent(t *testing.T, v any, jsonKey string) {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, ok := m[jsonKey]; ok {
		t.Errorf("JSON key %q unexpectedly present in %T output — omitempty missing?", jsonKey, v)
	}
}

// positionLiteral is the frozen position JSON. Changing a key here means changing the
// corresponding SQL and frontend schema too.
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
	assertField(t, "kategorie", p.Kategorie, "getraenk")
}

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
	assertPosition(t, data.Positionen[0].PositionEventData)
	assertJSONKeyPresent(t, data, "kommentar")
}

// D22: the Positionskommentar is an additive optional key; an event without it keeps the frozen bytes.
func TestEventContract_BestellungAufgenommenV1_Positionskommentar(t *testing.T) {
	const altLit = `{
		"bestellungId":    "11111111-1111-4111-8111-111111111111",
		"positionen":      [` + positionLiteral + `],
		"gesamtPreisCents": 700,
		"kommentar":       ""
	}`
	var alt BestellungAufgenommenV1Data
	unmarshalJSON(t, altLit, &alt)
	assertField(t, "positionen[0].kommentar", alt.Positionen[0].Kommentar, "")
	if err := bestellungAufgenommenV1DataSchema.Validate(&alt); err != nil {
		t.Errorf("alt event without position kommentar rejected by schema: %v", err)
	}

	const altKompakt = `{"bestellungId":"11111111-1111-4111-8111-111111111111","positionen":[{"positionId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","varianteId":7,"produktName":"Cola","varianteName":"0,5l","kategorie":"getraenk","steuersatz":"regel","einzelpreisCents":350,"menge":2}],"gesamtPreisCents":700,"kommentar":""}`
	raw, err := json.Marshal(alt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != altKompakt {
		t.Errorf("alt event not byte-identical:\n got %s\nwant %s", raw, altKompakt)
	}

	const mitLit = `{
		"bestellungId":    "11111111-1111-4111-8111-111111111111",
		"positionen":      [{
			"positionId":       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"varianteId":       7,
			"produktName":      "Cola",
			"varianteName":     "0,5l",
			"kategorie":        "getraenk",
			"steuersatz":       "regel",
			"einzelpreisCents": 350,
			"menge":            2,
			"kommentar":        "ohne Kümmel"
		}],
		"gesamtPreisCents": 700,
		"kommentar":       ""
	}`
	var mit BestellungAufgenommenV1Data
	unmarshalJSON(t, mitLit, &mit)
	assertPosition(t, mit.Positionen[0].PositionEventData)
	assertField(t, "positionen[0].kommentar", mit.Positionen[0].Kommentar, "ohne Kümmel")
	assertJSONKeyPresent(t, mit.Positionen[0], "kommentar")
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

// positionV2Literal is the frozen v2 sale position: the v1 position plus the Pfand flag.
const positionV2Literal = `{
	"positionId":       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	"varianteId":       7,
	"produktName":      "Cola",
	"varianteName":     "0,5l",
	"kategorie":        "getraenk",
	"steuersatz":       "regel",
	"einzelpreisCents": 350,
	"menge":            2,
	"pfand":            true
}`

func assertPositionV2(t *testing.T, positionen []PositionV2EventData) {
	t.Helper()
	if len(positionen) != 1 {
		t.Fatalf("positionen count: %d", len(positionen))
	}
	assertPosition(t, positionen[0].PositionEventData)
	assertField(t, "pfand", positionen[0].Pfand, true)
	assertJSONKeyPresent(t, PositionV2EventData{}, "pfand")
	assertJSONKeyPresent(t, PositionV2EventData{}, "positionId")
}

func TestEventContract_ZahlungKassiertV2(t *testing.T) {
	const lit = `{
		"zahlungId":          "22222222-2222-4222-8222-222222222222",
		"positionen":         [` + positionV2Literal + `],
		"gesamtZahlungCents": 700,
		"kommentar":          ""
	}`
	var data ZahlungKassiertV2Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "zahlungId", data.ZahlungID, "22222222-2222-4222-8222-222222222222")
	assertField(t, "gesamtZahlungCents", data.GesamtZahlungCents, 700)
	assertPositionV2(t, data.Positionen)
	assertJSONKeyPresent(t, data, "kommentar")

	// Readers that ignore the flag read v2 as v1.
	var v1 ZahlungKassiertV1Data
	unmarshalJSON(t, lit, &v1)
	if len(v1.Positionen) != 1 {
		t.Fatalf("v1 positionen count: %d", len(v1.Positionen))
	}
	assertPosition(t, v1.Positionen[0])
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
	// Alt-Event ohne benutzerKommentar: parst mit Leerstring.
	const lit = `{
		"umbuchungId":  "55555555-5555-4555-8555-555555555555",
		"quellTischId": 3,
		"zielTischId":  7,
		"positionen":   [` + positionLiteral + `],
		"gesamtCents":  700,
		"kommentar":    "Umbuchung von Tisch 3"
	}`
	var data BestellungUmgebuchtV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "umbuchungId", data.UmbuchungID, "55555555-5555-4555-8555-555555555555")
	assertField(t, "quellTischId", data.QuellTischID, 3)
	assertField(t, "zielTischId", data.ZielTischID, 7)
	assertField(t, "gesamtCents", data.GesamtCents, 700)
	assertField(t, "kommentar", data.Kommentar, "Umbuchung von Tisch 3")
	assertField(t, "benutzerKommentar", data.BenutzerKommentar, "")
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "kommentar")

	// Ein Alt-Event ohne benutzerKommentar muss weiterhin durch das Schema validieren;
	// bricht, falls das Feld je auf Required umgestellt wird (zog-Zero-Value-Falle).
	if err := bestellungUmgebuchtV1DataSchema.Validate(&data); err != nil {
		t.Errorf("alt event without benutzerKommentar rejected by schema: %v", err)
	}

	// Ohne benutzerKommentar bleibt die Serialisierung byte-identisch (omitempty).
	altData := BestellungUmgebuchtV1Data{
		UmbuchungID:  "55555555-5555-4555-8555-555555555555",
		QuellTischID: 3,
		ZielTischID:  7,
		Positionen:   data.Positionen,
		GesamtCents:  700,
		Kommentar:    "Umbuchung von Tisch 3",
	}
	assertJSONKeyAbsent(t, altData, "benutzerKommentar")

	const litMitBenutzer = `{
		"umbuchungId":       "55555555-5555-4555-8555-555555555555",
		"quellTischId":      3,
		"zielTischId":       7,
		"positionen":        [` + positionLiteral + `],
		"gesamtCents":       700,
		"kommentar":         "Umbuchung von Tisch 3",
		"benutzerKommentar": "Gast an anderen Tisch gewechselt"
	}`
	var dataMitBenutzer BestellungUmgebuchtV1Data
	unmarshalJSON(t, litMitBenutzer, &dataMitBenutzer)
	assertField(t, "benutzerKommentar", dataMitBenutzer.BenutzerKommentar, "Gast an anderen Tisch gewechselt")
	assertJSONKeyPresent(t, dataMitBenutzer, "benutzerKommentar")
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
	assertField(t, "kommentar", data.Kommentar, "Wechselgeld nachgefüllt")
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
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_DirektverkaufGetaetigtV2(t *testing.T) {
	const lit = `{
		"verkaufId":         "88888888-8888-4888-8888-888888888888",
		"positionen":        [` + positionV2Literal + `],
		"gesamtbetragCents": 700,
		"kommentar":         ""
	}`
	var data DirektverkaufGetaetigtV2Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "verkaufId", data.VerkaufID, "88888888-8888-4888-8888-888888888888")
	assertField(t, "gesamtbetragCents", data.GesamtbetragCents, 700)
	assertPositionV2(t, PositionenV2OhneKommentar(data.Positionen))
	assertJSONKeyPresent(t, data, "kommentar")

	// Readers that ignore the flag read v2 as v1.
	var v1 DirektverkaufGetaetigtV1Data
	unmarshalJSON(t, lit, &v1)
	if len(v1.Positionen) != 1 {
		t.Fatalf("v1 positionen count: %d", len(v1.Positionen))
	}
	assertPosition(t, v1.Positionen[0])
}

// D22: the Positionskommentar is an additive optional key; an event without it keeps the frozen bytes.
func TestEventContract_DirektverkaufGetaetigtV2_Positionskommentar(t *testing.T) {
	const altLit = `{
		"verkaufId":         "88888888-8888-4888-8888-888888888888",
		"positionen":        [` + positionV2Literal + `],
		"gesamtbetragCents": 700,
		"kommentar":         ""
	}`
	var alt DirektverkaufGetaetigtV2Data
	unmarshalJSON(t, altLit, &alt)
	assertField(t, "positionen[0].kommentar", alt.Positionen[0].Kommentar, "")
	if err := direktverkaufGetaetigtV2DataSchema.Validate(&alt); err != nil {
		t.Errorf("alt event without position kommentar rejected by schema: %v", err)
	}

	const altKompakt = `{"verkaufId":"88888888-8888-4888-8888-888888888888","positionen":[{"positionId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","varianteId":7,"produktName":"Cola","varianteName":"0,5l","kategorie":"getraenk","steuersatz":"regel","einzelpreisCents":350,"menge":2,"pfand":true}],"gesamtbetragCents":700,"kommentar":""}`
	raw, err := json.Marshal(alt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != altKompakt {
		t.Errorf("alt event not byte-identical:\n got %s\nwant %s", raw, altKompakt)
	}

	const mitLit = `{
		"verkaufId":         "88888888-8888-4888-8888-888888888888",
		"positionen":        [{
			"positionId":       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"varianteId":       7,
			"produktName":      "Cola",
			"varianteName":     "0,5l",
			"kategorie":        "getraenk",
			"steuersatz":       "regel",
			"einzelpreisCents": 350,
			"menge":            2,
			"pfand":            true,
			"kommentar":        "ohne Senf"
		}],
		"gesamtbetragCents": 700,
		"kommentar":         ""
	}`
	var mit DirektverkaufGetaetigtV2Data
	unmarshalJSON(t, mitLit, &mit)
	assertPositionV2(t, PositionenV2OhneKommentar(mit.Positionen))
	assertField(t, "positionen[0].kommentar", mit.Positionen[0].Kommentar, "ohne Senf")
	assertJSONKeyPresent(t, mit.Positionen[0], "kommentar")

	// Readers of DirektverkaufGetaetigtV1Data ignore the key.
	var v1 DirektverkaufGetaetigtV1Data
	unmarshalJSON(t, mitLit, &v1)
	if len(v1.Positionen) != 1 {
		t.Fatalf("v1 positionen count: %d", len(v1.Positionen))
	}
	assertPosition(t, v1.Positionen[0])
	if err := direktverkaufGetaetigtV1DataSchema.Validate(&v1); err != nil {
		t.Errorf("v1 reader rejects event with position kommentar: %v", err)
	}
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

func TestEventContract_PfandZurueckgegebenV1(t *testing.T) {
	const lit = `{
		"rueckgabeId":       "aaaaaaaa-0000-4000-8000-000000000001",
		"tischId":           0,
		"positionen":        [` + positionLiteral + `],
		"gesamtbetragCents": 700,
		"kommentar":         ""
	}`
	var data PfandZurueckgegebenV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "rueckgabeId", data.RueckgabeID, "aaaaaaaa-0000-4000-8000-000000000001")
	assertField(t, "tischId", data.TischID, 0)
	assertField(t, "gesamtbetragCents", data.GesamtbetragCents, 700)
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
	assertJSONKeyPresent(t, data, "tischId")
	assertJSONKeyPresent(t, data, "kommentar")
}

func TestEventContract_PfandrueckgabeStorniertV1(t *testing.T) {
	const lit = `{
		"stornierungId":          "bbbbbbbb-0000-4000-8000-000000000002",
		"rueckgabeId":            "aaaaaaaa-0000-4000-8000-000000000001",
		"tischId":                4,
		"positionen":             [` + positionLiteral + `],
		"gesamtStornierungCents": 700,
		"kommentar":              "Falsch gezählt"
	}`
	var data PfandrueckgabeStorniertV1Data
	unmarshalJSON(t, lit, &data)
	assertField(t, "stornierungId", data.StornierungID, "bbbbbbbb-0000-4000-8000-000000000002")
	assertField(t, "rueckgabeId", data.RueckgabeID, "aaaaaaaa-0000-4000-8000-000000000001")
	assertField(t, "tischId", data.TischID, 4)
	assertField(t, "gesamtStornierungCents", data.GesamtStornierungCents, 700)
	assertField(t, "kommentar", data.Kommentar, "Falsch gezählt")
	if len(data.Positionen) != 1 {
		t.Fatalf("positionen count: %d", len(data.Positionen))
	}
	assertPosition(t, data.Positionen[0])
}

// positionKompakt is positionLiteral as json.Marshal writes it.
const positionKompakt = `{"positionId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","varianteId":7,"produktName":"Cola","varianteName":"0,5l","kategorie":"getraenk","steuersatz":"regel","einzelpreisCents":350,"menge":2`

// Pins the stored bytes of every event type: a stored event decodes and re-encodes to exactly
// the same JSON, so key order, key names, omitempty and the time format cannot drift.
func TestEventContract_ByteIdentitaet(t *testing.T) {
	cases := []struct {
		typ  EventType
		data func() any
		lit  string
	}{
		{EventTypeBestellungAufgenommenV1, func() any { return &BestellungAufgenommenV1Data{} },
			`{"bestellungId":"11111111-1111-4111-8111-111111111111","positionen":[` + positionKompakt + `,"kommentar":"ohne Kümmel"}],"gesamtPreisCents":700,"kommentar":"Tisch am Fenster"}`},
		{EventTypeZahlungKassiertV1, func() any { return &ZahlungKassiertV1Data{} },
			`{"zahlungId":"22222222-2222-4222-8222-222222222222","positionen":[` + positionKompakt + `}],"gesamtZahlungCents":700,"kommentar":""}`},
		{EventTypeZahlungKassiertV2, func() any { return &ZahlungKassiertV2Data{} },
			`{"zahlungId":"22222222-2222-4222-8222-222222222222","positionen":[` + positionKompakt + `,"pfand":true}],"gesamtZahlungCents":700,"kommentar":""}`},
		{EventTypeStornierungErteiltV1, func() any { return &StornierungErteiltV1Data{} },
			`{"stornierungId":"33333333-3333-4333-8333-333333333333","zahlungId":"22222222-2222-4222-8222-222222222222","positionen":[` + positionKompakt + `}],"gesamtStornierungCents":700,"kommentar":"Gast hat reklamiert"}`},
		{EventTypeBestellungKorrigiertV1, func() any { return &BestellungKorrigiertV1Data{} },
			`{"korrekturId":"44444444-4444-4444-8444-444444444444","positionen":[` + positionKompakt + `}],"gesamtCents":700,"kommentar":""}`},
		{EventTypeBestellungUmgebuchtV1, func() any { return &BestellungUmgebuchtV1Data{} },
			`{"umbuchungId":"55555555-5555-4555-8555-555555555555","quellTischId":3,"zielTischId":7,"positionen":[` + positionKompakt + `}],"gesamtCents":700,"kommentar":"Umbuchung von Tisch 3","benutzerKommentar":"Gast an anderen Tisch gewechselt"}`},
		{EventTypeKassensitzungEroeffnetV1, func() any { return &KassensitzungEroeffnetV1Data{} },
			`{"datum":"2026-07-07","bezeichnung":"Maihock 2026","betragCents":5000,"eroeffnetVon":1}`},
		{EventTypeGeldtransitGebuchtV1, func() any { return &GeldtransitGebuchtV1Data{} },
			`{"geldtransitId":"77777777-7777-4777-8777-777777777777","richtung":"einlage","betragCents":10000,"kommentar":"Wechselgeld nachgefüllt","gebuchtVon":2}`},
		{EventTypeKassensturzDurchgefuehrtV1, func() any { return &KassensturzDurchgefuehrtV1Data{} },
			`{"sollBestandCents":15000,"istBestandCents":14500,"differenzCents":500,"durchgefuehrtVon":1}`},
		{EventTypeDifferenzSollIstGebuchtV1, func() any { return &DifferenzSollIstGebuchtV1Data{} },
			`{"betragCents":-500,"gebuchtVon":1}`},
		{EventTypeTagesabschlussErstelltV1, func() any { return &TagesabschlussErstelltV1Data{} },
			`{"zNr":3,"zeitraumVon":"2026-07-07T08:00:00Z","zeitraumBis":"2026-07-07T22:00:00Z","umsatzGesamtCents":250000,"stornierungCents":500,"geldtransitCents":10000,"erstelltVon":1}`},
		{EventTypeDirektverkaufGetaetigtV1, func() any { return &DirektverkaufGetaetigtV1Data{} },
			`{"verkaufId":"88888888-8888-4888-8888-888888888888","positionen":[` + positionKompakt + `}],"gesamtbetragCents":700,"kommentar":""}`},
		{EventTypeDirektverkaufGetaetigtV2, func() any { return &DirektverkaufGetaetigtV2Data{} },
			`{"verkaufId":"88888888-8888-4888-8888-888888888888","positionen":[` + positionKompakt + `,"pfand":true,"kommentar":"ohne Senf"}],"gesamtbetragCents":700,"kommentar":""}`},
		{EventTypeDirektverkaufStorniertV1, func() any { return &DirektverkaufStorniertV1Data{} },
			`{"stornierungId":"99999999-9999-4999-8999-999999999999","verkaufId":"88888888-8888-4888-8888-888888888888","positionen":[` + positionKompakt + `}],"gesamtStornierungCents":700,"kommentar":"Bestellung falsch"}`},
		{EventTypePfandZurueckgegebenV1, func() any { return &PfandZurueckgegebenV1Data{} },
			`{"rueckgabeId":"aaaaaaaa-0000-4000-8000-000000000001","tischId":0,"positionen":[` + positionKompakt + `}],"gesamtbetragCents":700,"kommentar":""}`},
		{EventTypePfandrueckgabeStorniertV1, func() any { return &PfandrueckgabeStorniertV1Data{} },
			`{"stornierungId":"bbbbbbbb-0000-4000-8000-000000000002","rueckgabeId":"aaaaaaaa-0000-4000-8000-000000000001","tischId":4,"positionen":[` + positionKompakt + `}],"gesamtStornierungCents":700,"kommentar":"Falsch gezählt"}`},
	}
	gesehen := map[EventType]bool{}
	for _, c := range cases {
		gesehen[c.typ] = true
		t.Run(string(c.typ), func(t *testing.T) {
			data := c.data()
			unmarshalJSON(t, c.lit, data)
			raw, err := json.Marshal(data)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(raw) != c.lit {
				t.Errorf("not byte-identical:\n got %s\nwant %s", raw, c.lit)
			}
		})
	}
	for _, et := range allEventTypes {
		if !gesehen[et] {
			t.Errorf("event type %q has no byte-identity row", et)
		}
	}
}
