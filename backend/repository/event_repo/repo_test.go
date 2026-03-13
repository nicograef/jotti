//go:build integration

package event_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
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

// insertEventRaw inserts an event directly via SQL, bypassing WriteEvent and the projection.
// Use this for test setup where the projection is not relevant.
func insertEventRaw(db *sql.DB, e event.Event) (int, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO events (user_id, user_name, type, subject, version, data, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		e.UserID, e.UserName, e.Type, e.Subject, e.Version, e.Data, e.Time,
	).Scan(&id)
	return id, err
}

// newTestEvent creates a test event with the given parameters and version.
func newTestEvent(userID int, eventType, subject string, version int, data any) event.Event {
	e, _ := event.New(userID, "nico", eventType, subject, data)
	e.Version = version
	return e
}

// validBestellungData returns valid bestellungAufgegebenV1 event data for testing.
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

// validZahlungData returns valid zahlungRegistriertV1 event data for testing.
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
	_, err := db.Exec("DELETE FROM table_state")
	if err != nil {
		t.Fatalf("Failed to clean table_state: %v", err)
	}
	_, err = db.Exec("ALTER TABLE events DISABLE TRIGGER events_no_delete")
	if err != nil {
		t.Fatalf("Failed to disable events_no_delete trigger: %v", err)
	}
	_, err = db.Exec("DELETE FROM events")
	if err != nil {
		t.Fatalf("Failed to clean events table: %v", err)
	}
	_, err = db.Exec("ALTER TABLE events ENABLE TRIGGER events_no_delete")
	if err != nil {
		t.Fatalf("Failed to enable events_no_delete trigger: %v", err)
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

func setup(t *testing.T) (int, Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	cleanDB(t, db)

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	return userID, NewRepository(db), func(t *testing.T) {
		cleanDB(t, db)
		db.Close()
	}
}

func TestWriteEvent(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := fmt.Sprintf("tisch:%d", tischID)
	data := validBestellungData("p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}
}

func TestReadEvent(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	e := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:42", 1, map[string]any{"k": "v"})

	eventID, err := insertEventRaw(repo.DB, e)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	readEvent, err := repo.ReadEvent(context.Background(), eventID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if readEvent.ID != eventID {
		t.Fatalf("Expected event ID %d, got %d", eventID, readEvent.ID)
	}
	if readEvent.UserID != e.UserID {
		t.Fatalf("Expected user ID %d, got %d", e.UserID, readEvent.UserID)
	}
	if readEvent.UserName != "nico" {
		t.Fatalf("Expected user name 'nico', got %s", readEvent.UserName)
	}
	if readEvent.Version != 1 {
		t.Fatalf("Expected version 1, got %d", readEvent.Version)
	}
	if readEvent.Type != e.Type {
		t.Fatalf("Expected event type %s, got %s", e.Type, readEvent.Type)
	}
	if readEvent.Subject != e.Subject {
		t.Fatalf("Expected subject %s, got %s", e.Subject, readEvent.Subject)
	}
	if readEvent.Time.Unix() != e.Time.Unix() {
		t.Fatalf("Expected time %v, got %v", e.Time, readEvent.Time)
	}
	var data map[string]any
	err = json.Unmarshal(readEvent.Data, &data)
	if err != nil {
		t.Fatalf("Expected data to be map[string]any, got %T", readEvent.Data)
	}
	if data["k"] != "v" {
		t.Fatalf("Expected data k=v, got k=%v", data["k"])
	}
}

func TestReadEvent_NotFound(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	_, err := repo.ReadEvent(context.Background(), 999999)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if errors.Is(err, dbpkg.ErrNotFound) == false {
		t.Fatalf("Expected not found error, got %v", err)
	}
}

func TestReadEventsBySubject(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	event1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 1, map[string]any{"k": "v"})
	event2 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:42", 1, map[string]any{"k": "v"})
	_, _ = insertEventRaw(repo.DB, event1)
	_, _ = insertEventRaw(repo.DB, event2)

	events, err := repo.ReadEventsBySubject(context.Background(), "tisch:42")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != "tisch:42" {
		t.Fatalf("Expected subject tisch:42, got %s", events[0].Subject)
	}
}

func TestGetMaxVersion(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	// No events yet
	version, err := repo.GetMaxVersion(context.Background(), "tisch:1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0 for empty subject, got %d", version)
	}

	// Add events
	e1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 1, map[string]any{"order": 1})
	e2 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 2, map[string]any{"order": 2})
	e3 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:2", 1, map[string]any{"order": 3})

	_, _ = insertEventRaw(repo.DB, e1)
	_, _ = insertEventRaw(repo.DB, e2)
	_, _ = insertEventRaw(repo.DB, e3)

	// Should return max version for tisch:1
	version, err = repo.GetMaxVersion(context.Background(), "tisch:1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 2 {
		t.Fatalf("Expected version 2, got %d", version)
	}

	// Should return max version for tisch:2
	version, err = repo.GetMaxVersion(context.Background(), "tisch:2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

// --- Projection integration tests ---

func TestWriteEvent_WithProjection(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Proj")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := fmt.Sprintf("tisch:%d", tischID)
	posID := "p0000000-0000-0000-0000-000000000001"
	data := validBestellungData(posID, 350, 2)
	e := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	state, err := repo.ReadTableState(context.Background(), tischID)
	if err != nil {
		t.Fatalf("Expected no error reading table state, got %v", err)
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
	if len(state.UngeliefertePositionen) != 1 {
		t.Fatalf("Expected 1 ungelieferte position, got %d", len(state.UngeliefertePositionen))
	}
	if state.LastEventID != eventID {
		t.Fatalf("Expected LastEventID %d, got %d", eventID, state.LastEventID)
	}
	if state.LastEventVersion != 1 {
		t.Fatalf("Expected LastEventVersion 1, got %d", state.LastEventVersion)
	}
}

func TestReadTableState_NotFound(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	state, err := repo.ReadTableState(context.Background(), 99999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent tisch, got %v", err)
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
	if len(state.UngeliefertePositionen) != 0 {
		t.Fatalf("Expected empty ungelieferte positionen, got %d", len(state.UngeliefertePositionen))
	}
}

func TestWriteEvent_MultipleEvents_ProjectionCorrect(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Multi")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := fmt.Sprintf("tisch:%d", tischID)
	posID := "p0000000-0000-0000-0000-000000000002"

	// Write a Bestellung (2x Bier @ 350 = 700 cents)
	bestellungData := validBestellungData(posID, 350, 2)
	e1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	// Write a Zahlung (pay for 1x Bier = 350 cents)
	zahlungData := validZahlungData(posID, 1, 350)
	e2 := newTestEvent(userID, "tisch.zahlung-registriert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	state, err := repo.ReadTableState(context.Background(), tischID)
	if err != nil {
		t.Fatalf("Expected no error reading table state, got %v", err)
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
	// Ungeliefert: still 1 position with Menge 2 (no delivery yet)
	if len(state.UngeliefertePositionen) != 1 {
		t.Fatalf("Expected 1 ungelieferte position, got %d", len(state.UngeliefertePositionen))
	}
	if state.UngeliefertePositionen[0].Menge != 2 {
		t.Fatalf("Expected ungeliefert Menge 2, got %d", state.UngeliefertePositionen[0].Menge)
	}
	if state.LastEventVersion != 2 {
		t.Fatalf("Expected LastEventVersion 2, got %d", state.LastEventVersion)
	}
}

func TestWriteEvent_InvalidData_Rollback(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Rollback")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := fmt.Sprintf("tisch:%d", tischID)

	// Use an unknown event type that ApplyEvent cannot handle → triggers rollback
	e := newTestEvent(userID, "tisch.unknown-event:v1", subject, 1, map[string]any{"k": "v"})

	_, err = repo.WriteEvent(context.Background(), e)
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

	// Verify no table_state was written
	state, err := repo.ReadTableState(context.Background(), tischID)
	if err != nil {
		t.Fatalf("Expected no error reading table state, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0 after rollback, got %d", state.SaldoCents)
	}
}

// --- Projection rebuild integration tests ---

func TestRebuildAllProjections_EmptyDB(t *testing.T) {
	_, repo, teardown := setup(t)
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
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.DB, "Tisch Rebuild")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := fmt.Sprintf("tisch:%d", tischID)
	posID := "p0000000-0000-0000-0000-000000000099"

	// Write events through normal path (creates projection)
	bestellungData := validBestellungData(posID, 500, 3) // 3x 500 = 1500
	e1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	zahlungData := validZahlungData(posID, 1, 500) // pay 1x 500
	e2 := newTestEvent(userID, "tisch.zahlung-registriert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	// Read expected state before rebuild
	expectedState, err := repo.ReadTableState(context.Background(), tischID)
	if err != nil {
		t.Fatalf("Expected no error reading state, got %v", err)
	}

	// Delete projection manually to simulate seed scenario
	_, err = repo.DB.Exec("DELETE FROM table_state")
	if err != nil {
		t.Fatalf("Failed to delete table_state: %v", err)
	}

	// Verify projection is gone
	emptyState, err := repo.ReadTableState(context.Background(), tischID)
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
	rebuiltState, err := repo.ReadTableState(context.Background(), tischID)
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
	if len(rebuiltState.UngeliefertePositionen) != len(expectedState.UngeliefertePositionen) {
		t.Fatalf("Expected %d ungelieferte positionen, got %d", len(expectedState.UngeliefertePositionen), len(rebuiltState.UngeliefertePositionen))
	}
	if rebuiltState.LastEventID != expectedState.LastEventID {
		t.Fatalf("Expected LastEventID %d, got %d", expectedState.LastEventID, rebuiltState.LastEventID)
	}
	if rebuiltState.LastEventVersion != expectedState.LastEventVersion {
		t.Fatalf("Expected LastEventVersion %d, got %d", expectedState.LastEventVersion, rebuiltState.LastEventVersion)
	}
}

func TestRebuildAllProjections_MultipleSubjects(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	tisch1ID, err := createTisch(repo.DB, "Tisch R1")
	if err != nil {
		t.Fatalf("Failed to create tisch 1: %v", err)
	}
	tisch2ID, err := createTisch(repo.DB, "Tisch R2")
	if err != nil {
		t.Fatalf("Failed to create tisch 2: %v", err)
	}

	subject1 := fmt.Sprintf("tisch:%d", tisch1ID)
	subject2 := fmt.Sprintf("tisch:%d", tisch2ID)

	// Write events via raw insert (bypassing projection, simulating seed.sql)
	e1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject1, 1,
		validBestellungData("p1-1", 200, 2)) // 400
	_, err = insertEventRaw(repo.DB, e1)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	e2 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", subject2, 1,
		validBestellungData("p2-1", 300, 1)) // 300
	_, err = insertEventRaw(repo.DB, e2)
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

	state1, err := repo.ReadTableState(context.Background(), tisch1ID)
	if err != nil {
		t.Fatalf("Expected no error reading state1, got %v", err)
	}
	if state1.SaldoCents != 400 {
		t.Fatalf("Expected SaldoCents 400 for tisch1, got %d", state1.SaldoCents)
	}

	state2, err := repo.ReadTableState(context.Background(), tisch2ID)
	if err != nil {
		t.Fatalf("Expected no error reading state2, got %v", err)
	}
	if state2.SaldoCents != 300 {
		t.Fatalf("Expected SaldoCents 300 for tisch2, got %d", state2.SaldoCents)
	}
}
