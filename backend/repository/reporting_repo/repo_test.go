//go:build integration

package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	// tisch_sessions referenziert kassenjournal (last_event_id) und muss zuerst weg —
	// auch Hinterlassenschaften anderer Testpakete auf der geteilten Test-DB.
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

func zahlungData(gesamtCents int) map[string]any {
	return map[string]any{
		"zahlungId": "z0000000-0000-0000-0000-000000000001",
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

func stornierungData(betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"stornierungId":          "s0000000-0000-0000-0000-000000000001",
		"zahlungId":              "z0000000-0000-0000-0000-000000000001",
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

func korrekturData(betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"korrekturId": "k0000000-0000-0000-0000-000000000001",
		"gesamtCents": betragCents,
		"kommentar":   kommentar,
		"positionen": []map[string]any{{
			"produktName":      "Limo",
			"varianteName":     "0.3L",
			"menge":            1,
			"einzelpreisCents": betragCents,
		}},
	}
}

// produktPosition baut eine Fat-Event-Position mit den für die
// Produkt-/Varianten-Statistik nötigen eingefrorenen Feldern (varianteId,
// produktName, varianteName, kategorie, einzelpreisCents, menge).
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

// TestGetProduktStatistik_MengeUndUmsatzMitVorzeichen prüft die beiden getrennten
// Zahlen je Variante über alle beteiligten Event-Typen: Bestellung (+Menge),
// Korrektur (−Menge, geldneutral), Zahlung (+Umsatz), Warenrücknahme (−Umsatz,
// menge-neutral), Direktverkauf (+Menge/+Umsatz) und Direktverkauf-Storno
// (−Umsatz). Zugleich die Konsistenz-Invariante: Σ Produkt-Umsatz == Σ Brutto
// der Umsatz-Positionszeilen derselben Sitzung (kein WHERE-type-Drift).
func TestGetProduktStatistik_MengeUndUmsatzMitVorzeichen(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
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

	// Pommes groß: bestellt 5, korrigiert 1 (geldneutral), kassiert 4, 1 zurückgenommen.
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

	// Cola 0,5 l: Direktverkauf 3, davon 1 storniert (Umsatz-mindernd, menge-neutral).
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
		t.Fatalf("expected 2 variant rows, got %d: %+v", len(zeilen), zeilen)
	}

	// Pommes: ausgegebene Menge 5 − 1 = 4; Umsatz 1200 − 300 = 900.
	pommesZeile := byVariante[10]
	if pommesZeile.ProduktName != "Pommes" || pommesZeile.VarianteName != "groß" || pommesZeile.Kategorie != "essen" {
		t.Errorf("unexpected Pommes identity: %+v", pommesZeile)
	}
	if pommesZeile.AusgegebeneMenge != 4 {
		t.Errorf("expected Pommes menge 4 (5 bestellt − 1 korrigiert), got %d", pommesZeile.AusgegebeneMenge)
	}
	if pommesZeile.UmsatzCents != 900 {
		t.Errorf("expected Pommes umsatz 900 (1200 kassiert − 300 Warenrücknahme), got %d", pommesZeile.UmsatzCents)
	}

	// Cola: Direktverkauf-Storno mindert nur den Umsatz, nicht die Menge.
	colaZeile := byVariante[20]
	if colaZeile.AusgegebeneMenge != 3 {
		t.Errorf("expected Cola menge 3 (Direktverkauf, Storno menge-neutral), got %d", colaZeile.AusgegebeneMenge)
	}
	if colaZeile.UmsatzCents != 500 {
		t.Errorf("expected Cola umsatz 500 (750 Direktverkauf − 250 Storno), got %d", colaZeile.UmsatzCents)
	}

	// Konsistenz-Invariante: Σ Produkt-Umsatz == Σ Brutto der Umsatz-Positionszeilen
	// derselben Sitzung — dieselbe WHERE-type-/Vorzeichenbasis, kein Drift.
	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}
	summeZeilenBrutto := 0
	for _, z := range data.UmsatzProSteuersatz {
		summeZeilenBrutto += z.BruttoCents
	}
	summeProduktUmsatz := 0
	for _, z := range zeilen {
		summeProduktUmsatz += z.UmsatzCents
	}
	if summeProduktUmsatz != summeZeilenBrutto {
		t.Errorf("Σ Produkt-Umsatz %d != Σ Positionszeilen-Brutto %d (WHERE-type-Drift)", summeProduktUmsatz, summeZeilenBrutto)
	}
}

// TestGetProduktStatistik_UmbuchungZaehltNicht stellt sicher, dass
// bestellung-umgebucht:v1 weder Menge noch Umsatz verändert (die Positionen sind
// bereits bei der Bestellung erfasst).
func TestGetProduktStatistik_UmbuchungZaehltNicht(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
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
	if zeilen[0].UmsatzCents != 0 {
		t.Errorf("expected umsatz 0 (nur Bestellung, keine Zahlung), got %d", zeilen[0].UmsatzCents)
	}
}

