//go:build integration

package seed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

// snapshotTischSessions liefert eine deterministische Textrepräsentation der gesamten
// tisch_sessions-Projektion (nach Subject sortiert), um sie vor/nach einem Rebuild zu vergleichen.
func snapshotTischSessions(t *testing.T, db *sql.DB) string {
	t.Helper()
	rows, err := db.Query(`SELECT subject, tisch_id, kassensitzung_nr, saldo_cents,
		unbezahlte_positionen::text, ausstehende_positionen::text, gesamt_zahlungen_cents,
		COALESCE(erste_bestellung_logtime::text, ''), last_event_id, last_event_version
		FROM tisch_sessions ORDER BY subject`)
	if err != nil {
		t.Fatalf("tisch_sessions abfragen: %v", err)
	}
	defer rows.Close()

	var b strings.Builder
	for rows.Next() {
		var (
			subject, unbezahlt, ausstehend, ersteBestellung           string
			tischID, ksNr, saldo, gesamtZahlungen, eventID, eventVers int
		)
		if err := rows.Scan(&subject, &tischID, &ksNr, &saldo, &unbezahlt, &ausstehend,
			&gesamtZahlungen, &ersteBestellung, &eventID, &eventVers); err != nil {
			t.Fatalf("tisch_sessions-Zeile lesen: %v", err)
		}
		fmt.Fprintf(&b, "%s|%d|%d|%d|%s|%s|%d|%s|%d|%d\n", subject, tischID, ksNr, saldo,
			unbezahlt, ausstehend, gesamtZahlungen, ersteBestellung, eventID, eventVers)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("tisch_sessions iterieren: %v", err)
	}
	return b.String()
}

func cleanSeedDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM tse_stoerungen",
		"DELETE FROM druckauftraege",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_position' WHERE kategorie IN ('essen', 'getraenk', 'sonstiges')",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_bestellung' WHERE kategorie = 'abholbon'",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = NULL WHERE kategorie = 'kassenbeleg'",
		"DELETE FROM tisch_favoriten",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM tische",
		"DELETE FROM betreiber",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanSeedDB %q: %v", stmt, err)
		}
	}
}

