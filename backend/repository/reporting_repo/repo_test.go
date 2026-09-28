//go:build integration

package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	// tisch_sessions references kassenjournal (last_event_id) and goes first, including other packages' leftovers.
	if _, err := db.Exec("DELETE FROM tisch_sessions"); err != nil {
		t.Fatalf("Failed to clean tisch_sessions table: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to disable kassenjournal_no_delete trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassenjournal"); err != nil {
		t.Fatalf("Failed to clean kassenjournal table: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to enable kassenjournal_no_delete trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassensitzungen"); err != nil {
		t.Fatalf("Failed to clean kassensitzungen table: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}
}

func createUser(t *testing.T, db *sql.DB, name, username, status string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, 'service', $3, 'hash', 'hash', now(), now()) RETURNING id",
		name, username, status,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert user %q: %v", username, err)
	}
	return id
}

func createKassensitzung(t *testing.T, db *sql.DB) int {
	t.Helper()
	var zNr int
	err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&zNr)
	if err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}
	return zNr
}

func insertEvent(t *testing.T, db *sql.DB, userID int, userName, eventType, subject string, version int, data map[string]any, ksNr int) {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal event data: %v", err)
	}
	_, err = db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		userID, userName, eventType, subject, version, raw, time.Now().UTC(), ksNr,
	)
	if err != nil {
		t.Fatalf("Failed to insert %s event: %v", eventType, err)
	}
}

// zahlungData: a later return finds its cashier through zahlungId.
func zahlungData(zahlungID string, gesamtCents int) map[string]any {
	return map[string]any{
		"zahlungId": zahlungID,
		"positionen": []map[string]any{{
			"positionId":       "p0000000-0000-0000-0000-000000000001",
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"steuersatz":       "regel",
			"einzelpreisCents": gesamtCents,
			"menge":            1,
		}},
		"gesamtZahlungCents": gesamtCents,
		"kommentar":          "",
	}
}

// stornierungData builds a cash-relevant return referencing the returned payment.
func stornierungData(zahlungID string, betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"stornierungId":          "s0000000-0000-0000-0000-000000000001",
		"zahlungId":              zahlungID,
		"gesamtStornierungCents": betragCents,
		"kommentar":              kommentar,
		"positionen": []map[string]any{{
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"steuersatz":       "regel",
			"menge":            1,
			"einzelpreisCents": betragCents,
		}},
	}
}

// bestellungData builds one unit per position ID. The bestellungId derives from the first position ID, so several
// orders per session respect the UNIQUE index on bestellungId.
func bestellungData(positionIDs []string, einzelpreisCents int) map[string]any {
	positionen := make([]map[string]any, len(positionIDs))
	for i, id := range positionIDs {
		positionen[i] = map[string]any{
			"positionId":       id,
			"produktName":      "Limo",
			"varianteName":     "0.3L",
			"steuersatz":       "regel",
			"einzelpreisCents": einzelpreisCents,
			"menge":            1,
		}
	}
	return map[string]any{
		"bestellungId":     "b-" + positionIDs[0],
		"gesamtPreisCents": einzelpreisCents * len(positionIDs),
		"kommentar":        "",
		"positionen":       positionen,
	}
}

// korrekturData: the cash-neutral correction finds its orderers through the position IDs.
func korrekturData(positionIDs []string, betragCents int, kommentar string) map[string]any {
	data := bestellungData(positionIDs, betragCents/len(positionIDs))
	return map[string]any{
		"korrekturId": "k0000000-0000-0000-0000-000000000001",
		"gesamtCents": betragCents,
		"kommentar":   kommentar,
		"positionen":  data["positionen"],
	}
}

// direktverkaufData: a later storno finds its seller through verkaufId.
func direktverkaufData(verkaufID string, gesamtCents int) map[string]any {
	return map[string]any{
		"verkaufId":         verkaufID,
		"gesamtbetragCents": gesamtCents,
		"positionen": []map[string]any{{
			"positionId":       "dp000000-0000-0000-0000-000000000001",
			"produktName":      "Cola",
			"varianteName":     "0,5 l",
			"steuersatz":       "regel",
			"einzelpreisCents": gesamtCents,
			"menge":            1,
		}},
	}
}

