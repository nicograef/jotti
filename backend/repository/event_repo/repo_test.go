//go:build integration

package event_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

// newTestEvent creates a test event with the given parameters and version.
func newTestEvent(userID int, eventType, subject string, version int, data any) event.Event {
	e, _ := event.New(userID, "nico", eventType, subject, data)
	e.Version = version
	return e
}

func setup(t *testing.T) (int, Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	_, err := db.Exec("ALTER TABLE events DISABLE TRIGGER events_no_delete")
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
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	return userID, NewRepository(db), func(t *testing.T) {
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
		_, err = db.Exec("DELETE FROM users")
		if err != nil {
			t.Fatalf("Failed to clean users table: %v", err)
		}

		db.Close()
	}
}

func TestWriteEvent(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	e := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:42", 1, map[string]any{"k": "v"})

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

	eventID, err := repo.WriteEvent(context.Background(), e)
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
	_, _ = repo.WriteEvent(context.Background(), event1)
	_, _ = repo.WriteEvent(context.Background(), event2)

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

func TestReadEventsSinceID(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	// Create multiple events for the same subject
	event1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 1, map[string]any{"order": 1})
	event2 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 2, map[string]any{"order": 2})
	event3 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 3, map[string]any{"order": 3})

	id1, _ := repo.WriteEvent(context.Background(), event1)
	id2, _ := repo.WriteEvent(context.Background(), event2)
	_, _ = repo.WriteEvent(context.Background(), event3)

	// Read events since id2 (should return event2 and event3)
	events, err := repo.ReadEventsSinceID(context.Background(), "tisch:1", id2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	if events[0].ID != id2 {
		t.Fatalf("Expected first event ID %d, got %d", id2, events[0].ID)
	}

	// Read events since id1 (should return all 3)
	events, err = repo.ReadEventsSinceID(context.Background(), "tisch:1", id1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}
}

func TestReadEventsSinceID_DifferentSubjects(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	// Events for tisch:1
	event1 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:1", 1, map[string]any{"order": 1})
	// Events for tisch:2
	event2 := newTestEvent(userID, "tisch.bestellung-aufgegeben:v1", "tisch:2", 1, map[string]any{"order": 2})

	id1, _ := repo.WriteEvent(context.Background(), event1)
	_, _ = repo.WriteEvent(context.Background(), event2)

	// Read events for tisch:1 since id1 (should only return tisch:1 events)
	events, err := repo.ReadEventsSinceID(context.Background(), "tisch:1", id1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != "tisch:1" {
		t.Fatalf("Expected subject tisch:1, got %s", events[0].Subject)
	}
}

func TestGetLastSnapshotID(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "tisch.snapshot:v1"
	orderType := "tisch.bestellung-aufgegeben:v1"

	// No snapshot exists yet
	id, err := repo.GetLastSnapshotID(context.Background(), "tisch:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != 0 {
		t.Fatalf("Expected 0 when no snapshot exists, got %d", id)
	}

	// Add some events
	order1 := newTestEvent(userID, orderType, "tisch:1", 1, map[string]any{"order": 1})
	snapshot1 := newTestEvent(userID, snapshotType, "tisch:1", 2, map[string]any{"balance": 100})
	order2 := newTestEvent(userID, orderType, "tisch:1", 3, map[string]any{"order": 2})
	snapshot2 := newTestEvent(userID, snapshotType, "tisch:1", 4, map[string]any{"balance": 200})
	order3 := newTestEvent(userID, orderType, "tisch:1", 5, map[string]any{"order": 3})

	_, _ = repo.WriteEvent(context.Background(), order1)
	snapshotID1, _ := repo.WriteEvent(context.Background(), snapshot1)
	_, _ = repo.WriteEvent(context.Background(), order2)
	snapshotID2, _ := repo.WriteEvent(context.Background(), snapshot2)
	_, _ = repo.WriteEvent(context.Background(), order3)

	// Should return the ID of the most recent snapshot
	id, err = repo.GetLastSnapshotID(context.Background(), "tisch:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != snapshotID2 {
		t.Fatalf("Expected snapshot ID %d, got %d", snapshotID2, id)
	}

	// Verify first snapshot ID is lower
	if snapshotID1 >= snapshotID2 {
		t.Fatalf("Expected snapshotID1 < snapshotID2, got %d >= %d", snapshotID1, snapshotID2)
	}
}

func TestGetLastSnapshotID_DifferentSubjects(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "tisch.snapshot:v1"

	// Add snapshots for different tables
	snapshot1 := newTestEvent(userID, snapshotType, "tisch:1", 1, map[string]any{"balance": 100})
	snapshot2 := newTestEvent(userID, snapshotType, "tisch:2", 1, map[string]any{"balance": 200})

	snapshotID1, _ := repo.WriteEvent(context.Background(), snapshot1)
	snapshotID2, _ := repo.WriteEvent(context.Background(), snapshot2)

	// Get snapshot for tisch:1
	id, err := repo.GetLastSnapshotID(context.Background(), "tisch:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != snapshotID1 {
		t.Fatalf("Expected snapshot ID %d for tisch:1, got %d", snapshotID1, id)
	}

	// Get snapshot for tisch:2
	id, err = repo.GetLastSnapshotID(context.Background(), "tisch:2", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != snapshotID2 {
		t.Fatalf("Expected snapshot ID %d for tisch:2, got %d", snapshotID2, id)
	}
}

func TestReadEventsWithSnapshot(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "tisch.snapshot:v1"
	orderType := "tisch.bestellung-aufgegeben:v1"

	// Add events: order -> snapshot -> order -> order
	order1 := newTestEvent(userID, orderType, "tisch:1", 1, map[string]any{"order": 1})
	snapshot := newTestEvent(userID, snapshotType, "tisch:1", 2, map[string]any{"balance": 100})
	order2 := newTestEvent(userID, orderType, "tisch:1", 3, map[string]any{"order": 2})
	order3 := newTestEvent(userID, orderType, "tisch:1", 4, map[string]any{"order": 3})

	_, _ = repo.WriteEvent(context.Background(), order1)
	snapshotID, _ := repo.WriteEvent(context.Background(), snapshot)
	_, _ = repo.WriteEvent(context.Background(), order2)
	_, _ = repo.WriteEvent(context.Background(), order3)

	// Should return snapshot + 2 orders (3 events total, not including order1)
	events, err := repo.ReadEventsWithSnapshot(context.Background(), "tisch:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3 events (snapshot + 2 orders), got %d", len(events))
	}
	if events[0].ID != snapshotID {
		t.Fatalf("Expected first event to be snapshot (ID %d), got ID %d", snapshotID, events[0].ID)
	}
	if events[0].Type != snapshotType {
		t.Fatalf("Expected first event type %s, got %s", snapshotType, events[0].Type)
	}
}

func TestReadEventsWithSnapshot_NoSnapshot(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "tisch.snapshot:v1"
	orderType := "tisch.bestellung-aufgegeben:v1"

	// Add only orders, no snapshot
	order1 := newTestEvent(userID, orderType, "tisch:1", 1, map[string]any{"order": 1})
	order2 := newTestEvent(userID, orderType, "tisch:1", 2, map[string]any{"order": 2})

	_, _ = repo.WriteEvent(context.Background(), order1)
	_, _ = repo.WriteEvent(context.Background(), order2)

	// Should return all events when no snapshot exists
	events, err := repo.ReadEventsWithSnapshot(context.Background(), "tisch:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
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

	_, _ = repo.WriteEvent(context.Background(), e1)
	_, _ = repo.WriteEvent(context.Background(), e2)
	_, _ = repo.WriteEvent(context.Background(), e3)

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
