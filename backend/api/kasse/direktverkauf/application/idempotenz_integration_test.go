//go:build integration

package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
)

func cleanDVDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanDVDB %q: %v", stmt, err)
		}
	}
}

func setupDVIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID, ksNr, produktID, varianteID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanDVDB(t, db)
	t.Cleanup(func() {
		cleanDVDB(t, db)
		_ = db.Close()
	})

	ctx = context.Background()

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test', 'offen', now(), now()) RETURNING z_nr",
	).Scan(&ksNr); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ('Bier', 'getraenk', 'regel', 'active', now(), now()) RETURNING id",
	).Scan(&produktID); err != nil {
		t.Fatalf("create produkt: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.5L', 350, 'active', now(), now()) RETURNING id",
		produktID,
	).Scan(&varianteID); err != nil {
		t.Fatalf("create variante: %v", err)
	}

	cmd = Command{
		EventRepo:           kassenjournal_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		// DruckstationRepo: nil — nil-bewacht, keine Druckaufträge im Test
	}
	return
}

// TestDirektverkaufTaetigen_DuplikatVerkaufId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben verkaufId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestDirektverkaufTaetigen_DuplikatVerkaufId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-getaetigt:v1'").Scan(&eventCount); err != nil {
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

// TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict: Existiert für den
// Direktverkauf-Stream bereits version 1 mit einer anderen verkaufId im Datenfeld,
// antwortet der Command mit ErrConflict (echter OCC-Konflikt, keine Duplikat-Einreichung).
func TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(ksNr, verkaufID)

	// Vorab-Insert: version 1 im selben Stream, aber mit anderer verkaufId im Datenfeld.
	// Simuliert einen echten OCC-Konflikt: der Command-Write trifft auf eine bereits belegte
	// (subject, version). Da die verkaufId nicht übereinstimmt, schlägt der Idempotenz-Check fehl.
	andereVerkaufID := uuid.New().String()
	data, err := json.Marshal(map[string]any{
		"verkaufId":         andereVerkaufID,
		"positionen":        []any{},
		"gesamtbetragCents": 0,
		"kommentar":         "",
	})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'test', 'direktverkauf-getaetigt:v1', $2, 1, $3::jsonb, now(), $4)",
		userID, subject, string(data), ksNr,
	); err != nil {
		t.Fatalf("Vorab-Insert: %v", err)
	}

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); !errors.Is(err, ErrConflict) {
		t.Errorf("erwartet ErrConflict, bekam: %v", err)
	}
}
