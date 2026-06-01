//go:build integration

package kassenjournal_repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

func createUser(db *sql.DB) (int, error) {
	var userID int
	err := db.QueryRow("INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, now(), now()) RETURNING id", "nico", "nico", "admin", "active", "hashedpassword", "onetimesethash").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func createTisch(db *sql.DB, name string) (int, error) {
	var tischID int
	err := db.QueryRow("INSERT INTO tische (name, status, created_at, updated_at) VALUES ($1, 'active', now(), now()) RETURNING id", name).Scan(&tischID)
	if err != nil {
		return 0, err
	}
	return tischID, nil
}

func createKassensitzung(db *sql.DB) (int, error) {
	var zNr int
	err := db.QueryRow("INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING z_nr", time.Now(), "Test-Sitzung", kasse.KassensitzungStatusOffen).Scan(&zNr)
	if err != nil {
		return 0, err
	}
	return zNr, nil
}

// insertEventRaw inserts an event directly via SQL, bypassing WriteEvent and the projection.
// Use this for test setup where the projection is not relevant.
func insertEventRaw(db *sql.DB, e event.Event, kassensitzungNr int) (int, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		e.UserID, e.UserName, e.Type, e.Subject, e.Version, e.Data, e.Time, kassensitzungNr,
	).Scan(&id)
	return id, err
}

// newTestEvent creates a test event with the given parameters and version.
func newTestEvent(userID int, eventType, subject string, version int, data any) event.Event {
	e, _ := event.New(userID, "nico", eventType, subject, data)
	e.Version = version
	return e
}

// validBestellungData returns valid bestellung-aufgenommen:v1 event data for testing.
func validBestellungData(positionID string, einzelpreis, menge int) map[string]any {
	return map[string]any{
		"bestellungId": "b0000000-0000-0000-0000-000000000001",
		"positionen": []map[string]any{
			{
				"positionId":   positionID,
				"varianteId":   1,
				"produktName":  "Bier",
				"varianteName": "0.5L",
				"kategorie":    "getraenk",
				"einzelpreis":  einzelpreis,
				"menge":        menge,
			},
		},
		"gesamtPreisCents": einzelpreis * menge,
		"kommentar":        "",
	}
}

// validZahlungData returns valid zahlung-kassiert:v1 event data for testing.
func validZahlungData(positionID string, menge, gesamtCents int) map[string]any {
	return map[string]any{
		"zahlungId": "z0000000-0000-0000-0000-000000000001",
		"positionen": []map[string]any{
			{
				"positionId": positionID,
				"menge":      menge,
			},
		},
		"gesamtZahlungCents": gesamtCents,
		"kommentar":          "",
	}
}

func cleanDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM tisch_sessions")
	if err != nil {
		t.Fatalf("Failed to clean tisch_sessions: %v", err)
	}
	_, err = db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete")
	if err != nil {
		t.Fatalf("Failed to disable kassenjournal_no_delete trigger: %v", err)
	}
	_, err = db.Exec("DELETE FROM kassenjournal")
	if err != nil {
		t.Fatalf("Failed to clean kassenjournal table: %v", err)
	}
	_, err = db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete")
	if err != nil {
		t.Fatalf("Failed to enable kassenjournal_no_delete trigger: %v", err)
	}
	_, err = db.Exec("DELETE FROM kassensitzungen")
	if err != nil {
		t.Fatalf("Failed to clean kassensitzungen table: %v", err)
	}
	_, err = db.Exec("DELETE FROM tische")
	if err != nil {
		t.Fatalf("Failed to clean tische table: %v", err)
	}
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}
}

func setup(t *testing.T) (int, int, Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	cleanDB(t, db)

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	ksNr, err := createKassensitzung(db)
	if err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}

	return userID, ksNr, NewRepository(db), func(t *testing.T) {
		cleanDB(t, db)
		db.Close()
	}
}