func direktverkaufStornoData(verkaufID string, betragCents int, kommentar string) map[string]any {
	data := direktverkaufData(verkaufID, betragCents)
	return map[string]any{
		"stornierungId":          "ds000000-0000-0000-0000-000000000001",
		"verkaufId":              verkaufID,
		"gesamtStornierungCents": betragCents,
		"kommentar":              kommentar,
		"positionen":             data["positionen"],
	}
}

// assertBetroffene expects each frozen username exactly once, in any order.
func assertBetroffene(t *testing.T, storno reporting.StornierungDetail, wantUsernames ...string) {
	t.Helper()
	got := make([]string, len(storno.Betroffene))
	for i, b := range storno.Betroffene {
		got[i] = b.UserName
	}
	want := slices.Clone(wantUsernames)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Storno %q: expected betroffene %v, got %v", storno.Kommentar, want, got)
	}
}

// stornierungenByKommentar keys the storno rows by comment, which the tests use as a unique name.
func stornierungenByKommentar(stornierungen []reporting.StornierungDetail) map[string]reporting.StornierungDetail {
	out := map[string]reporting.StornierungDetail{}
	for _, s := range stornierungen {
		out[s.Kommentar] = s
	}
	return out
}

func produktPosition(varianteID int, produktName, varianteName, kategorie string, einzelpreisCents, menge int) map[string]any {
	return map[string]any{
		"positionId":       "p0000000-0000-0000-0000-000000000001",
		"varianteId":       varianteID,
		"produktName":      produktName,
		"varianteName":     varianteName,
		"kategorie":        kategorie,
		"steuersatz":       "regel",
		"einzelpreisCents": einzelpreisCents,
		"menge":            menge,
	}
}

// TestGetProduktStatistik_MengeUndUmsatzAufBestellbasis guards that payments, returns and direct-sale stornos leave
// quantity and revenue untouched: both are order-based. See the GetProduktStatistik query in sqlc/queries/reporting.sql.
func TestGetProduktStatistik_MengeUndUmsatzAufBestellbasis(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	pommes := func(menge int) map[string]any {
		return produktPosition(10, "Pommes", "groß", "essen", 300, menge)
	}
	cola := func(menge int) map[string]any {
		return produktPosition(20, "Cola", "0,5 l", "getraenk", 250, menge)
	}

	// Pommes groß: ordered 5, corrected 1 (cash-neutral), collected 4, 1 returned.
	insertEvent(t, db, userID, "anna", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 1, map[string]any{
		"bestellungId": "b1", "gesamtPreisCents": 1500, "positionen": []map[string]any{pommes(5)},
	}, ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 2, map[string]any{
		"korrekturId": "k1", "gesamtCents": 300, "kommentar": "eine zu viel", "positionen": []map[string]any{pommes(1)},
	}, ksNr)
	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 3, map[string]any{
		"zahlungId": "z1", "gesamtZahlungCents": 1200, "kommentar": "", "positionen": []map[string]any{pommes(4)},
	}, ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 4, map[string]any{
		"stornierungId": "s1", "zahlungId": "z1", "gesamtStornierungCents": 300, "kommentar": "Warenrücknahme", "positionen": []map[string]any{pommes(1)},
	}, ksNr)

	// Cola 0,5 l: direct sale of 3, 1 of them cancelled.
	insertEvent(t, db, userID, "anna", "direktverkauf-getaetigt:v1", "kassensitzung-1/direktverkauf-d1", 5, map[string]any{
		"verkaufId": "d1", "gesamtbetragCents": 750, "positionen": []map[string]any{cola(3)},
	}, ksNr)
	insertEvent(t, db, userID, "anna", "direktverkauf-storniert:v1", "kassensitzung-1/direktverkauf-d1", 6, map[string]any{
		"stornierungId": "ds1", "verkaufId": "d1", "gesamtStornierungCents": 250, "kommentar": "Fehlbuchung", "positionen": []map[string]any{cola(1)},
	}, ksNr)

	zeilen, err := repo.GetProduktStatistik(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetProduktStatistik failed: %v", err)
	}

	byVariante := map[int]reporting.ProduktStatistikZeile{}
	for _, z := range zeilen {
		byVariante[z.VarianteID] = z
	}
	if len(zeilen) != 2 {
		t.Errorf("expected 2 variant rows, got %d: %+v", len(zeilen), zeilen)
	}

	// Pommes: quantity 5 − 1 = 4, revenue 1500 − 300 = 1200; the later return does not reduce it.
	pommesZeile := byVariante[10]
	if pommesZeile.ProduktName != "Pommes" || pommesZeile.VarianteName != "groß" || pommesZeile.Kategorie != "essen" {
		t.Errorf("unexpected Pommes identity: %+v", pommesZeile)
	}
	if pommesZeile.AusgegebeneMenge != 4 {
		t.Errorf("expected Pommes menge 4 (5 bestellt − 1 korrigiert), got %d", pommesZeile.AusgegebeneMenge)
	}
	if pommesZeile.UmsatzCents != 1200 {
		t.Errorf("expected Pommes umsatz 1200 (1500 bestellt − 300 korrigiert, Warenrücknahme umsatzneutral), got %d", pommesZeile.UmsatzCents)
	}

	// Cola: the direct sale counts, its storno reduces neither quantity nor revenue.
	colaZeile := byVariante[20]
	if colaZeile.AusgegebeneMenge != 3 {
		t.Errorf("expected Cola menge 3 (Direktverkauf, Storno menge-neutral), got %d", colaZeile.AusgegebeneMenge)
	}
	if colaZeile.UmsatzCents != 750 {
		t.Errorf("expected Cola umsatz 750 (Direktverkauf, Storno umsatzneutral), got %d", colaZeile.UmsatzCents)
	}
}

