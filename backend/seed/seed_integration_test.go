//go:build integration

package seed

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func cleanSeedDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturen",
		"DELETE FROM tse_nachsignier_auftraege",
		"DELETE FROM druckauftraege",
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

	// TSE: Zahlungen tragen Signaturdaten mit QR-Code im V0-Format (Belegansicht und -druck).
	var signierteZahlungen int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kassenjournal
		WHERE type = 'zahlung-kassiert:v1' AND data->'tseData'->>'tseQrCodeData' LIKE 'V0;%'`).Scan(&signierteZahlungen); err != nil {
		t.Fatalf("signierte Zahlungen zählen: %v", err)
	}
	if signierteZahlungen == 0 {
		t.Error("keine Zahlung mit TSE-Signaturdaten und V0-QR-Code")
	}

	// Nicht-fiskalische Event-Typen bleiben ohne TSE-Felder.
	var nichtFiskalischMitTSE int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kassenjournal
		WHERE type IN ('kassensitzung-eroeffnet:v1', 'ausgabe-bestaetigt:v1', 'kassensturz-durchgefuehrt:v1')
		  AND data->>'tseTxId' IS NOT NULL`).Scan(&nichtFiskalischMitTSE); err != nil {
		t.Fatalf("nicht-fiskalische Events prüfen: %v", err)
	}
	if nichtFiskalischMitTSE != 0 {
		t.Errorf("%d nicht-fiskalische Events tragen TSE-Felder", nichtFiskalischMitTSE)
	}

	// Nachsignier-Aufträge existieren in allen vier Status; genau einer ist verworfen.
	for status, mindestens := range map[string]int{"offen": 1, "erledigt": 2, "fehlgeschlagen": 1, "verworfen": 1} {
		var anzahl int
		if err := db.QueryRow("SELECT COUNT(*) FROM tse_nachsignier_auftraege WHERE status = $1", status).Scan(&anzahl); err != nil {
			t.Fatalf("Nachsignier-Aufträge (%s) zählen: %v", status, err)
		}
		if anzahl < mindestens {
			t.Errorf("Nachsignier-Aufträge mit Status %s: %d, erwartet mindestens %d", status, anzahl, mindestens)
		}
		if status == "verworfen" && anzahl != 1 {
			t.Errorf("%d verworfene Nachsignier-Aufträge, erwartet genau 1", anzahl)
		}
	}

	// Jeder erledigte Auftrag hat die nachgetragene Signatur-Zeile — und umgekehrt.
	var erledigtOhneSignatur int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tse_nachsignier_auftraege a
		LEFT JOIN tse_signaturen s ON s.tx_id = a.tx_id
		WHERE a.status = 'erledigt' AND s.tx_id IS NULL`).Scan(&erledigtOhneSignatur); err != nil {
		t.Fatalf("erledigte Aufträge ohne Signatur zählen: %v", err)
	}
	if erledigtOhneSignatur != 0 {
		t.Errorf("%d erledigte Nachsignier-Aufträge ohne tse_signaturen-Zeile", erledigtOhneSignatur)
	}
	var signaturOhneErledigt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tse_signaturen s
		LEFT JOIN tse_nachsignier_auftraege a ON a.tx_id = s.tx_id AND a.status = 'erledigt'
		WHERE a.tx_id IS NULL`).Scan(&signaturOhneErledigt); err != nil {
		t.Fatalf("Signaturen ohne erledigten Auftrag zählen: %v", err)
	}
	if signaturOhneErledigt != 0 {
		t.Errorf("%d tse_signaturen-Zeilen ohne erledigten Nachsignier-Auftrag", signaturOhneErledigt)
	}

	// TSE-Konfiguration bleibt leer — der Nachsignier-Worker bleibt inaktiv.
	var tseKonfiguration string
	if err := db.QueryRow("SELECT api_key || api_secret || tss_id || client_id FROM tse_konfiguration WHERE id = 1").Scan(&tseKonfiguration); err != nil {
		t.Fatalf("TSE-Konfiguration abfragen: %v", err)
	}
	if tseKonfiguration != "" {
		t.Errorf("TSE-Konfiguration nicht leer: %q", tseKonfiguration)
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
