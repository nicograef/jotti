//go:build integration

package tse_repo

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// einrichtungsUmgebung holds the test DB with Kassensitzung and user; each Signaturauftrag needs its own
// Kassenjournal event (event_id NOT NULL UNIQUE).
type einrichtungsUmgebung struct {
	db      *sql.DB
	userID  int
	ksNr    int
	version int
}

func setupEinrichtung(t *testing.T) (Repository, *einrichtungsUmgebung, func(t *testing.T)) {
	t.Helper()
	database := dbtest.Open()

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

// The Einrichtung sweep: the transition from unconfigured to configured marks the open orders from the unconfigured
// period tse_nicht_konfiguriert and closes the keine_konfiguration Störungszeitraum in the same transaction.
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
		t.Errorf("expected saved configuration, got %+v", conf)
	}

	if s := umgebung.status(t, ersterID); s != tse.StatusTSENichtKonfiguriert {
		t.Errorf("expected first auftrag marked tse_nicht_konfiguriert, got %q", s)
	}
	if s := umgebung.status(t, zweiterID); s != tse.StatusTSENichtKonfiguriert {
		t.Errorf("expected second auftrag marked tse_nicht_konfiguriert, got %q", s)
	}

	var offeneStoerungen int
	if err := umgebung.db.QueryRow(
		"SELECT COUNT(*) FROM tse_stoerungen WHERE ende IS NULL AND grund_art = 'keine_konfiguration'",
	).Scan(&offeneStoerungen); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if offeneStoerungen != 0 {
		t.Errorf("expected keine_konfiguration stoerung closed, got %d still open", offeneStoerungen)
	}
}

// With a TSE already configured, saving again only stores: running open orders are never marked tse_nicht_konfiguriert.
func TestSaveEinrichtung_DurchgehendKonfiguriertSweeptNicht(t *testing.T) {
	repo, umgebung, teardown := setupEinrichtung(t)
	defer teardown(t)
	ctx := context.Background()

	// TSE already configured.
	if err := repo.SaveEinrichtung(ctx, gueltigeKonfiguration(t, "tss-alt")); err != nil {
		t.Fatalf("initial konfiguration: %v", err)
	}
	offenID := umgebung.insertOffenerAuftrag(t, "tx-laufend")

	// Saving again (e.g. a takeover on a configured TSE) must not touch the running order.
	if err := repo.SaveEinrichtung(ctx, gueltigeKonfiguration(t, "tss-neu")); err != nil {
		t.Fatalf("SaveEinrichtung: %v", err)
	}

	if s := umgebung.status(t, offenID); s != tse.StatusOffen {
		t.Errorf("expected laufender auftrag to stay offen, got %q", s)
	}
	conf, err := repo.GetTSEKonfiguration(ctx)
	if err != nil {
		t.Fatalf("read konfiguration: %v", err)
	}
	if conf.TssID != "tss-neu" {
		t.Errorf("expected updated configuration, got %+v", conf)
	}
}

// Saving an empty configuration (clearing via the credentials endpoint) is no transition to configured.
// It sweeps nothing and leaves the keine_konfiguration Störungszeitraum to the Signatur-Worker.
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
		t.Errorf("expected auftrag to stay offen, got %q", s)
	}
	var offeneStoerungen int
	if err := umgebung.db.QueryRow(
		"SELECT COUNT(*) FROM tse_stoerungen WHERE ende IS NULL AND grund_art = 'keine_konfiguration'",
	).Scan(&offeneStoerungen); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if offeneStoerungen != 1 {
		t.Errorf("expected keine_konfiguration stoerung to stay open, got %d open", offeneStoerungen)
	}
}

// A TSS change replaces the current TSS but keeps the old one's master data, which verify its signatures.
func TestUpsertTSEStammdaten_BehaeltJedeTSS(t *testing.T) {
	repo, _, teardown := setupEinrichtung(t)
	defer teardown(t)
	ctx := context.Background()

	// The table refuses DELETE, so each run uses its own serials.
	lauf := fmt.Sprintf("%d", time.Now().UnixNano())
	alt := tse.NewStammdaten("ALT-"+lauf, "ecdsa-plain-SHA256", "ALTKEY==", "ALTCERT", "unixTime")
	neu := tse.NewStammdaten("NEU-"+lauf, "ecdsa-plain-SHA256", "NEUKEY==", "NEUCERT", "unixTime")
	for _, s := range []tse.Stammdaten{alt, neu} {
		if err := repo.UpsertTSEStammdaten(ctx, s); err != nil {
			t.Fatalf("UpsertTSEStammdaten(%s): %v", s.Seriennummer, err)
		}
	}

	aktuell, err := repo.GetTSEStammdaten(ctx)
	if err != nil {
		t.Fatalf("GetTSEStammdaten: %v", err)
	}
	if aktuell.Seriennummer != neu.Seriennummer {
		t.Errorf("current TSS = %q, want %q", aktuell.Seriennummer, neu.Seriennummer)
	}
	jeTSS, err := repo.GetTSEStammdatenJeTSS(ctx)
	if err != nil {
		t.Fatalf("GetTSEStammdatenJeTSS: %v", err)
	}
	for _, want := range []tse.Stammdaten{alt, neu} {
		got := jeTSS[want.Seriennummer]
		if got.PublicKey != want.PublicKey || got.Zertifikat != want.Zertifikat {
			t.Errorf("stammdaten of %s = %+v, want key %q and certificate %q", want.Seriennummer, got, want.PublicKey, want.Zertifikat)
		}
	}
}
