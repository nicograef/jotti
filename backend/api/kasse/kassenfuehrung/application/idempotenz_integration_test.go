//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

func cleanKassenfuehrungDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM tse_stoerungen",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM betreiber",
		"DELETE FROM vorgang_idempotenz",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanKassenfuehrungDB %q: %v", stmt, err)
		}
	}
}

func setupKassenfuehrungIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanKassenfuehrungDB(t, db)
	t.Cleanup(func() {
		cleanKassenfuehrungDB(t, db)
		_ = db.Close()
	})

	ctx = context.Background()

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := db.Exec(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test', 'offen', now(), now())",
	); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	kjRepo := kassenjournal_repo.NewRepository(db)
	cmd = Command{
		KassenjournalRepo:   kjRepo,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
	}
	return
}

// TestGeldtransitBuchen_DuplikatGeldtransitId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben geldtransitId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestGeldtransitBuchen_DuplikatGeldtransitId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID := setupKassenfuehrungIntegration(t)

	geldtransitID := uuid.New().String()

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 1000, "Test"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 1000, "Test"); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'geldtransit-gebucht:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet 1 Event im kassenjournal, gespeichert: %d", eventCount)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 1 {
		t.Errorf("erwartet 1 tse_signaturauftrag, vorhanden: %d", auftraegeCount)
	}
}

// Dieselbe geldtransitId mit geändertem Betrag ist weder ein Duplikat noch eine
// neue Buchung: Der Command meldet ErrVorgangDatenAbweichend und schreibt kein
// zweites Event.
func TestGeldtransitBuchen_SelbeGeldtransitIdAndereNutzdaten_KeinZweitesEvent(t *testing.T) {
	ctx, cmd, db, userID := setupKassenfuehrungIntegration(t)

	geldtransitID := uuid.New().String()
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 1000, "Test"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 2000, "Test"); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'geldtransit-gebucht:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet weiterhin 1 geldtransit-gebucht:v1-Event, gespeichert: %d", eventCount)
	}

	var vorgaenge int
	if err := db.QueryRow("SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", geldtransitID).Scan(&vorgaenge); err != nil {
		t.Fatalf("vorgang_idempotenz zählen: %v", err)
	}
	if vorgaenge != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", vorgaenge)
	}
}

// veralteteVersionRepo liefert eine um eins zu niedrige Stream-Version und
// erzwingt damit im realen Schreibpfad einen echten OCC-Konflikt: Das Event
// trifft auf eine bereits belegte (subject, version).
type veralteteVersionRepo struct{ kassenjournalRepo }

func (r veralteteVersionRepo) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	version, err := r.kassenjournalRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, err
	}
	return version - 1, nil
}

// Ein echter Versionskonflikt ist keine Duplikat-Einreichung: Eine neue
// geldtransitId gegen eine veraltete Stream-Version scheitert an
// UNIQUE(subject, version) und ergibt weiterhin ErrConflict — weder die stille
// Erfolgsantwort noch ErrVorgangDatenAbweichend. Die Idempotenz-Zeile des
// gescheiterten Vorgangs rollt mit zurück.
func TestGeldtransitBuchen_NeueGeldtransitIdVeralteteVersion_ErrConflict(t *testing.T) {
	ctx, cmd, db, userID := setupKassenfuehrungIntegration(t)

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", uuid.New().String(), "einlage", 1000, "Test"); err != nil {
		t.Fatalf("erster Geldtransit: %v", err)
	}

	cmd.KassenjournalRepo = veralteteVersionRepo{cmd.KassenjournalRepo}

	zweiteID := uuid.New().String()
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", zweiteID, "einlage", 2500, "Nachschlag"); !errors.Is(err, ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'geldtransit-gebucht:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet weiterhin 1 geldtransit-gebucht:v1-Event, gespeichert: %d", eventCount)
	}

	var vorgaenge int
	if err := db.QueryRow("SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", zweiteID).Scan(&vorgaenge); err != nil {
		t.Fatalf("vorgang_idempotenz zählen: %v", err)
	}
	if vorgaenge != 0 {
		t.Errorf("erwartet keine vorgang_idempotenz-Zeile nach Rollback, vorhanden: %d", vorgaenge)
	}
}
