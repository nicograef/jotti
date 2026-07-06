//go:build integration

package reporting_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

// TestComputeAbschlussSummen_AequivalenzMitSQLReporting ist ein Integrationstest
// gegen die echte SQL-Schicht (kj_extract_*-Funktionen in reporting.sql).
// Er seedet eine Kassensitzung mit gemischten geldrelevanten Events, liest die
// Events via ReadKassensitzungEvents aus der Datenbank, ruft
// ComputeAbschlussSummen auf und vergleicht die drei Summen Feld für Feld
// mit dem Ergebnis von GetReportingStats. Damit wird sichergestellt, dass
// Go-Aggregation und SQL-Aggregation nicht auseinanderlaufen können.
func TestComputeAbschlussSummen_AequivalenzMitSQLReporting(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()
	cleanDB(t, db)
	defer cleanDB(t, db)

	ctx := context.Background()
	userID := createUser(t, db, "Test User", "testuser", "active")
	ksNr := createKassensitzung(t, db)

	tischSubject := kasse.TischSessionSubject(ksNr, 1)
	ksSubject := kasse.KassensitzungSubject(ksNr)
	dvSubject := kasse.DirektverkaufSubject(ksNr, "d1")

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

	// geldtransit-gebucht einlage: kj_extract_geldtransit_cents → richtung + betragCents
	insertEvent(t, db, userID, "testuser", "geldtransit-gebucht:v1", ksSubject, 1, map[string]any{
		"betragCents":  500,
		"richtung":     "einlage",
		"beschreibung": "Wechselgeld",
	}, ksNr)

	// geldtransit-gebucht entnahme
	insertEvent(t, db, userID, "testuser", "geldtransit-gebucht:v1", ksSubject, 2, map[string]any{
		"betragCents":  150,
		"richtung":     "entnahme",
		"beschreibung": "Restgeld entnommen",
	}, ksNr)

	// Erwartete Werte (zur Lesekontrolle, nicht für den Go-vs-SQL-Vergleich maßgeblich):
	// Umsatz:      2238 − 1455 (storno) + 880 (dv) − 335 (dvStorno) = 1328
	// Storno:      1455 (storno) + 200 (korrektur) + 335 (dvStorno) = 1990
	// Geldtransit: 500 (einlage) − 150 (entnahme) = 350

	// Go-Seite: Events aus der echten DB lesen und aggregieren.
	kjRepo := kassenjournal_repo.NewRepository(db)
	events, err := kjRepo.ReadKassensitzungEvents(ctx, ksNr)
	if err != nil {
		t.Fatalf("ReadKassensitzungEvents: %v", err)
	}
	got, err := kasse.ComputeAbschlussSummen(events)
	if err != nil {
		t.Fatalf("ComputeAbschlussSummen: %v", err)
	}

	// SQL-Seite: GetReportingStats via kj_extract_*-Funktionen.
	reportingData, err := NewRepository(db).GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting: %v", err)
	}
	sql := reportingData.Summary

	if got.UmsatzCents != sql.GesamtUmsatzCents {
		t.Errorf("UmsatzCents: Go=%d, SQL=%d", got.UmsatzCents, sql.GesamtUmsatzCents)
	}
	if got.StornierungCents != sql.GesamtStornierungenCents {
		t.Errorf("StornierungCents: Go=%d, SQL=%d", got.StornierungCents, sql.GesamtStornierungenCents)
	}
	if got.GeldtransitCents != sql.GeldtransitCents {
		t.Errorf("GeldtransitCents: Go=%d, SQL=%d", got.GeldtransitCents, sql.GeldtransitCents)
	}
}
