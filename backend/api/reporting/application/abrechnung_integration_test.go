//go:build integration

package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
)

// Integrationstests der Abrechnung pro Servicekraft: echte Events in der DB,
// echte Storno-Zuordnung aus GetStornierungen, echte Aggregation der
// Anwendungsschicht. Geprüft wird ausschließlich, welche Beträge und welche
// Servicekraft in breakdowns.abrechnungProServicekraft landen.

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

// abrechnungSetup öffnet die Test-DB, räumt sie auf und liefert die Query mit
// dem echten Reporting-Repository sowie die Nummer einer offenen Kassensitzung.
func abrechnungSetup(t *testing.T) (*sql.DB, Query, int) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()
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

// position baut eine Fat-Event-Position mit den Feldern, die Reporting-Queries
// und Storno-Zuordnung auswerten.
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

// abrechnungByUser indiziert die Abrechnungszeilen über den eingefrorenen
// Username; fehlt eine Person, ist sie nicht in der Liste.
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

// Nimmt die Serviceleitung stellvertretend eine von der Servicekraft kassierte
// Zahlung zurück, mindert das die Abrechnung der Servicekraft. Die
// Serviceleitung bleibt unbelastet — sie hat weder kassiert noch einen eigenen
// Vorgang rückgängig gemacht und erscheint deshalb gar nicht.
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
		t.Fatalf("expected only anna in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 500, 1500, 1)
}

// Dieselbe Rücknahme, diesmal von der Servicekraft selbst erteilt: Die
// Zuordnung folgt dem Kassierer und ist keine Sonderregel für Vertretungsfälle.
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
		t.Fatalf("expected only anna in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 500, 1500, 1)
}

// Bezahlt Servicekraft A, was B bestellt hat, belastet die Rücknahme A: Der
// Bargeldfluss folgt der Zahlung, nicht der Bestellung. B hat weder kassiert
// noch einen zugeordneten Storno und erscheint deshalb nicht.
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

// Die geldneutrale Korrektur erzeugt nur einen Kontroll-Marker beim Besteller —
// kein Betrag, kein veränderter Abzugeben-Saldo. Der Besteller erscheint dafür
// mit eigener Zeile, auch ohne eigenes Kassieren.
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

// Direktverkauf und Direktverkauf-Storno laufen über eine eigene Kasse: Sie
// verändern keine Zeile der Abrechnung pro Servicekraft, obwohl sie in den
// Gesamtkennzahlen und in der Storno-Detailliste erscheinen.
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
		t.Fatalf("expected only anna (Tischservice) in the abrechnung, got %+v", data.Breakdowns.AbrechnungProServicekraft)
	}
	assertAbrechnung(t, byUser["anna"], 2000, 0, 2000, 0)

	// Gegenprobe: Der Direktverkauf-Storno ist trotzdem erfasst.
	if data.Summary.DirektverkaufUmsatzCents != 500 {
		t.Errorf("expected direktverkauf umsatz 500 (750 − 250), got %d", data.Summary.DirektverkaufUmsatzCents)
	}
	if len(data.Stornierungen) != 1 || data.Stornierungen[0].Quelle != reporting.QuelleDirektverkauf {
		t.Errorf("expected the DV-Storno in the detail list, got %+v", data.Stornierungen)
	}
}

// Wird eine Zahlung vollständig zurückgenommen, ist Abzugeben null — nie
// negativ. Die Rücknahme kann nur Positionen der referenzierten Zahlung
// zurücknehmen, und beide Seiten werden demselben Kassierer zugeordnet.
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