// TestGetProduktStatistik_UmbuchungZaehltNicht: a transfer changes neither quantity nor revenue, since the order
// already counted the positions.
func TestGetProduktStatistik_UmbuchungZaehltNicht(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	pos := produktPosition(10, "Pommes", "groß", "essen", 300, 2)
	insertEvent(t, db, userID, "anna", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 1, map[string]any{
		"bestellungId": "b1", "gesamtPreisCents": 600, "positionen": []map[string]any{pos},
	}, ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-umgebucht:v1", "kassensitzung-1/tisch-3", 2, map[string]any{
		"umbuchungId": "u1", "quellTischId": 1, "zielTischId": 3, "gesamtCents": 600, "kommentar": "", "positionen": []map[string]any{pos},
	}, ksNr)

	zeilen, err := repo.GetProduktStatistik(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetProduktStatistik failed: %v", err)
	}
	if len(zeilen) != 1 {
		t.Fatalf("expected 1 variant row, got %d: %+v", len(zeilen), zeilen)
	}
	if zeilen[0].AusgegebeneMenge != 2 {
		t.Errorf("expected menge 2 (Umbuchung zählt nicht), got %d", zeilen[0].AusgegebeneMenge)
	}
	if zeilen[0].UmsatzCents != 600 {
		t.Errorf("expected umsatz 600 (2 × 300 bestellt, Umbuchung zählt nicht), got %d", zeilen[0].UmsatzCents)
	}
}