// TestGetReporting_ResolvesKlarnameIncludingSoftDeleted verifies that the live LEFT JOIN
// resolves the current Klarname for both active and soft-deleted users, while the frozen
// username stays the maßgebliche identity in the event rows.
func TestGetReporting_ResolvesKlarnameIncludingSoftDeleted(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "deleted")
	ksNr := createKassensitzung(t, db)

	// Anna gets more revenue so she sorts first (ORDER BY zahlungen_cents DESC).
	insertEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData(2000), ksNr)
	insertEvent(t, db, bobID, "bob", "zahlung-kassiert:v1", "kassensitzung-1/tisch-2", 1, zahlungData(1000), ksNr)

	insertEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData(500, "Anna storniert"), ksNr)
	insertEvent(t, db, bobID, "bob", "stornierung-erteilt:v1", "kassensitzung-1/tisch-2", 2, stornierungData(300, "Bob storniert"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	// Umsatz pro Servicekraft: username frozen in the event, Klarname resolved live.
	klarnameByUsername := map[string]string{}
	for _, sk := range data.Breakdowns.UmsatzProServicekraft {
		klarnameByUsername[sk.UserName] = sk.Name
	}
	if got := klarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna Klarname 'Anna Müller', got %q", got)
	}
	if got := klarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob Klarname 'Bob Schmidt', got %q", got)
	}

	// Stornierungen: same resolution including the soft-deleted user.
	stornoKlarnameByUsername := map[string]string{}
	for _, s := range data.Stornierungen {
		stornoKlarnameByUsername[s.UserName] = s.Name
	}
	if got := stornoKlarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna storno Klarname 'Anna Müller', got %q", got)
	}
	if got := stornoKlarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob storno Klarname 'Bob Schmidt', got %q", got)
	}
}

// TestGetReporting_IncludesBeideStornoArten verifies that the Stornierungsliste and the
// Stornoquote count both storno kinds: the cash-relevant Warenrücknahme (stornierung-erteilt,
// marked as Bar-Rückgabe) and the geldneutral Korrektur (bestellung-korrigiert).
func TestGetReporting_IncludesBeideStornoArten(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData(2000), ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData(500, "Warenruecknahme"), ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData(300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	if len(data.Stornierungen) != 2 {
		t.Fatalf("expected 2 Stornierungen (Warenrücknahme + Korrektur), got %d", len(data.Stornierungen))
	}

	byKommentar := map[string]reporting.StornierungDetail{}
	for _, s := range data.Stornierungen {
		byKommentar[s.Kommentar] = s
	}

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

	// Die Stornoquote/Summe zählt beide Arten (500 + 300) und beide Events.
	if data.Summary.GesamtStornierungenCents != 800 {
		t.Errorf("expected gesamt stornierungen 800 (beide Arten), got %d", data.Summary.GesamtStornierungenCents)
	}
	if data.Summary.AnzahlStornierungen != 2 {
		t.Errorf("expected anzahl stornierungen 2, got %d", data.Summary.AnzahlStornierungen)
	}
}

// Die Brutto-Positionszeilen müssen kassenwirksame Warenrücknahmen als
// negative Zeilen einbeziehen: Sonst übersteigt die Summe über alle
// Steuersätze den Gesamtumsatz und divergiert vom DSFinV-K-Export.
// Geldneutrale Korrekturen bleiben außen vor (kein Umsatz). Das Repo liefert
// unaggregierte Zeilen; die Aufschlüsselung rechnet die Anwendungsschicht.
func TestGetReporting_UmsatzProSteuersatzZiehtWarenruecknahmeAb(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData(2000), ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData(500, "Warenruecknahme"), ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData(300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	// Zwei Zeilen: die Zahlung positiv, die Warenrücknahme negativ (jeweils regel).
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

// TestGetReporting_MetadatenAusJournalEvents verifiziert, dass der Berichtskopf
// seine Metadaten rein aus den Journal-Events projiziert: Eröffnungs- und
// Abschlusszeitpunkt, den abschließenden Benutzer (eingefrorener user_name) und
// die Kassensturz-Differenz aus dem kassensturz-durchgefuehrt:v1-Event.
func TestGetReporting_MetadatenAusJournalEvents(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
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

// TestGetReporting_MetadatenLeerOhneAbschlussEvents stellt sicher, dass die
// Projektion für eine noch nicht abgeschlossene Sitzung (nur Eröffnung, kein
// Kassensturz/Tagesabschluss) die optionalen Metadaten sauber leer lässt statt
// beim NULL-Scan zu scheitern.
func TestGetReporting_MetadatenLeerOhneAbschlussEvents(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
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
