//go:build integration

package tisch_repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	// tisch_favoriten trägt einen Fremdschlüssel auf tische und muss deshalb
	// zuerst geleert werden.
	clean := func() error {
		if _, err := db.Exec("DELETE FROM tisch_favoriten"); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM tische"); err != nil {
			return err
		}
		_, err := db.Exec("DELETE FROM users WHERE username LIKE 'favorit-tester-%'")
		return err
	}

	if err := clean(); err != nil {
		t.Fatalf("Failed to clean tische table: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		if err := clean(); err != nil {
			t.Fatalf("Failed to clean tische table: %v", err)
		}

		_ = db.Close()
	}
}

// setupFavoritenUser legt eine Servicekraft an, deren Markierungen die
// Favoriten-Tests setzen können, und gibt ihre ID zurück. Der Benutzername
// trägt den Testnamen, damit parallele bzw. aufeinanderfolgende Tests nicht am
// Unique-Index kollidieren; aufgeräumt wird in setup.
func setupFavoritenUser(t *testing.T, repo Repository) int {
	t.Helper()

	now := time.Now().UTC()
	var userID int
	err := repo.db.QueryRow(
		`INSERT INTO users (name, username, role, status, created_at, updated_at)
		 VALUES ('Favorit Tester', $1, 'service', 'active', $2, $2) RETURNING id`,
		"favorit-tester-"+t.Name(), now,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	return userID
}

// tischStatus liest den Status direkt aus der Tabelle — GetTable filtert
// 'deleted' weg und taugt deshalb nicht, um ein Soft-Delete nachzuweisen.
func tischStatus(t *testing.T, repo Repository, tischID int) string {
	t.Helper()

	var status string
	if err := repo.db.QueryRow("SELECT status FROM tische WHERE id = $1", tischID).Scan(&status); err != nil {
		t.Fatalf("Failed to read tisch status: %v", err)
	}
	return status
}

func markiereFavorit(t *testing.T, repo Repository, userID, tischID int) {
	t.Helper()

	_, err := repo.db.Exec(
		"INSERT INTO tisch_favoriten (user_id, tisch_id, created_at) VALUES ($1, $2, NOW())",
		userID, tischID,
	)
	if err != nil {
		t.Fatalf("Failed to mark favorit: %v", err)
	}
}

func favoritenVonUser(t *testing.T, repo Repository, userID int) []int {
	t.Helper()

	rows, err := repo.db.Query("SELECT tisch_id FROM tisch_favoriten WHERE user_id = $1 ORDER BY tisch_id", userID)
	if err != nil {
		t.Fatalf("Failed to read favoriten: %v", err)
	}
	defer rows.Close() //nolint:errcheck // Lesefehler deckt rows.Err() ab

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("Failed to scan favorit: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Failed to iterate favoriten: %v", err)
	}
	return ids
}

// Ein gelöschter Tisch verschwindet aus der Tischauswahl; seine Markierungen
// müssen mit ihm gehen, sonst hängen sie unabwählbar in der Tischübersicht der
// betroffenen Servicekräfte. Markierungen anderer Tische bleiben unberührt.
func TestDeleteTableMitFavoritenDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	geloeschtID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Favorit Loeschen", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	bleibtID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Favorit Bleibt", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	userID := setupFavoritenUser(t, repo)
	markiereFavorit(t, repo, userID, geloeschtID)
	markiereFavorit(t, repo, userID, bleibtID)

	err := repo.DeleteTableMitFavoriten(ctx, tisch.Tisch{
		ID: geloeschtID, Name: "Favorit Loeschen", Status: tisch.DeletedStatus, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if status := tischStatus(t, repo, geloeschtID); status != string(tisch.DeletedStatus) {
		t.Errorf("expected status deleted, got %q", status)
	}

	favoriten := favoritenVonUser(t, repo, userID)
	if len(favoriten) != 1 || favoriten[0] != bleibtID {
		t.Errorf("expected only tisch %d to stay marked, got %v", bleibtID, favoriten)
	}
}

// Statuswechsel und Favoriten-Cleanup teilen sich eine Transaktion. Scheitert
// der Schreibvorgang auf tische, muss auch das Löschen der Markierungen
// zurückgerollt werden — sonst verlöre eine Servicekraft ihre Markierungen,
// obwohl der Tisch weiter existiert. Der Fehlschlag wird hier über den
// partiellen Unique-Index auf dem Tischnamen erzwungen (idx_tische_name_active),
// weil das der einzige Weg ist, den zweiten Schreibvorgang scheitern zu lassen,
// nachdem der erste Zeilen entfernt hat.
func TestDeleteTableMitFavoritenDB_RollbackBeiSchreibfehler(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	tischID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Rollback Quelle", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "Rollback Kollision", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	userID := setupFavoritenUser(t, repo)
	markiereFavorit(t, repo, userID, tischID)

	err := repo.DeleteTableMitFavoriten(ctx, tisch.Tisch{
		ID: tischID, Name: "Rollback Kollision", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now,
	})
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("expected already-exists error, got %v", err)
	}

	favoriten := favoritenVonUser(t, repo, userID)
	if len(favoriten) != 1 || favoriten[0] != tischID {
		t.Errorf("favoriten must be rolled back with the failed write, got %v", favoriten)
	}

	unveraendert, err := repo.GetTable(ctx, tischID)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if unveraendert.Name != "Rollback Quelle" {
		t.Errorf("expected tisch name to stay 'Rollback Quelle', got %q", unveraendert.Name)
	}
}

func TestGetAllTablesDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 1", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 2", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected exactly 2 tables, got %d", len(tables))
	}
}

// Regression: Der DSFinV-K-Export benennt Abrechnungskreise vergangener
// Kassensitzungen. Ein nach dem Tagesabschluss gelöschter Tisch muss dort
// weiterhin seinen Namen tragen — GetAllTableNames darf 'deleted' nicht
// wegfiltern, sonst fällt der Export auf "Tisch <ID>" zurück.
func TestGetAllTableNamesDB_EnthaeltGeloeschteTische(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	aktivID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Zelt A1", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	geloeschtID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Stehtisch Bar", Status: tisch.DeletedStatus, CreatedAt: now, UpdatedAt: now})

	namen, err := repo.GetAllTableNames(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if namen[aktivID] != "Zelt A1" {
		t.Errorf("name of active tisch = %q, want Zelt A1", namen[aktivID])
	}
	if namen[geloeschtID] != "Stehtisch Bar" {
		t.Errorf("name of deleted tisch = %q, want Stehtisch Bar", namen[geloeschtID])
	}
}

func TestGetActiveTablesDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 1", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 2", Status: tisch.InactiveStatus, CreatedAt: now, UpdatedAt: now})

	tables, err := repo.GetActiveTables(ctx, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected exactly 1 active table, got %d", len(tables))
	}
}

func TestCreateTableInDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	tableID, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Integration Test Table", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableID < 1 {
		t.Fatalf("expected valid table ID, got %d", tableID)
	}
}

func TestUpdateTableDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	tableID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Update Test Table", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	err := repo.UpdateTable(ctx, tisch.Tisch{ID: tableID, Name: "Updated Table Name", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error getting table, got %v", err)
	}
	if tables[0].Name != "Updated Table Name" {
		t.Fatalf("expected name 'Updated Table Name', got %s", tables[0].Name)
	}
}

func TestUpdateTableDB_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	err := repo.UpdateTable(ctx, tisch.Tisch{ID: 999999, Name: "New Name", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("expected table not found error, got %v", err)
	}
}

// setupSaldo prepares an isolated environment for the saldo-projection tests. It
// clears the journal (and its dependants) so the tisch_sessions projection is
// built solely from the events written in each test, then returns the tisch_repo
// under test together with a kassenjournal_repo (used to write real events that
// drive the projection) and a freshly created user + offene Kassensitzung.
func setupSaldo(t *testing.T) (Repository, kassenjournal_repo.Repository, *sql.DB, int, int, func()) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	cleanSaldo(t, db)

	var userID int
	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('nico', 'nico', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var zNr int
	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&zNr); err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}

	return NewRepository(db), kassenjournal_repo.NewRepository(db), db, userID, zNr, func() {
		cleanSaldo(t, db)
		_ = db.Close()
	}
}

