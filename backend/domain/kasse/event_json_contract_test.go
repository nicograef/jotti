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
// break those queries silently (no compile error), so these tests pin the JSON
// keys the SQL depends on. If a key changes below, update the SQL too (and vice versa).

func eventJSONKeys(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("unmarshal %T to map: %v", v, err)
	}
	return keys
}

func assertJSONKeys(t *testing.T, v any, required ...string) {
	t.Helper()
	keys := eventJSONKeys(t, v)
	for _, key := range required {
		if _, ok := keys[key]; !ok {
			t.Errorf("%T: missing JSON key %q that the SQL reporting layer depends on", v, key)
		}
	}
}

func TestEventDataJSONContract_MonetaryKeys(t *testing.T) {
	assertJSONKeys(t, ZahlungKassiertV1Data{}, "gesamtZahlungCents", "positionen")
	assertJSONKeys(t, BestellungAufgenommenV1Data{}, "gesamtPreisCents", "positionen")
	assertJSONKeys(t, StornierungErteiltV1Data{}, "gesamtStornierungCents", "positionen", "kommentar")
	assertJSONKeys(t, BestellungUmgebuchtV1Data{}, "gesamtCents", "positionen", "quellTischId", "zielTischId", "umbuchungId")
	assertJSONKeys(t, DirektverkaufGetaetigtV1Data{}, "gesamtbetragCents", "positionen")
	assertJSONKeys(t, DirektverkaufStorniertV1Data{}, "gesamtStornierungCents", "positionen")
	assertJSONKeys(t, KassensitzungEroeffnetV1Data{}, "betragCents")
	assertJSONKeys(t, GeldtransitGebuchtV1Data{}, "richtung", "betragCents")
	assertJSONKeys(t, DifferenzSollIstGebuchtV1Data{}, "betragCents")
}

func TestEventDataJSONContract_PositionKeys(t *testing.T) {
	// kj_extract_umsatz_pro_steuersatz reads steuersatz/einzelpreis/menge per position;
	// reporting_repo additionally reads produktName/varianteName for the Stornierungs-Detail.
	assertJSONKeys(t, PositionEventData{},
		"steuersatz", "einzelpreis", "menge", "produktName", "varianteName")
}
