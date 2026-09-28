//go:build integration

package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
)

// Integration tests of the per-staff settlement with real events, storno attribution and aggregation.
// Rules under test: docs/handbuch.md §7.2.

func cleanAbrechnungDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanAbrechnungDB %q: %v", stmt, err)
		}
	}
}

func abrechnungSetup(t *testing.T) (*sql.DB, Query, int) {
	t.Helper()
	db := dbtest.Open()
	cleanAbrechnungDB(t, db)
	t.Cleanup(func() {
		cleanAbrechnungDB(t, db)
		_ = db.Close()
	})

	var zNr int
	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&zNr); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	return db, Query{ReportingRepo: reporting_repo.NewRepository(db)}, zNr
}

func createAbrechnungUser(t *testing.T, db *sql.DB, name, username string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, 'service', 'active', 'hash', 'hash', now(), now()) RETURNING id",
		name, username,
	).Scan(&id); err != nil {
		t.Fatalf("create user %q: %v", username, err)
	}
	return id
}

func insertAbrechnungEvent(t *testing.T, db *sql.DB, userID int, userName, eventType, subject string, version int, data map[string]any, zNr int) {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal %s: %v", eventType, err)
	}
	if _, err := db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		userID, userName, eventType, subject, version, raw, time.Now().UTC(), zNr,
	); err != nil {
		t.Fatalf("insert %s: %v", eventType, err)
	}
}

// position holds only the fat-event fields the report queries and storno attribution read.
func position(positionID string, einzelpreisCents int) map[string]any {
	return map[string]any{
		"positionId":       positionID,
		"varianteId":       10,
		"produktName":      "Bier",
		"varianteName":     "0,5 l",
		"kategorie":        "getraenk",
		"steuersatz":       "regel",
		"einzelpreisCents": einzelpreisCents,
		"menge":            1,
	}
}

func bestellungEvent(positionID string, betragCents int) map[string]any {
	return map[string]any{
		"bestellungId":     "b-" + positionID,
		"gesamtPreisCents": betragCents,
		"kommentar":        "",
		"positionen":       []map[string]any{position(positionID, betragCents)},
	}
}

func zahlungEvent(zahlungID, positionID string, betragCents int) map[string]any {
	return map[string]any{
		"zahlungId":          zahlungID,
		"gesamtZahlungCents": betragCents,
		"kommentar":          "",
		"positionen":         []map[string]any{position(positionID, betragCents)},
	}
}

func ruecknahmeEvent(zahlungID, positionID string, betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"stornierungId":          "s-" + zahlungID,
		"zahlungId":              zahlungID,
		"gesamtStornierungCents": betragCents,
		"kommentar":              kommentar,
		"positionen":             []map[string]any{position(positionID, betragCents)},
	}
}

func korrekturEvent(positionID string, betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"korrekturId": "k-" + positionID,
		"gesamtCents": betragCents,
		"kommentar":   kommentar,
		"positionen":  []map[string]any{position(positionID, betragCents)},
	}
}

func direktverkaufEvent(verkaufID string, betragCents int) map[string]any {
	return map[string]any{
		"verkaufId":         verkaufID,
		"gesamtbetragCents": betragCents,
		"positionen":        []map[string]any{position("dp-"+verkaufID, betragCents)},
	}
}

func direktverkaufStornoEvent(verkaufID string, betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"stornierungId":          "ds-" + verkaufID,
		"verkaufId":              verkaufID,
		"gesamtStornierungCents": betragCents,
		"kommentar":              kommentar,
		"positionen":             []map[string]any{position("dp-"+verkaufID, betragCents)},
	}
}

// abrechnungByUser keys the rows by the frozen username.
func abrechnungByUser(zeilen []reporting.AbrechnungServicekraft) map[string]reporting.AbrechnungServicekraft {
	out := map[string]reporting.AbrechnungServicekraft{}
	for _, z := range zeilen {
		out[z.UserName] = z
	}
	return out
}

func assertAbrechnung(t *testing.T, zeile reporting.AbrechnungServicekraft, kassiert, ruecknahmen, abzugeben, anzahlStornos int) {
	t.Helper()
	if zeile.KassiertCents != kassiert || zeile.RuecknahmenCents != ruecknahmen || zeile.AbzugebenCents != abzugeben || zeile.AnzahlStornierungen != anzahlStornos {
		t.Errorf("%s: expected kassiert %d / ruecknahmen %d / abzugeben %d / stornos %d, got %+v",
			zeile.UserName, kassiert, ruecknahmen, abzugeben, anzahlStornos, zeile)
	}
}