func TestWriteEvent_TischSession(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	data := validBestellungData("p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}
}

func TestReadEventsBySubject(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject1 := kasse.TischSessionSubject(ksNr, 1)
	subject2 := kasse.TischSessionSubject(ksNr, 42)

	event1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1, map[string]any{"k": "v"})
	event2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1, map[string]any{"k": "v"})
	_, _ = insertEventRaw(repo.DB, event1, ksNr)
	_, _ = insertEventRaw(repo.DB, event2, ksNr)

	events, err := repo.ReadEventsBySubject(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != subject2 {
		t.Fatalf("Expected subject %s, got %s", subject2, events[0].Subject)
	}
}

func TestGetMaxVersion(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject1 := kasse.TischSessionSubject(ksNr, 1)
	subject2 := kasse.TischSessionSubject(ksNr, 2)

	// No events yet
	version, err := repo.GetMaxVersion(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0 for empty subject, got %d", version)
	}

	// Add events
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1, map[string]any{"order": 1})
	e2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 2, map[string]any{"order": 2})
	e3 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1, map[string]any{"order": 3})

	_, _ = insertEventRaw(repo.DB, e1, ksNr)
	_, _ = insertEventRaw(repo.DB, e2, ksNr)
	_, _ = insertEventRaw(repo.DB, e3, ksNr)

	// Should return max version for subject1
	version, err = repo.GetMaxVersion(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 2 {
		t.Fatalf("Expected version 2, got %d", version)
	}

	// Should return max version for subject2
	version, err = repo.GetMaxVersion(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

// --- Projection integration tests ---

func TestWriteEvent_WithTischSessionProjection(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Proj")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000001"
	data := validBestellungData(posID, 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}

	if state.SaldoCents != 700 {
		t.Fatalf("Expected SaldoCents 700, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].PositionID != posID {
		t.Fatalf("Expected position ID %s, got %s", posID, state.UnbezahltePositionen[0].PositionID)
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Fatalf("Expected Menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("Expected 1 ausstehende position, got %d", len(state.AusstehendePositionen))
	}
	if state.TischID != tischID {
		t.Fatalf("Expected TischID %d, got %d", tischID, state.TischID)
	}
	if state.KassensitzungNr != ksNr {
		t.Fatalf("Expected KassensitzungNr %d, got %d", ksNr, state.KassensitzungNr)
	}
	if state.LastEventID != eventID {
		t.Fatalf("Expected LastEventID %d, got %d", eventID, state.LastEventID)
	}
	if state.LastEventVersion != 1 {
		t.Fatalf("Expected LastEventVersion 1, got %d", state.LastEventVersion)
	}
}

func TestReadTischSession_NotFound(t *testing.T) {
	_, _, repo, teardown := setup(t)
	defer teardown(t)

	state, err := repo.ReadTischSession(context.Background(), "kassensitzung-99/tisch-99999")
	if err != nil {
		t.Fatalf("Expected no error for non-existent subject, got %v", err)
	}

	if state.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0, got %d", state.SaldoCents)
	}
	if state.GesamtZahlungenCents != 0 {
		t.Fatalf("Expected GesamtZahlungenCents 0, got %d", state.GesamtZahlungenCents)
	}
	if len(state.UnbezahltePositionen) != 0 {
		t.Fatalf("Expected empty unbezahlte positionen, got %d", len(state.UnbezahltePositionen))
	}
	if len(state.AusstehendePositionen) != 0 {
		t.Fatalf("Expected empty ausstehende positionen, got %d", len(state.AusstehendePositionen))
	}
}

func TestWriteEvent_MultipleEvents_ProjectionCorrect(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Multi")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000002"

	// Write a Bestellung (2x Bier @ 350 = 700 cents)
	bestellungData := validBestellungData(posID, 350, 2)
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	// Write a Zahlung (pay for 1x Bier = 350 cents)
	zahlungData := validZahlungData(posID, 1, 350)
	e2 := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}

	// Saldo: 700 - 350 = 350
	if state.SaldoCents != 350 {
		t.Fatalf("Expected SaldoCents 350, got %d", state.SaldoCents)
	}
	// GesamtZahlungen: 350
	if state.GesamtZahlungenCents != 350 {
		t.Fatalf("Expected GesamtZahlungenCents 350, got %d", state.GesamtZahlungenCents)
	}
	// Unbezahlt: 1 position with Menge 1 (original 2, paid 1)
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 1 {
		t.Fatalf("Expected remaining Menge 1, got %d", state.UnbezahltePositionen[0].Menge)
	}
	// Ausstehend: still 1 position with Menge 2 (no delivery yet)
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("Expected 1 ausstehende position, got %d", len(state.AusstehendePositionen))
	}
	if state.AusstehendePositionen[0].Menge != 2 {
		t.Fatalf("Expected ausstehend Menge 2, got %d", state.AusstehendePositionen[0].Menge)
	}
	if state.LastEventVersion != 2 {
		t.Fatalf("Expected LastEventVersion 2, got %d", state.LastEventVersion)
	}
}

func TestWriteEvent_InvalidData_Rollback(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Rollback")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)

	// Use an unknown event type that ApplyEvent cannot handle → triggers rollback
	e := newTestEvent(userID, "unknown-event:v1", subject, 1, map[string]any{"k": "v"})

	_, err = repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)
	if err == nil {
		t.Fatalf("Expected error for unknown event type, got nil")
	}

	// Verify no event was written (transaction rolled back)
	events, err := repo.ReadEventsBySubject(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading events, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected 0 events after rollback, got %d", len(events))
	}

	// Verify no tisch_session was written
	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0 after rollback, got %d", state.SaldoCents)
	}
}