func TestSeedRun_ErstlaufUndGuard(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	cleanSeedDB(t, db)
	t.Cleanup(func() { cleanSeedDB(t, db) })

	ctx := context.Background()

	// --- Erstlauf: erfolgreiches Seeding ---
	if err := Run(ctx, db); err != nil {
		t.Fatalf("Erstlauf Run: %v", err)
	}

	// Demo-Login: Benutzer hat den jotti123-Passwort-Hash gesetzt.
	var hash sql.NullString
	if err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", "maria").Scan(&hash); err != nil {
		t.Fatalf("Benutzer maria abfragen: %v", err)
	}
	if !hash.Valid || hash.String != demoArgon2idHash {
		t.Error("Benutzer maria: password_hash nicht korrekt gesetzt")
	}

	// Betreiber sichtbar.
	var vereinsname string
	if err := db.QueryRow("SELECT vereinsname FROM betreiber WHERE id = 1").Scan(&vereinsname); err != nil {
		t.Fatalf("Betreiber abfragen: %v", err)
	}
	if vereinsname != "TSV Musterstadt e.V." {
		t.Errorf("Betreiber vereinsname = %q, erwartet %q", vereinsname, "TSV Musterstadt e.V.")
	}

	// Drei Kassensitzungen: Freitag und Samstag abgeschlossen, Sonntag offen.
	for znr, erwartet := range map[int]string{1: "abgeschlossen", 2: "abgeschlossen", 3: "offen"} {
		var status string
		if err := db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = $1", znr).Scan(&status); err != nil {
			t.Fatalf("Kassensitzung %d abfragen: %v", znr, err)
		}
		if status != erwartet {
			t.Errorf("Kassensitzung %d Status = %q, erwartet %q", znr, status, erwartet)
		}
	}

	// Tisch-Favoriten sind angelegt.
	var favoriten int
	if err := db.QueryRow("SELECT COUNT(*) FROM tisch_favoriten").Scan(&favoriten); err != nil {
		t.Fatalf("Favoriten zählen: %v", err)
	}
	if favoriten == 0 {
		t.Error("keine Tisch-Favoriten nach dem Seeding")
	}

	// Kassenjournal enthält Events.
	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&eventCount); err != nil {
		t.Fatalf("Kassenjournal zählen: %v", err)
	}
	if eventCount == 0 {
		t.Fatal("Kassenjournal ist nach dem Seeding leer")
	}

	// Tisch-Session-Projektion: Tisch 1 ist vollständig bezahlt (Saldo 0).
	var saldo int
	if err := db.QueryRow("SELECT saldo_cents FROM tisch_sessions WHERE kassensitzung_nr = 3 AND tisch_id = 1").Scan(&saldo); err != nil {
		t.Fatalf("Tisch-Session Tisch 1 abfragen: %v", err)
	}
	if saldo != 0 {
		t.Errorf("Tisch 1 Saldo = %d, erwartet 0 (bezahlt)", saldo)
	}

	// TSE: Kein Event trägt TSE-Felder im Payload — die Signaturen liegen am Auftrag.
	var eventsMitTSEFeldern int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kassenjournal
		WHERE data->>'tseTxId' IS NOT NULL OR data->>'tseData' IS NOT NULL OR data->>'tseAusfall' IS NOT NULL`).Scan(&eventsMitTSEFeldern); err != nil {
		t.Fatalf("Events mit TSE-Feldern zählen: %v", err)
	}
	if eventsMitTSEFeldern != 0 {
		t.Errorf("%d Events tragen noch TSE-Felder im Payload", eventsMitTSEFeldern)
	}

	// Jedes fiskalische Event hat genau einen Signaturauftrag (event_id UNIQUE sichert
	// höchstens einen); nur Ausgabe und Kassensturz sind im Szenario nicht fiskalisch.
	var fiskalischOhneAuftrag int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kassenjournal e
		LEFT JOIN tse_signaturauftraege a ON a.event_id = e.id
		WHERE e.type NOT IN ('ausgabe-bestaetigt:v1', 'kassensturz-durchgefuehrt:v1')
		  AND a.id IS NULL`).Scan(&fiskalischOhneAuftrag); err != nil {
		t.Fatalf("fiskalische Events ohne Auftrag zählen: %v", err)
	}
	if fiskalischOhneAuftrag != 0 {
		t.Errorf("%d fiskalische Events ohne Signaturauftrag", fiskalischOhneAuftrag)
	}
	var nichtFiskalischMitAuftrag int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tse_signaturauftraege a
		JOIN kassenjournal e ON e.id = a.event_id
		WHERE e.type IN ('ausgabe-bestaetigt:v1', 'kassensturz-durchgefuehrt:v1')`).Scan(&nichtFiskalischMitAuftrag); err != nil {
		t.Fatalf("nicht-fiskalische Events mit Auftrag zählen: %v", err)
	}
	if nichtFiskalischMitAuftrag != 0 {
		t.Errorf("%d nicht-fiskalische Events tragen einen Signaturauftrag", nichtFiskalischMitAuftrag)
	}

	// Signaturaufträge existieren in allen vier Status; genau einer ist verworfen.
	for status, mindestens := range map[string]int{"offen": 1, "erledigt": 2, "fehlgeschlagen": 1, "verworfen": 1} {
		var anzahl int
		if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege WHERE status = $1", status).Scan(&anzahl); err != nil {
			t.Fatalf("Signaturaufträge (%s) zählen: %v", status, err)
		}
		if anzahl < mindestens {
			t.Errorf("Signaturaufträge mit Status %s: %d, erwartet mindestens %d", status, anzahl, mindestens)
		}
		if status == "verworfen" && anzahl != 1 {
			t.Errorf("%d verworfene Signaturaufträge, erwartet genau 1", anzahl)
		}
	}

	// Erledigte Aufträge und gefüllte Signaturspalten bedingen einander; erledigte tragen
	// QR-Daten im V0-Format (Belegansicht und -druck).
	var erledigtOhneSignatur int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tse_signaturauftraege
		WHERE status = 'erledigt' AND (signatur IS NULL OR transaktion_nummer IS NULL
			OR signatur_zaehler IS NULL OR tse_seriennummer IS NULL OR log_time_start IS NULL
			OR log_time_end IS NULL OR qr_code_data NOT LIKE 'V0;%')`).Scan(&erledigtOhneSignatur); err != nil {
		t.Fatalf("erledigte Aufträge ohne Signaturspalten zählen: %v", err)
	}
	if erledigtOhneSignatur != 0 {
		t.Errorf("%d erledigte Signaturaufträge ohne vollständige Signaturspalten", erledigtOhneSignatur)
	}
	var signaturOhneErledigt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tse_signaturauftraege
		WHERE status <> 'erledigt' AND signatur IS NOT NULL`).Scan(&signaturOhneErledigt); err != nil {
		t.Fatalf("nicht erledigte Aufträge mit Signatur zählen: %v", err)
	}
	if signaturOhneErledigt != 0 {
		t.Errorf("%d nicht erledigte Signaturaufträge mit gefüllten Signaturspalten", signaturOhneErledigt)
	}

	// Störungsprotokoll: das aufgelöste Ausfallfenster (Samstag) liegt als geschlossener
	// tse_fehler-Zeitraum vor; kein Zeitraum ist aktiv — das offene Fenster der laufenden
	// Sitzung materialisiert erst zur Laufzeit über Worker und Watchdog.
	var geschlosseneStoerungen, aktiveStoerungen int
	if err := db.QueryRow(`SELECT COUNT(*) FILTER (WHERE ende IS NOT NULL AND grund_art = 'tse_fehler'),
		COUNT(*) FILTER (WHERE ende IS NULL) FROM tse_stoerungen`).Scan(&geschlosseneStoerungen, &aktiveStoerungen); err != nil {
		t.Fatalf("Störungszeiträume zählen: %v", err)
	}
	if geschlosseneStoerungen != 1 {
		t.Errorf("%d geschlossene tse_fehler-Störungszeiträume, erwartet genau 1", geschlosseneStoerungen)
	}
	if aktiveStoerungen != 0 {
		t.Errorf("%d aktive Störungszeiträume, erwartet 0", aktiveStoerungen)
	}

	// Druckstationen: alle fünf Stationen sind mit einer Drucker-IP konfiguriert.
	var stationen, ohneDrucker int
	if err := db.QueryRow("SELECT COUNT(*), COUNT(*) FILTER (WHERE drucker_ip = '') FROM druckstationen").Scan(&stationen, &ohneDrucker); err != nil {
		t.Fatalf("Druckstationen abfragen: %v", err)
	}
	if stationen != 5 || ohneDrucker != 0 {
		t.Errorf("%d Druckstationen, davon %d ohne Drucker-IP — erwartet 5 konfigurierte", stationen, ohneDrucker)
	}

	// Druckaufträge existieren in allen vier Status; genau einer ist verworfen.
	for status, mindestens := range map[string]int{"offen": 1, "gedruckt": 1, "fehlgeschlagen": 2, "verworfen": 1} {
		var anzahl int
		if err := db.QueryRow("SELECT COUNT(*) FROM druckauftraege WHERE status = $1", status).Scan(&anzahl); err != nil {
			t.Fatalf("Druckaufträge (%s) zählen: %v", status, err)
		}
		if anzahl < mindestens {
			t.Errorf("Druckaufträge mit Status %s: %d, erwartet mindestens %d", status, anzahl, mindestens)
		}
		if status == "verworfen" && anzahl != 1 {
			t.Errorf("%d verworfene Druckaufträge, erwartet genau 1", anzahl)
		}
	}

	// Beide Bon-Arten sind vorhanden; fehlgeschlagene Aufträge tragen Fehlertext und
	// ausgeschöpfte Versuche.
	for _, bonArt := range []string{"arbeitsbon", "kassenbeleg"} {
		var anzahl int
		if err := db.QueryRow("SELECT COUNT(*) FROM druckauftraege WHERE bon_art = $1", bonArt).Scan(&anzahl); err != nil {
			t.Fatalf("Druckaufträge (%s) zählen: %v", bonArt, err)
		}
		if anzahl == 0 {
			t.Errorf("keine Druckaufträge mit Bon-Art %s", bonArt)
		}
	}
	var fehlerhaftOhneGrund int
	if err := db.QueryRow(`SELECT COUNT(*) FROM druckauftraege
		WHERE status IN ('fehlgeschlagen', 'verworfen') AND (letzter_fehler IS NULL OR versuche < 3)`).Scan(&fehlerhaftOhneGrund); err != nil {
		t.Fatalf("fehlgeschlagene Druckaufträge prüfen: %v", err)
	}
	if fehlerhaftOhneGrund != 0 {
		t.Errorf("%d fehlgeschlagene/verworfene Druckaufträge ohne Fehlertext oder mit zu wenigen Versuchen", fehlerhaftOhneGrund)
	}

	// TSE-Konfiguration bleibt leer — der Signatur-Worker bleibt inaktiv.
	var tseKonfiguration string
	if err := db.QueryRow("SELECT api_key || api_secret || tss_id || client_id FROM tse_konfiguration WHERE id = 1").Scan(&tseKonfiguration); err != nil {
		t.Fatalf("TSE-Konfiguration abfragen: %v", err)
	}
	if tseKonfiguration != "" {
		t.Errorf("TSE-Konfiguration nicht leer: %q", tseKonfiguration)
	}

	// --- Idempotenz: ein erneuter Projektions-Rebuild ändert nichts ---
	// Der Seeder baut die Projektion bereits selbst auf; ein zweiter Rebuild aus denselben
	// Events muss byte-identische tisch_sessions-Zeilen liefern.
	vorRebuild := snapshotTischSessions(t, db)
	if _, err := kassenjournal_repo.NewRepository(db).RebuildAllProjections(ctx); err != nil {
		t.Fatalf("erneuter RebuildAllProjections: %v", err)
	}
	nachRebuild := snapshotTischSessions(t, db)
	if vorRebuild != nachRebuild {
		t.Errorf("erneuter RebuildAllProjections hat Projektionen verändert:\nvorher:\n%s\nnachher:\n%s", vorRebuild, nachRebuild)
	}

	// --- Zweiter Lauf: Guard greift, ohne etwas zu schreiben ---
	if err := Run(ctx, db); err == nil {
		t.Fatal("zweiter Run sollte am Guard scheitern, lieferte aber keinen Fehler")
	}

	var eventCountNachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&eventCountNachher); err != nil {
		t.Fatalf("Kassenjournal erneut zählen: %v", err)
	}
	if eventCountNachher != eventCount {
		t.Errorf("Guard hat geschrieben: %d Events vorher, %d nachher", eventCount, eventCountNachher)
	}
}
