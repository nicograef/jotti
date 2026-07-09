//go:build integration

package tse_repo

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// einrichtungsUmgebung haelt die Test-DB samt Kassensitzung und Benutzer: Jeder
// Signaturauftrag referenziert ein Kassenjournal-Event (event_id NOT NULL
// UNIQUE), daher braucht jeder offene Auftrag ein eigenes Event.
type einrichtungsUmgebung struct {
	db      *sql.DB
	userID  int
	ksNr    int
	version int
}

func setupEinrichtung(t *testing.T) (Repository, *einrichtungsUmgebung, func(t *testing.T)) {
	t.Helper()
	database := dbpkg.OpenTestDatabase()

	reset := func(t *testing.T) {
		t.Helper()
		stmts := []string{
			"DELETE FROM tse_stoerungen",
			"DELETE FROM tse_signaturauftraege",
			"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassenjournal",
			"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassensitzungen",
			"DELETE FROM users",
			"UPDATE tse_konfiguration SET api_key = '', api_secret = '', tss_id = '', client_id = '' WHERE id = 1",
		}
		for _, stmt := range stmts {
			if _, err := database.Exec(stmt); err != nil {
				t.Fatalf("reset %q: %v", stmt, err)
			}
		}
	}
	reset(t)

	umgebung := &einrichtungsUmgebung{db: database}
	if err := database.QueryRow(
		"INSERT INTO users (name, username, role, status, created_at, updated_at) VALUES ('Test', 'settings-repo-test', 'admin', 'active', NOW(), NOW()) RETURNING id",
	).Scan(&umgebung.userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := database.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&umgebung.ksNr); err != nil {
		t.Fatalf("insert kassensitzung: %v", err)
	}

	return NewRepository(database), umgebung, func(t *testing.T) {
		reset(t)
		_ = database.Close()
	}
}

func (u *einrichtungsUmgebung) insertOffenerAuftrag(t *testing.T, txID string) int {
	t.Helper()
	u.version++
	var eventID int
	if err := u.db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'Test', 'zahlung-kassiert:v1', $2, $3, '{}', NOW(), $4) RETURNING id",
		u.userID, fmt.Sprintf("kassensitzung-%d/tisch-1", u.ksNr), u.version, u.ksNr,
	).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	var auftragID int
	if err := u.db.QueryRow(`
		INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
		VALUES ($1, $2, 'Kassenbeleg-V1', 'Beleg^2.55', 'offen', NOW(), NOW())
		RETURNING id
	`, eventID, txID).Scan(&auftragID); err != nil {
		t.Fatalf("insert auftrag: %v", err)
	}
	return auftragID
}