// A return issued by the Serviceleitung reduces the cashier's settlement.
// The Serviceleitung neither collected nor had its own transaction reversed, so it gets no row.
func TestAbrechnung_StellvertretendeRuecknahmeTrifftDenKassierer(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 500, "Ruecknahme"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	if len(byUser) != 1 {
		t.Errorf("expected only anna in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 500, 1500, 1)
}

// The same return issued by the cashier gives the same result: attribution follows the cashier, not a stand-in rule.
func TestAbrechnung_EigeneRuecknahmeErgibtDasselbe(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 500, "Ruecknahme"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	if len(byUser) != 1 {
		t.Errorf("expected only anna in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 500, 1500, 1)
}

// If A collects what B ordered, the return hits A: cash follows the payment, not the order.
// B neither collected nor has an attributed storno, so B gets no row.
func TestAbrechnung_RuecknahmeTrifftKassiererNichtBesteller(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	bobID := createAbrechnungUser(t, db, "Bob Schmidt", "bob")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, bobID, "bob", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 500, "Ruecknahme"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	assertAbrechnung(t, byUser["anna"], 2000, 500, 1500, 1)
	if _, ok := byUser["bob"]; ok {
		t.Errorf("expected bob (nur Besteller) to stay out of the abrechnung, got %+v", byUser["bob"])
	}
}

// A cash-neutral correction only counts for the orderer, without an amount or Abzugeben change.
// The orderer gets a row even without collections.
func TestAbrechnung_KorrekturZaehltNurAlsMarkerBeimBesteller(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	bobID := createAbrechnungUser(t, db, "Bob Schmidt", "bob")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, bobID, "bob", "bestellung-aufgenommen:v1", tisch, 3, bestellungEvent("pos-2", 300), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "bestellung-korrigiert:v1", tisch, 4, korrekturEvent("pos-2", 300, "Korrektur"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	assertAbrechnung(t, byUser["anna"], 2000, 0, 2000, 0)
	bob, ok := byUser["bob"]
	if !ok {
		t.Fatalf("expected bob to appear with his zugeordneter Storno, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, bob, 0, 0, 0, 1)
	if _, ok := byUser["lena"]; ok {
		t.Errorf("expected the stornierende lena to stay out of the abrechnung, got %+v", byUser["lena"])
	}
}

// Direct sales and their stornos use their own till: they change no settlement row but appear in the totals
// and the storno list.
func TestAbrechnung_DirektverkaufBleibtAussen(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"
	dv := "kassensitzung-1/direktverkauf-v-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "direktverkauf-getaetigt:v1", dv, 1, direktverkaufEvent("v-1", 750), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "direktverkauf-storniert:v1", dv, 2, direktverkaufStornoEvent("v-1", 250, "DV-Storno"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	if len(byUser) != 1 {
		t.Errorf("expected only anna (Tischservice) in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 0, 2000, 0)

	// Cross-check: the direct-sale storno is still listed.
	if data.Summary.DirektverkaufUmsatzCents != 500 {
		t.Errorf("expected direktverkauf umsatz 500 (750 − 250), got %d", data.Summary.DirektverkaufUmsatzCents)
	}
	if len(data.Stornierungen) != 1 || data.Stornierungen[0].Quelle != reporting.QuelleDirektverkauf {
		t.Errorf("expected the DV-Storno in the detail list, got %+v", data.Stornierungen)
	}
}

// A fully returned payment leaves Abzugeben at zero, never negative.
// See docs/handbuch.md §7.2.
func TestAbrechnung_VollstaendigeRuecknahmeErgibtNullNichtNegativ(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 2000, "Alles zurueck"), zNr)

	data, err := q.GetReporting(context.Background(), zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	byUser := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)
	assertAbrechnung(t, byUser["anna"], 2000, 2000, 0, 1)
}

func assertEigeneUebersicht(t *testing.T, u reporting.EigeneUebersicht, kassiert, ruecknahmen, anzahlRuecknahmen, abzugeben int) {
	t.Helper()
	if u.ZahlungenCents != kassiert || u.RuecknahmenCents != ruecknahmen ||
		u.AnzahlRuecknahmen != anzahlRuecknahmen || u.AbzugebenCents != abzugeben {
		t.Errorf("expected kassiert %d / ruecknahmen %d (%dx) / abzugeben %d, got %+v",
			kassiert, ruecknahmen, anzahlRuecknahmen, abzugeben, u)
	}
}

// Someone else's return on this cashier's payment lowers their Abzugeben but not the collected amount.
// Their own overview must show the same Abzugeben as their settlement row.
func TestEigeneUebersicht_FremdeRuecknahmeMindertAbzugeben(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 500, "Ruecknahme"), zNr)

	ctx := context.Background()
	uebersicht, err := q.ReportingRepo.GetEigeneUebersicht(ctx, annaID, zNr)
	if err != nil {
		t.Fatalf("GetEigeneUebersicht failed: %v", err)
	}
	assertEigeneUebersicht(t, uebersicht, 2000, 500, 1, 1500)

	data, err := q.GetReporting(ctx, zNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if got := abrechnungByUser(data.Breakdowns.AbrechnungProServicekraft)["anna"].AbzugebenCents; got != uebersicht.AbzugebenCents {
		t.Errorf("eigene Übersicht (%d) und Abrechnungszeile (%d) müssen denselben Abzugeben-Betrag nennen", uebersicht.AbzugebenCents, got)
	}
}

// A return on someone else's payment leaves the issuer's overview unchanged; it hits the cashier's till.
func TestEigeneUebersicht_EigeneRuecknahmeFremderZahlungZaehltNicht(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	bobID := createAbrechnungUser(t, db, "Bob Schmidt", "bob")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-anna", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, bobID, "bob", "bestellung-aufgenommen:v1", tisch, 3, bestellungEvent("pos-2", 1500), zNr)
	insertAbrechnungEvent(t, db, bobID, "bob", "zahlung-kassiert:v1", tisch, 4, zahlungEvent("z-bob", "pos-2", 1500), zNr)
	// anna returns against bob's payment.
	insertAbrechnungEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", tisch, 5, ruecknahmeEvent("z-bob", "pos-2", 400, "Ruecknahme Bob"), zNr)

	ctx := context.Background()
	anna, err := q.ReportingRepo.GetEigeneUebersicht(ctx, annaID, zNr)
	if err != nil {
		t.Fatalf("GetEigeneUebersicht(anna) failed: %v", err)
	}
	assertEigeneUebersicht(t, anna, 2000, 0, 0, 2000)

	bob, err := q.ReportingRepo.GetEigeneUebersicht(ctx, bobID, zNr)
	if err != nil {
		t.Fatalf("GetEigeneUebersicht(bob) failed: %v", err)
	}
	assertEigeneUebersicht(t, bob, 1500, 400, 1, 1100)
}

// A cash-neutral correction moves no cash, so it changes none of the overview's return fields.
func TestEigeneUebersicht_KorrekturVeraendertNichts(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 3, bestellungEvent("pos-2", 700), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "bestellung-korrigiert:v1", tisch, 4, korrekturEvent("pos-2", 700, "Korrektur"), zNr)

	uebersicht, err := q.ReportingRepo.GetEigeneUebersicht(context.Background(), annaID, zNr)
	if err != nil {
		t.Fatalf("GetEigeneUebersicht failed: %v", err)
	}
	assertEigeneUebersicht(t, uebersicht, 2000, 0, 0, 2000)
}

// A fully returned payment leaves the overview's Abzugeben at zero, never negative.
// See docs/handbuch.md §7.2.
func TestEigeneUebersicht_VollstaendigeRuecknahmeErgibtNull(t *testing.T) {
	db, q, zNr := abrechnungSetup(t)
	annaID := createAbrechnungUser(t, db, "Anna Müller", "anna")
	lenaID := createAbrechnungUser(t, db, "Lena Chef", "lena")
	tisch := "kassensitzung-1/tisch-1"

	insertAbrechnungEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", tisch, 1, bestellungEvent("pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", tisch, 2, zahlungEvent("z-1", "pos-1", 2000), zNr)
	insertAbrechnungEvent(t, db, lenaID, "lena", "stornierung-erteilt:v1", tisch, 3, ruecknahmeEvent("z-1", "pos-1", 2000, "Voll zurueck"), zNr)

	uebersicht, err := q.ReportingRepo.GetEigeneUebersicht(context.Background(), annaID, zNr)
	if err != nil {
		t.Fatalf("GetEigeneUebersicht failed: %v", err)
	}
	assertEigeneUebersicht(t, uebersicht, 2000, 2000, 1, 0)
}