// TestGetReporting_ResolvesKlarnameIncludingSoftDeleted: the live LEFT JOIN resolves the current Klarname for active
// and soft-deleted users, while the frozen username stays the authoritative identity.
func TestGetReporting_ResolvesKlarnameIncludingSoftDeleted(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "deleted")
	ksNr := createKassensitzung(t, db)

	// Anna gets more revenue so she sorts first (ORDER BY kassiert_cents DESC).
	insertEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData("z-anna", 2000), ksNr)
	insertEvent(t, db, bobID, "bob", "zahlung-kassiert:v1", "kassensitzung-1/tisch-2", 1, zahlungData("z-bob", 1000), ksNr)

	insertEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData("z-anna", 500, "Anna storniert"), ksNr)
	insertEvent(t, db, bobID, "bob", "stornierung-erteilt:v1", "kassensitzung-1/tisch-2", 2, stornierungData("z-bob", 300, "Bob storniert"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	// Abrechnung pro Servicekraft: username frozen in the event, Klarname resolved live.
	klarnameByUsername := map[string]string{}
	for _, sk := range data.Breakdowns.AbrechnungProServicekraft {
		klarnameByUsername[sk.UserName] = sk.Name
	}
	if got := klarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna Klarname 'Anna Müller', got %q", got)
	}
	if got := klarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob Klarname 'Bob Schmidt', got %q", got)
	}

	// Stornierungen resolve the same way, for the actor and the affected staff member.
	stornoKlarnameByUsername := map[string]string{}
	betroffeneKlarnameByUsername := map[string]string{}
	for _, s := range data.Stornierungen {
		stornoKlarnameByUsername[s.Akteur.UserName] = s.Akteur.Name
		for _, b := range s.Betroffene {
			betroffeneKlarnameByUsername[b.UserName] = b.Name
		}
	}
	if got := stornoKlarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna storno Klarname 'Anna Müller', got %q", got)
	}
	if got := stornoKlarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob storno Klarname 'Bob Schmidt', got %q", got)
	}
	if got := betroffeneKlarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna betroffene Klarname 'Anna Müller', got %q", got)
	}
	if got := betroffeneKlarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob betroffene Klarname 'Bob Schmidt', got %q", got)
	}
}