func (u *einrichtungsUmgebung) status(t *testing.T, auftragID int) string {
	t.Helper()
	var status string
	if err := u.db.QueryRow("SELECT status FROM tse_signaturauftraege WHERE id = $1", auftragID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	return status
}

func gueltigeKonfiguration(t *testing.T, tssID string) tse.Konfiguration {
	t.Helper()
	conf, err := tse.NewKonfiguration("api-key", "api-secret", tssID, "client-1")
	if err != nil {
		t.Fatalf("build konfiguration: %v", err)
	}
	return conf
}

// Der Einrichtungs-Sweep: der Uebergang von nicht konfiguriert zu konfiguriert
// markiert in derselben Transaktion die noch offenen, vor-konfigurationellen
// Auftraege endgueltig als tse_nicht_konfiguriert und schliesst den
// keine_konfiguration-Stoerungszeitraum.
func TestSaveEinrichtung_UebergangSweeptOffeneUndSchliesstStoerung(t *testing.T) {
	repo, umgebung, teardown := setupEinrichtung(t)
	defer teardown(t)
	ctx := context.Background()

	ersterID := umgebung.insertOffenerAuftrag(t, "tx-sweep-1")
	zweiterID := umgebung.insertOffenerAuftrag(t, "tx-sweep-2")
	if _, err := umgebung.db.Exec(
		"INSERT INTO tse_stoerungen (beginn, grund_art, fehlertext) VALUES (NOW(), 'keine_konfiguration', 'keine TSE-Konfiguration')",
	); err != nil {
		t.Fatalf("open stoerung: %v", err)
	}

	if err := repo.SaveEinrichtung(ctx, gueltigeKonfiguration(t, "tss-neu")); err != nil {
		t.Fatalf("SaveEinrichtung: %v", err)
	}

	conf, err := repo.GetTSEKonfiguration(ctx)
	if err != nil {
		t.Fatalf("read konfiguration: %v", err)
	}
	if !conf.IstKonfiguriert() || conf.TssID != "tss-neu" {
		t.Fatalf("expected saved configuration, got %+v", conf)
	}

	if s := umgebung.status(t, ersterID); s != tse.StatusTSENichtKonfiguriert {
		t.Fatalf("expected first auftrag marked tse_nicht_konfiguriert, got %q", s)
	}
	if s := umgebung.status(t, zweiterID); s != tse.StatusTSENichtKonfiguriert {
		t.Fatalf("expected second auftrag marked tse_nicht_konfiguriert, got %q", s)
	}

	var offeneStoerungen int
	if err := umgebung.db.QueryRow(
		"SELECT COUNT(*) FROM tse_stoerungen WHERE ende IS NULL AND grund_art = 'keine_konfiguration'",
	).Scan(&offeneStoerungen); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if offeneStoerungen != 0 {
		t.Fatalf("expected keine_konfiguration stoerung closed, got %d still open", offeneStoerungen)
	}
}

// War die TSE schon vorher konfiguriert (durchgehend vorhandene Konfiguration),
// bleibt es beim reinen Speichern: laufende offene Auftraege werden nie
// versehentlich als nicht konfiguriert markiert.
func TestSaveEinrichtung_DurchgehendKonfiguriertSweeptNicht(t *testing.T) {
	repo, umgebung, teardown := setupEinrichtung(t)
	defer teardown(t)
	ctx := context.Background()

	// TSE ist bereits konfiguriert.
	if err := repo.SaveEinrichtung(ctx, gueltigeKonfiguration(t, "tss-alt")); err != nil {
		t.Fatalf("initial konfiguration: %v", err)
	}
	offenID := umgebung.insertOffenerAuftrag(t, "tx-laufend")

	// Erneutes Speichern (etwa Uebernahme bei bereits konfigurierter TSE) darf
	// den laufenden Auftrag nicht antasten.
	if err := repo.SaveEinrichtung(ctx, gueltigeKonfiguration(t, "tss-neu")); err != nil {
		t.Fatalf("SaveEinrichtung: %v", err)
	}

	if s := umgebung.status(t, offenID); s != tse.StatusOffen {
		t.Fatalf("expected laufender auftrag to stay offen, got %q", s)
	}
	conf, err := repo.GetTSEKonfiguration(ctx)
	if err != nil {
		t.Fatalf("read konfiguration: %v", err)
	}
	if conf.TssID != "tss-neu" {
		t.Fatalf("expected updated configuration, got %+v", conf)
	}
}

// Das Speichern einer leeren Konfiguration (Leeren ueber den
// Zugangsdaten-Endpunkt) ist kein Uebergang zu konfiguriert: Es sweept nichts
// und schliesst keinen keine_konfiguration-Stoerungszeitraum — der
// Dauerzustand ohne Konfiguration gehoert dem Signatur-Worker.
func TestSaveEinrichtung_LeereKonfigurationSweeptNicht(t *testing.T) {
	repo, umgebung, teardown := setupEinrichtung(t)
	defer teardown(t)
	ctx := context.Background()

	offenID := umgebung.insertOffenerAuftrag(t, "tx-leer")
	if _, err := umgebung.db.Exec(
		"INSERT INTO tse_stoerungen (beginn, grund_art, fehlertext) VALUES (NOW(), 'keine_konfiguration', 'keine TSE-Konfiguration')",
	); err != nil {
		t.Fatalf("open stoerung: %v", err)
	}

	leer, err := tse.NewKonfiguration("", "", "", "")
	if err != nil {
		t.Fatalf("build leere konfiguration: %v", err)
	}
	if err := repo.SaveEinrichtung(ctx, leer); err != nil {
		t.Fatalf("SaveEinrichtung: %v", err)
	}

	if s := umgebung.status(t, offenID); s != tse.StatusOffen {
		t.Fatalf("expected auftrag to stay offen, got %q", s)
	}
	var offeneStoerungen int
	if err := umgebung.db.QueryRow(
		"SELECT COUNT(*) FROM tse_stoerungen WHERE ende IS NULL AND grund_art = 'keine_konfiguration'",
	).Scan(&offeneStoerungen); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if offeneStoerungen != 1 {
		t.Fatalf("expected keine_konfiguration stoerung to stay open, got %d open", offeneStoerungen)
	}
}