func cleanSaldo(t *testing.T, db *sql.DB) {
	t.Helper()
	// tse_signaturauftraege und tisch_sessions referenzieren kassenjournal
	// (event_id bzw. last_event_id) und müssen zuerst weg — eine signierte
	// Bestellung reiht einen Signaturauftrag ein.
	if _, err := db.Exec("DELETE FROM tse_signaturauftraege"); err != nil {
		t.Fatalf("Failed to clean tse_signaturauftraege: %v", err)
	}
	if _, err := db.Exec("DELETE FROM tisch_sessions"); err != nil {
		t.Fatalf("Failed to clean tisch_sessions: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to disable kassenjournal trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassenjournal"); err != nil {
		t.Fatalf("Failed to clean kassenjournal: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to enable kassenjournal trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassensitzungen"); err != nil {
		t.Fatalf("Failed to clean kassensitzungen: %v", err)
	}
	if _, err := db.Exec("DELETE FROM tische"); err != nil {
		t.Fatalf("Failed to clean tische: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("Failed to clean users: %v", err)
	}
}

// writeBestellung writes a bestellung-aufgenommen:v1 event for the given tisch via
// the real WriteEvent path so the tisch_sessions projection (and its saldo_cents)
// is built exactly as in production — a Journal aufgebaut aus Events.
func writeBestellung(t *testing.T, kjRepo kassenjournal_repo.Repository, userID, zNr, tischID, gesamtCents int) {
	t.Helper()
	subject := kasse.TischSessionSubject(zNr, tischID)
	data := map[string]any{
		"bestellungId": "b0000000-0000-0000-0000-00000000000" + itoaLast(tischID),
		"positionen": []map[string]any{{
			"positionId":       "p0000000-0000-0000-0000-00000000000" + itoaLast(tischID),
			"varianteId":       1,
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"kategorie":        "getraenk",
			"einzelpreisCents": gesamtCents,
			"menge":            1,
		}},
		"gesamtPreisCents": gesamtCents,
		"kommentar":        "",
	}
	e, err := event.New(userID, "nico", "bestellung-aufgenommen:v1", subject, data)
	if err != nil {
		t.Fatalf("Failed to build bestellung event: %v", err)
	}
	if _, err := kjRepo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, zNr); err != nil {
		t.Fatalf("Failed to write bestellung event: %v", err)
	}
}

// itoaLast returns the last decimal digit of n as a string, keeping the fabricated
// UUIDs above unique per tisch without pulling in strconv for a single digit.
func itoaLast(n int) string {
	return string(rune('0' + n%10))
}

func TestGetTischSaldiOffeneSitzungDB(t *testing.T) {
	repo, kjRepo, db, userID, zNr, teardown := setupSaldo(t)
	defer teardown()
	ctx := context.Background()

	tischMitSaldo, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Tisch mit Saldo", Status: tisch.ActiveStatus, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}
	tischOhneSaldo, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Tisch ohne Saldo", Status: tisch.ActiveStatus, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	// Nur der erste Tisch bekommt eine Bestellung → offener Saldo aus der Projektion.
	writeBestellung(t, kjRepo, userID, zNr, tischMitSaldo, 1250)

	saldi, err := repo.GetTischSaldiOffeneSitzung(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if saldi[tischMitSaldo] != 1250 {
		t.Errorf("expected saldo 1250 for tisch %d, got %d", tischMitSaldo, saldi[tischMitSaldo])
	}
	if _, ok := saldi[tischOhneSaldo]; ok {
		t.Errorf("tisch %d without a bestellung must not appear in the saldo map", tischOhneSaldo)
	}

	// Der Saldo einer abgeschlossenen Sitzung darf nicht durchsickern (JOIN auf
	// status = 'offen'). Nach dem Abschluss ist die Map leer.
	if _, err := db.Exec("UPDATE kassensitzungen SET status = 'abgeschlossen' WHERE z_nr = $1", zNr); err != nil {
		t.Fatalf("Failed to close kassensitzung: %v", err)
	}
	saldiNachAbschluss, err := repo.GetTischSaldiOffeneSitzung(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(saldiNachAbschluss) != 0 {
		t.Errorf("closed session must not leak a saldo, got %v", saldiNachAbschluss)
	}
}

func TestTischHatOffenenSaldoDB(t *testing.T) {
	repo, kjRepo, db, userID, zNr, teardown := setupSaldo(t)
	defer teardown()
	ctx := context.Background()

	tischMitSaldo, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Tisch mit Saldo", Status: tisch.ActiveStatus, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}
	tischOhneSaldo, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Tisch ohne Saldo", Status: tisch.ActiveStatus, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	writeBestellung(t, kjRepo, userID, zNr, tischMitSaldo, 800)

	hat, err := repo.TischHatOffenenSaldo(ctx, tischMitSaldo)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !hat {
		t.Errorf("tisch %d with an open bestellung should report an open saldo", tischMitSaldo)
	}

	hat, err = repo.TischHatOffenenSaldo(ctx, tischOhneSaldo)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if hat {
		t.Errorf("tisch %d without a bestellung must not report an open saldo", tischOhneSaldo)
	}

	// Nach Abschluss der Sitzung trägt kein Tisch mehr einen offenen Saldo.
	if _, err := db.Exec("UPDATE kassensitzungen SET status = 'abgeschlossen' WHERE z_nr = $1", zNr); err != nil {
		t.Fatalf("Failed to close kassensitzung: %v", err)
	}
	hat, err = repo.TischHatOffenenSaldo(ctx, tischMitSaldo)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if hat {
		t.Errorf("closed session must not leak an open saldo for tisch %d", tischMitSaldo)
	}
}