// TestGetReporting_IncludesBeideStornoArten: the storno list and rate count the cash-relevant return
// (stornierung-erteilt, Bar-Rückgabe) and the cash-neutral correction (bestellung-korrigiert).
func TestGetReporting_IncludesBeideStornoArten(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData("z-anna", 2000), ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData("z-anna", 500, "Warenruecknahme"), ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData([]string{"pos-1"}, 300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	if len(data.Stornierungen) != 2 {
		t.Errorf("expected 2 Stornierungen (Warenrücknahme + Korrektur), got %d", len(data.Stornierungen))
	}

	byKommentar := stornierungenByKommentar(data.Stornierungen)

	warenruecknahme, ok := byKommentar["Warenruecknahme"]
	if !ok {
		t.Fatal("expected a Warenrücknahme entry")
	}
	if !warenruecknahme.BarRueckgabe {
		t.Error("expected Warenrücknahme to be marked as Bar-Rückgabe")
	}
	if warenruecknahme.BetragCents != 500 {
		t.Errorf("expected Warenrücknahme betrag 500, got %d", warenruecknahme.BetragCents)
	}

	korrektur, ok := byKommentar["Korrektur"]
	if !ok {
		t.Fatal("expected a Korrektur entry")
	}
	if korrektur.BarRueckgabe {
		t.Error("expected geldneutrale Korrektur to NOT be marked as Bar-Rückgabe")
	}
	if korrektur.BetragCents != 300 {
		t.Errorf("expected Korrektur betrag 300 (aus gesamtCents), got %d", korrektur.BetragCents)
	}

	// The storno sum counts both kinds (500 + 300) and both events.
	if data.Summary.GesamtStornierungenCents != 800 {
		t.Errorf("expected gesamt stornierungen 800 (beide Arten), got %d", data.Summary.GesamtStornierungenCents)
	}
	if data.Summary.AnzahlStornierungen != 2 {
		t.Errorf("expected anzahl stornierungen 2, got %d", data.Summary.AnzahlStornierungen)
	}
}

// Returns must enter the gross rows as negative rows, or the VAT breakdown exceeds revenue and diverges from the
// DSFinV-K export. Cash-neutral corrections carry no revenue and stay out.
func TestGetReporting_UmsatzProSteuersatzZiehtWarenruecknahmeAb(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData("z-anna", 2000), ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData("z-anna", 500, "Warenruecknahme"), ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData([]string{"pos-1"}, 300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	// Two regel rows: the payment positive, the return negative.
	if len(data.UmsatzProSteuersatz) != 2 {
		t.Fatalf("expected 2 Positionszeilen, got %d: %+v", len(data.UmsatzProSteuersatz), data.UmsatzProSteuersatz)
	}

	summe := 0
	regelSumme := 0
	for _, zeile := range data.UmsatzProSteuersatz {
		summe += zeile.BruttoCents
		if zeile.Satz == "regel" {
			regelSumme += zeile.BruttoCents
		}
	}

	if regelSumme != 1500 {
		t.Errorf("expected regel-Umsatz 1500 (2000 Zahlung - 500 Warenrücknahme), got %d", regelSumme)
	}
	if summe != data.Summary.GesamtUmsatzCents {
		t.Errorf("expected Zeilen-Summe %d == GesamtUmsatz %d", summe, data.Summary.GesamtUmsatzCents)
	}
}

// TestGetReporting_MetadatenAusJournalEvents: the report header projects opening and closing time, the closing user
// (frozen user_name) and the count difference purely from journal events.
func TestGetReporting_MetadatenAusJournalEvents(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	nicoID := createUser(t, db, "Nico Gräf", "nico", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, annaID, "anna", "kassensitzung-eroeffnet:v1", "kassensitzung-1", 1, map[string]any{
		"datum": "2026-07-05", "bezeichnung": "Sommerfest", "betragCents": 10000, "eroeffnetVon": annaID,
	}, ksNr)
	insertEvent(t, db, nicoID, "nico", "kassensturz-durchgefuehrt:v1", "kassensitzung-1", 2, map[string]any{
		"sollBestandCents": 20000, "istBestandCents": 19850, "differenzCents": -150, "durchgefuehrtVon": nicoID,
	}, ksNr)
	insertEvent(t, db, nicoID, "nico", "tagesabschluss-erstellt:v1", "kassensitzung-1", 3, map[string]any{
		"zNr": ksNr, "umsatzGesamtCents": 341200, "erstelltVon": nicoID,
	}, ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	m := data.Metadaten
	if m.EroeffnetAm == nil {
		t.Error("expected eroeffnetAm to be set from kassensitzung-eroeffnet:v1")
	}
	if m.AbgeschlossenAm == nil {
		t.Error("expected abgeschlossenAm to be set from tagesabschluss-erstellt:v1")
	}
	if m.AbgeschlossenVon != "nico" {
		t.Errorf("expected abgeschlossenVon 'nico', got %q", m.AbgeschlossenVon)
	}
	if m.KassensturzDifferenzCents == nil || *m.KassensturzDifferenzCents != -150 {
		t.Errorf("expected kassensturzDifferenzCents -150, got %v", m.KassensturzDifferenzCents)
	}
}

// TestGetReporting_MetadatenLeerOhneAbschlussEvents: an open session leaves the optional metadata empty instead of
// failing the NULL scan.
func TestGetReporting_MetadatenLeerOhneAbschlussEvents(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, annaID, "anna", "kassensitzung-eroeffnet:v1", "kassensitzung-1", 1, map[string]any{
		"datum": "2026-07-05", "bezeichnung": "Sommerfest", "betragCents": 10000, "eroeffnetVon": annaID,
	}, ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	m := data.Metadaten
	if m.EroeffnetAm == nil {
		t.Error("expected eroeffnetAm to be set even for the open session")
	}
	if m.AbgeschlossenAm != nil {
		t.Errorf("expected abgeschlossenAm to be nil without tagesabschluss, got %v", m.AbgeschlossenAm)
	}
	if m.AbgeschlossenVon != "" {
		t.Errorf("expected abgeschlossenVon empty without tagesabschluss, got %q", m.AbgeschlossenVon)
	}
	if m.KassensturzDifferenzCents != nil {
		t.Errorf("expected kassensturzDifferenzCents nil without kassensturz, got %v", m.KassensturzDifferenzCents)
	}
}

// Storno attribution: each row names the actor and the never-empty betroffene, resolved at read time from the event's
// back-reference. See docs/handbuch.md §7.2.

// TestGetStornierungen_RuecknahmeTrifftDenKassierer: a return issued by the Serviceleitung names the cashier as
// betroffen; the Serviceleitung stays the actor.
func TestGetStornierungen_RuecknahmeTrifftDenKassierer(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData("z-anna", 2000), ksNr)
	insertEvent(t, db, leitungID, "lena", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData("z-anna", 500, "Ruecknahme"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if len(data.Stornierungen) != 1 {
		t.Fatalf("expected 1 Storno, got %d", len(data.Stornierungen))
	}

	storno := data.Stornierungen[0]
	if storno.Akteur.UserName != "lena" || storno.Akteur.UserID != leitungID {
		t.Errorf("expected Akteur lena (%d), got %+v", leitungID, storno.Akteur)
	}
	assertBetroffene(t, storno, "anna")
	if storno.Betroffene[0].UserID != annaID || storno.Betroffene[0].Name != "Anna Müller" {
		t.Errorf("expected betroffene anna (%d, 'Anna Müller'), got %+v", annaID, storno.Betroffene[0])
	}
}

// TestGetStornierungen_JedeZahlungTrifftIhrenKassierer: a return over two cashiers' payments yields two events
// (FIFO per payment), each naming only its own cashier.
func TestGetStornierungen_JedeZahlungTrifftIhrenKassierer(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData("z-anna", 2000), ksNr)
	insertEvent(t, db, bobID, "bob", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 2, zahlungData("z-bob", 1500), ksNr)
	insertEvent(t, db, leitungID, "lena", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 3, stornierungData("z-anna", 500, "Ruecknahme Anna"), ksNr)
	insertEvent(t, db, leitungID, "lena", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 4, stornierungData("z-bob", 300, "Ruecknahme Bob"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	byKommentar := stornierungenByKommentar(data.Stornierungen)
	if len(byKommentar) != 2 {
		t.Errorf("expected 2 Stornos, got %d", len(data.Stornierungen))
	}

	assertBetroffene(t, byKommentar["Ruecknahme Anna"], "anna")
	assertBetroffene(t, byKommentar["Ruecknahme Bob"], "bob")
}

// TestGetStornierungen_KorrekturNenntAlleBesteller: a correction over two orderers' positions lists each exactly once,
// even with several positions per orderer.
func TestGetStornierungen_KorrekturNenntAlleBesteller(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	// Anna orders two positions, Bob one; the correction covers all three.
	insertEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 1, bestellungData([]string{"pos-a1", "pos-a2"}, 300), ksNr)
	insertEvent(t, db, bobID, "bob", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 2, bestellungData([]string{"pos-b1"}, 300), ksNr)
	insertEvent(t, db, leitungID, "lena", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData([]string{"pos-a1", "pos-a2", "pos-b1"}, 900, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if len(data.Stornierungen) != 1 {
		t.Fatalf("expected 1 Storno, got %d", len(data.Stornierungen))
	}

	storno := data.Stornierungen[0]
	if storno.Akteur.UserName != "lena" {
		t.Errorf("expected Akteur lena, got %+v", storno.Akteur)
	}
	assertBetroffene(t, storno, "anna", "bob")
}

// TestGetStornierungen_UmbenannteServicekraftErscheintEinmal: a rename mid-session leaves one user_id with two frozen
// user_name values, and a correction over both orders must still name the person once.
func TestGetStornierungen_UmbenannteServicekraftErscheintEinmal(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna_neu", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	// The same staff member orders twice and is renamed in between.
	insertEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 1, bestellungData([]string{"pos-1"}, 300), ksNr)
	insertEvent(t, db, annaID, "anna_neu", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 2, bestellungData([]string{"pos-2"}, 300), ksNr)
	insertEvent(t, db, leitungID, "lena", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData([]string{"pos-1", "pos-2"}, 600, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if len(data.Stornierungen) != 1 {
		t.Fatalf("expected 1 Storno, got %d", len(data.Stornierungen))
	}

	storno := data.Stornierungen[0]
	if len(storno.Betroffene) != 1 {
		t.Fatalf("expected 1 Betroffene, got %d: %+v", len(storno.Betroffene), storno.Betroffene)
	}
	if storno.Betroffene[0].UserID != annaID {
		t.Errorf("expected Betroffene anna (id %d), got %+v", annaID, storno.Betroffene[0])
	}
}

// TestGetStornierungen_KorrekturUmgebuchterPositionFaelltAufAkteurZurueck: a transfer issues fresh position IDs
// (kasse.NewBestellungUmgebuchtEvents) that no order carries, so their correction falls back to the actor.
func TestGetStornierungen_KorrekturUmgebuchterPositionFaelltAufAkteurZurueck(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	// Anna orders at table 1, Bob transfers to table 2 (new position ID), Lena corrects the transferred position.
	insertEvent(t, db, annaID, "anna", "bestellung-aufgenommen:v1", "kassensitzung-1/tisch-1", 1, bestellungData([]string{"pos-a1"}, 300), ksNr)
	umbuchung := bestellungData([]string{"pos-a1"}, 300)
	umbuchung["umbuchungId"] = "u-1"
	umbuchung["quellTischId"] = 1
	umbuchung["zielTischId"] = 2
	umbuchung["gesamtCents"] = 300
	insertEvent(t, db, bobID, "bob", "bestellung-umgebucht:v1", "kassensitzung-1/tisch-1", 2, umbuchung, ksNr)
	umbuchungZiel := bestellungData([]string{"pos-neu"}, 300)
	umbuchungZiel["umbuchungId"] = "u-1"
	umbuchungZiel["quellTischId"] = 1
	umbuchungZiel["zielTischId"] = 2
	umbuchungZiel["gesamtCents"] = 300
	insertEvent(t, db, bobID, "bob", "bestellung-umgebucht:v1", "kassensitzung-1/tisch-2", 1, umbuchungZiel, ksNr)
	insertEvent(t, db, leitungID, "lena", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-2", 2, korrekturData([]string{"pos-neu"}, 300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if len(data.Stornierungen) != 1 {
		t.Fatalf("expected 1 Storno, got %d", len(data.Stornierungen))
	}

	storno := data.Stornierungen[0]
	if storno.Akteur.UserName != "lena" {
		t.Errorf("expected Akteur lena, got %+v", storno.Akteur)
	}
	assertBetroffene(t, storno, "lena")
}

// TestGetStornierungen_DirektverkaufStornoNenntDenVerkaeufer: a direct-sale storno by another user names the seller
// as betroffen.
func TestGetStornierungen_DirektverkaufStornoNenntDenVerkaeufer(t *testing.T) {
	db := dbtest.Open()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	leitungID := createUser(t, db, "Lena Chef", "lena", "active")
	ksNr := createKassensitzung(t, db)

	dvSubject := "kassensitzung-1/direktverkauf-v-1"
	insertEvent(t, db, annaID, "anna", "direktverkauf-getaetigt:v1", dvSubject, 1, direktverkaufData("v-1", 750), ksNr)
	insertEvent(t, db, leitungID, "lena", "direktverkauf-storniert:v1", dvSubject, 2, direktverkaufStornoData("v-1", 250, "DV-Storno"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	if len(data.Stornierungen) != 1 {
		t.Fatalf("expected 1 Storno, got %d", len(data.Stornierungen))
	}

	storno := data.Stornierungen[0]
	if storno.Quelle != "direktverkauf" {
		t.Errorf("expected quelle 'direktverkauf', got %q", storno.Quelle)
	}
	if storno.Akteur.UserName != "lena" {
		t.Errorf("expected Akteur lena, got %+v", storno.Akteur)
	}
	assertBetroffene(t, storno, "anna")
}