// --- Kassensitzung integration tests ---

func TestWriteEvent_KassensitzungEroeffnet(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	// The kassensitzung is already created by setup (simulating application layer).
	// WriteEvent for kassensitzung-eroeffnet:v1 only inserts the event into kassenjournal;
	// the kassensitzungen CRUD entity is managed by the application layer.
	datum := "2026-03-22"
	bezeichnung := "Sommerfest Tag 1"
	data := map[string]any{
		"datum":        datum,
		"bezeichnung":  bezeichnung,
		"eroeffnetVon": userID,
	}
	subject := kasse.KassensitzungSubject(ksNr)
	e := newTestEvent(userID, "kassensitzung-eroeffnet:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	// Verify the kassensitzung still exists and is offen
	ks, err := repo.GetOffeneKassensitzung(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ks == nil {
		t.Fatalf("Expected open kassensitzung, got nil")
	}
	if ks.Status != kasse.KassensitzungStatusOffen {
		t.Fatalf("Expected status 'offen', got %s", ks.Status)
	}
}

func TestWriteEvent_TagesabschlussErstellt(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	// Verify kassensitzung is offen
	ks, err := repo.GetOffeneKassensitzung(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ks == nil {
		t.Fatalf("Expected open kassensitzung, got nil")
	}

	subject := kasse.KassensitzungSubject(ksNr)
	data := map[string]any{
		"zNr":               ksNr,
		"zeitraumVon":       time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
		"zeitraumBis":       time.Now().Format(time.RFC3339),
		"umsatzGesamtCents": 15000,
		"stornierungCents":  500,
		"auszahlungenCents": 1000,
		"geldtransitCents":  0,
		"erstelltVon":       userID,
	}
	e := newTestEvent(userID, "tagesabschluss-erstellt:v1", subject, 1, data)

	_, err = repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify kassensitzung is now abgeschlossen
	ksAfter, err := repo.GetOffeneKassensitzung(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ksAfter != nil {
		t.Fatalf("Expected no open kassensitzung after tagesabschluss, got z_nr=%d", ksAfter.ZNr)
	}
}

func TestGetOffeneKassensitzung_NoneOpen(t *testing.T) {
	_, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	// Close the kassensitzung created by setup
	_, err := repo.DB.Exec("UPDATE kassensitzungen SET status = $1 WHERE z_nr = $2", kasse.KassensitzungStatusAbgeschlossen, ksNr)
	if err != nil {
		t.Fatalf("Failed to close kassensitzung: %v", err)
	}

	ks, err := repo.GetOffeneKassensitzung(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ks != nil {
		t.Fatalf("Expected nil for no open kassensitzung, got z_nr=%d", ks.ZNr)
	}
}

// --- Projection rebuild integration tests ---

func TestRebuildAllProjections_EmptyDB(t *testing.T) {
	_, _, repo, teardown := setup(t)
	defer teardown(t)

	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 rebuilt subjects, got %d", count)
	}
}

func TestRebuildAllProjections_RebuildsFromEvents(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Rebuild")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000099"

	// Write events through normal path (creates projection)
	bestellungData := validBestellungData(posID, 500, 3) // 3x 500 = 1500
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	zahlungData := validZahlungData(posID, 1, 500) // pay 1x 500
	e2 := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	// Read expected state before rebuild
	expectedState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading state, got %v", err)
	}

	// Delete projection manually to simulate seed scenario
	_, err = repo.DB.Exec("DELETE FROM tisch_sessions")
	if err != nil {
		t.Fatalf("Failed to delete tisch_sessions: %v", err)
	}

	// Verify projection is gone
	emptyState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if emptyState.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0 after delete, got %d", emptyState.SaldoCents)
	}

	// Rebuild
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 rebuilt subject, got %d", count)
	}

	// Read rebuilt state
	rebuiltState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading rebuilt state, got %v", err)
	}

	// Verify it matches the expected state
	if rebuiltState.SaldoCents != expectedState.SaldoCents {
		t.Fatalf("Expected SaldoCents %d, got %d", expectedState.SaldoCents, rebuiltState.SaldoCents)
	}
	if rebuiltState.GesamtZahlungenCents != expectedState.GesamtZahlungenCents {
		t.Fatalf("Expected GesamtZahlungenCents %d, got %d", expectedState.GesamtZahlungenCents, rebuiltState.GesamtZahlungenCents)
	}
	if len(rebuiltState.UnbezahltePositionen) != len(expectedState.UnbezahltePositionen) {
		t.Fatalf("Expected %d unbezahlte positionen, got %d", len(expectedState.UnbezahltePositionen), len(rebuiltState.UnbezahltePositionen))
	}
	if len(rebuiltState.AusstehendePositionen) != len(expectedState.AusstehendePositionen) {
		t.Fatalf("Expected %d ausstehende positionen, got %d", len(expectedState.AusstehendePositionen), len(rebuiltState.AusstehendePositionen))
	}
	if rebuiltState.LastEventID != expectedState.LastEventID {
		t.Fatalf("Expected LastEventID %d, got %d", expectedState.LastEventID, rebuiltState.LastEventID)
	}
	if rebuiltState.LastEventVersion != expectedState.LastEventVersion {
		t.Fatalf("Expected LastEventVersion %d, got %d", expectedState.LastEventVersion, rebuiltState.LastEventVersion)
	}
}

