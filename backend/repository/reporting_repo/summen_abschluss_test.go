//go:build integration

package reporting_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

// TestComputeAbschlussSummen_AequivalenzMitSQLReporting guards that the Go closing sums (ComputeAbschlussSummen)
// and the SQL aggregation (GetReportingStats via kj_extract_*) never diverge.
func TestComputeAbschlussSummen_AequivalenzMitSQLReporting(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	ctx := context.Background()
	userID := createUser(t, db, "Test User", "testuser", "active")
	ksNr := createKassensitzung(t, db)

	tischSubject := kasse.TischSessionSubject(ksNr, 1)
	ksSubject := kasse.KassensitzungSubject(ksNr)
	dvSubject := kasse.DirektverkaufSubject(ksNr, "d1")
	pfandSubject := kasse.PfandrueckgabeSubject(ksNr, "r1")

	// zahlung-kassiert: kj_extract_zahlung_cents → gesamtZahlungCents
	insertEvent(t, db, userID, "testuser", "zahlung-kassiert:v1", tischSubject, 1, map[string]any{
		"zahlungId":          "z1",
		"gesamtZahlungCents": 2238,
		"positionen": []map[string]any{{
			"positionId":       "p1",
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"steuersatz":       "regel",
			"einzelpreisCents": 2238,
			"menge":            1,
		}},
		"kommentar": "",
	}, ksNr)

	// stornierung-erteilt: kj_extract_stornierung_cents → gesamtStornierungCents
	insertEvent(t, db, userID, "testuser", "stornierung-erteilt:v1", tischSubject, 2, map[string]any{
		"stornierungId":          "s1",
		"zahlungId":              "z1",
		"gesamtStornierungCents": 1455,
		"positionen": []map[string]any{{
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"steuersatz":       "regel",
			"menge":            1,
			"einzelpreisCents": 1455,
		}},
		"kommentar": "Warenruecknahme",
	}, ksNr)

	// bestellung-korrigiert: kj_extract_korrektur_cents → gesamtCents
	insertEvent(t, db, userID, "testuser", "bestellung-korrigiert:v1", tischSubject, 3, map[string]any{
		"korrekturId": "k1",
		"gesamtCents": 200,
		"kommentar":   "Korrektur",
		"positionen": []map[string]any{{
			"produktName":      "Limo",
			"varianteName":     "0.3L",
			"menge":            1,
			"einzelpreisCents": 200,
		}},
	}, ksNr)

	// direktverkauf-getaetigt: kj_extract_direktverkauf_cents → gesamtbetragCents
	insertEvent(t, db, userID, "testuser", "direktverkauf-getaetigt:v1", dvSubject, 1, map[string]any{
		"verkaufId":         "d1",
		"gesamtbetragCents": 880,
		"positionen": []map[string]any{{
			"produktName":      "T-Shirt",
			"steuersatz":       "regel",
			"einzelpreisCents": 880,
			"menge":            1,
		}},
	}, ksNr)

	// direktverkauf-storniert: kj_extract_direktverkauf_storno_cents → gesamtStornierungCents
	insertEvent(t, db, userID, "testuser", "direktverkauf-storniert:v1", dvSubject, 2, map[string]any{
		"stornierungId":          "ds1",
		"verkaufId":              "d1",
		"gesamtStornierungCents": 335,
		"positionen": []map[string]any{{
			"produktName":      "T-Shirt",
			"steuersatz":       "regel",
			"einzelpreisCents": 335,
			"menge":            1,
		}},
		"kommentar": "Fehlbuchung",
	}, ksNr)

	// v2 sale events carry the Pfand flag per position and count like v1.
	insertEvent(t, db, userID, "testuser", string(kasse.EventTypeZahlungKassiertV2), tischSubject, 4, map[string]any{
		"zahlungId":          "z2",
		"gesamtZahlungCents": 300,
		"positionen": []map[string]any{{
			"positionId":       "p2",
			"produktName":      "Pfand",
			"varianteName":     "Glas",
			"steuersatz":       "regel",
			"einzelpreisCents": 300,
			"menge":            1,
			"pfand":            true,
		}},
		"kommentar": "",
	}, ksNr)
	insertEvent(t, db, userID, "testuser", string(kasse.EventTypeDirektverkaufGetaetigtV2), kasse.DirektverkaufSubject(ksNr, "d2"), 1, map[string]any{
		"verkaufId":         "d2",
		"gesamtbetragCents": 120,
		"positionen": []map[string]any{{
			"produktName":      "Pfand",
			"steuersatz":       "regel",
			"einzelpreisCents": 120,
			"menge":            1,
			"pfand":            true,
		}},
	}, ksNr)

	// pfand-zurueckgegeben: kj_extract_pfandrueckgabe_cents → gesamtbetragCents
	insertEvent(t, db, userID, "testuser", string(kasse.EventTypePfandZurueckgegebenV1), pfandSubject, 1, map[string]any{
		"rueckgabeId":       "r1",
		"tischId":           0,
		"gesamtbetragCents": 400,
		"positionen": []map[string]any{{
			"produktName":      "Pfand",
			"varianteName":     "Glas",
			"steuersatz":       "regel",
			"einzelpreisCents": 200,
			"menge":            2,
		}},
		"kommentar": "",
	}, ksNr)

	// pfandrueckgabe-storniert: kj_extract_pfandrueckgabe_storno_cents → gesamtStornierungCents
	insertEvent(t, db, userID, "testuser", string(kasse.EventTypePfandrueckgabeStorniertV1), pfandSubject, 2, map[string]any{
		"stornierungId":          "rs1",
		"rueckgabeId":            "r1",
		"tischId":                0,
		"gesamtStornierungCents": 150,
		"positionen": []map[string]any{{
			"produktName":      "Pfand",
			"varianteName":     "Becher",
			"steuersatz":       "regel",
			"einzelpreisCents": 150,
			"menge":            1,
		}},
		"kommentar": "Falsch gezählt",
	}, ksNr)

	// geldtransit-gebucht deposit: kj_extract_geldtransit_cents → richtung + betragCents
	insertEvent(t, db, userID, "testuser", "geldtransit-gebucht:v1", ksSubject, 1, map[string]any{
		"betragCents":  500,
		"richtung":     "einlage",
		"beschreibung": "Wechselgeld",
	}, ksNr)

	// geldtransit-gebucht withdrawal
	insertEvent(t, db, userID, "testuser", "geldtransit-gebucht:v1", ksSubject, 2, map[string]any{
		"betragCents":  150,
		"richtung":     "entnahme",
		"beschreibung": "Restgeld entnommen",
	}, ksNr)

	// Umsatz 2238 − 1455 + 880 − 335 + 300 + 120 − 400 + 150 = 1498, Storno 1455 + 200 + 335 = 1990,
	// Geldtransit 500 − 150 = 350. Umsatz is also pinned to the literal, so a Pfand type
	// missing on both sides cannot pass.

	// Go side: read the events from the DB and aggregate.
	kjRepo := kassenjournal_repo.NewRepository(db)
	events, err := kjRepo.ReadKassensitzungEvents(ctx, ksNr)
	if err != nil {
		t.Fatalf("ReadKassensitzungEvents: %v", err)
	}
	got, err := kasse.ComputeAbschlussSummen(events)
	if err != nil {
		t.Fatalf("ComputeAbschlussSummen: %v", err)
	}

	// SQL side: GetReportingStats via the kj_extract_* functions.
	reportingData, err := NewRepository(db).GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting: %v", err)
	}
	sql := reportingData.Summary

	if got.UmsatzCents != sql.GesamtUmsatzCents {
		t.Errorf("UmsatzCents: Go=%d, SQL=%d", got.UmsatzCents, sql.GesamtUmsatzCents)
	}
	if got.UmsatzCents != 1498 {
		t.Errorf("UmsatzCents: Go=%d, erwartet 1498", got.UmsatzCents)
	}
	if got.StornierungCents != sql.GesamtStornierungenCents {
		t.Errorf("StornierungCents: Go=%d, SQL=%d", got.StornierungCents, sql.GesamtStornierungenCents)
	}
	if got.GeldtransitCents != sql.GeldtransitCents {
		t.Errorf("GeldtransitCents: Go=%d, SQL=%d", got.GeldtransitCents, sql.GeldtransitCents)
	}
}