func TestRebuildAllProjections_MultipleSubjects(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tisch1ID, err := createTisch(repo.DB, "Tisch R1")
	if err != nil {
		t.Fatalf("Failed to create tisch 1: %v", err)
	}
	tisch2ID, err := createTisch(repo.DB, "Tisch R2")
	if err != nil {
		t.Fatalf("Failed to create tisch 2: %v", err)
	}

	subject1 := kasse.TischSessionSubject(ksNr, tisch1ID)
	subject2 := kasse.TischSessionSubject(ksNr, tisch2ID)

	// Write events via raw insert (bypassing projection, simulating seed.sql)
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1,
		validBestellungData("p1-1", 200, 2)) // 400
	_, err = insertEventRaw(repo.DB, e1, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	e2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1,
		validBestellungData("p2-1", 300, 1)) // 300
	_, err = insertEventRaw(repo.DB, e2, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// No projections exist (raw insert bypasses them)

	// Rebuild
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 rebuilt subjects, got %d", count)
	}

	state1, err := repo.ReadTischSession(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error reading state1, got %v", err)
	}
	if state1.SaldoCents != 400 {
		t.Fatalf("Expected SaldoCents 400 for tisch1, got %d", state1.SaldoCents)
	}

	state2, err := repo.ReadTischSession(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error reading state2, got %v", err)
	}
	if state2.SaldoCents != 300 {
		t.Fatalf("Expected SaldoCents 300 for tisch2, got %d", state2.SaldoCents)
	}
}

func TestRebuildAllProjections_SkipsKassensitzungSubjects(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch SkipKS")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	// Insert a kassensitzung event (should be skipped during rebuild)
	ksSubject := kasse.KassensitzungSubject(ksNr)
	ksEvent := newTestEvent(userID, "kassensitzung-eroeffnet:v1", ksSubject, 1, map[string]any{
		"datum":        "2026-03-22",
		"bezeichnung":  "Test",
		"eroeffnetVon": userID,
	})
	_, err = insertEventRaw(repo.DB, ksEvent, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert kassensitzung event: %v", err)
	}

	// Insert a tisch-session event
	tischSubject := kasse.TischSessionSubject(ksNr, tischID)
	tischEvent := newTestEvent(userID, "bestellung-aufgenommen:v1", tischSubject, 1,
		validBestellungData("p1-1", 200, 1))
	_, err = insertEventRaw(repo.DB, tischEvent, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert tisch event: %v", err)
	}

	// Rebuild should only count tisch-session subjects
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 rebuilt subject (only tisch-session), got %d", count)
	}

	state, err := repo.ReadTischSession(context.Background(), tischSubject)
	if err != nil {
		t.Fatalf("Expected no error reading rebuilt state, got %v", err)
	}
	if state.SaldoCents != 200 {
		t.Fatalf("Expected SaldoCents 200, got %d", state.SaldoCents)
	}
}

func TestGetBestellungEventsSinceCursor(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject := kasse.TischSessionSubject(ksNr, 1)

	// Insert events of different types
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("p1", 100, 1))
	id1, err := insertEventRaw(repo.DB, e1, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	e2 := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("p1", 1, 100))
	_, err = insertEventRaw(repo.DB, e2, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	e3 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 3, validBestellungData("p2", 200, 1))
	_, err = insertEventRaw(repo.DB, e3, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Get bestellung events since cursor=id1 (should only return the second bestellung)
	events, err := repo.GetBestellungEventsSinceCursor(context.Background(), id1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 bestellung event since cursor, got %d", len(events))
	}

	// Get all bestellung events (cursor=0)
	allEvents, err := repo.GetBestellungEventsSinceCursor(context.Background(), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(allEvents) != 2 {
		t.Fatalf("Expected 2 bestellung events total, got %d", len(allEvents))
	}
}

func TestWriteEvent_KassensitzungOtherEvent_NoCRUDChange(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject := kasse.KassensitzungSubject(ksNr)

	// Write an anfangsbestand event — should NOT change kassensitzungen CRUD entity status
	data := map[string]any{
		"betragCents": 50000,
		"gesetztVon":  userID,
	}
	e := newTestEvent(userID, "anfangsbestand-gesetzt:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	// Verify kassensitzung is still offen
	ks, err := repo.GetOffeneKassensitzung(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ks == nil {
		t.Fatalf("Expected open kassensitzung, got nil")
	}
	if ks.Status != kasse.KassensitzungStatusOffen {
		t.Fatalf("Expected status 'offen', got %s", ks.Status)
	}
}
