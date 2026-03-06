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
	err := db.QueryRow("INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at) VALUES ($1, $2, $3, $4, $5, $6, now()) RETURNING id", "nico", "nico", "admin", "active", "hashedpassword", "onetimesethash").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
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

	event, err := event.New(userID, "table.order-placed:v1", "table:42", map[string]any{"k": "v"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventID, err := repo.WriteEvent(context.Background(), event)

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

	event, err := event.New(userID, "table.order-placed:v1", "table:42", map[string]any{"k": "v"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventID, err := repo.WriteEvent(context.Background(), event)
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
	if readEvent.UserID != event.UserID {
		t.Fatalf("Expected user ID %d, got %d", event.UserID, readEvent.UserID)
	}
	if readEvent.Type != event.Type {
		t.Fatalf("Expected event type %s, got %s", event.Type, readEvent.Type)
	}
	if readEvent.Subject != event.Subject {
		t.Fatalf("Expected subject %s, got %s", event.Subject, readEvent.Subject)
	}
	if readEvent.Time.Unix() != event.Time.Unix() {
		t.Fatalf("Expected time %v, got %v", event.Time, readEvent.Time)
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

	event1, err := event.New(userID, "table.order-placed:v1", "table:1", map[string]any{"k": "v"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	event2, err := event.New(userID, "table.order-placed:v1", "table:42", map[string]any{"k": "v"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	_, _ = repo.WriteEvent(context.Background(), event1)
	_, _ = repo.WriteEvent(context.Background(), event2)

	events, err := repo.ReadEventsBySubject(context.Background(), "table:42")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != "table:42" {
		t.Fatalf("Expected subject table:42, got %s", events[0].Subject)
	}
}

func TestReadEventsSinceID(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	// Create multiple events for the same subject
	event1, _ := event.New(userID, "table.order-placed:v1", "table:1", map[string]any{"order": 1})
	event2, _ := event.New(userID, "table.order-placed:v1", "table:1", map[string]any{"order": 2})
	event3, _ := event.New(userID, "table.order-placed:v1", "table:1", map[string]any{"order": 3})

	id1, _ := repo.WriteEvent(context.Background(), event1)
	id2, _ := repo.WriteEvent(context.Background(), event2)
	_, _ = repo.WriteEvent(context.Background(), event3)

	// Read events since id2 (should return event2 and event3)
	events, err := repo.ReadEventsSinceID(context.Background(), "table:1", id2)
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
	events, err = repo.ReadEventsSinceID(context.Background(), "table:1", id1)
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

	// Events for table:1
	event1, _ := event.New(userID, "table.order-placed:v1", "table:1", map[string]any{"order": 1})
	// Events for table:2
	event2, _ := event.New(userID, "table.order-placed:v1", "table:2", map[string]any{"order": 2})

	id1, _ := repo.WriteEvent(context.Background(), event1)
	_, _ = repo.WriteEvent(context.Background(), event2)

	// Read events for table:1 since id1 (should only return table:1 events)
	events, err := repo.ReadEventsSinceID(context.Background(), "table:1", id1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != "table:1" {
		t.Fatalf("Expected subject table:1, got %s", events[0].Subject)
	}
}

func TestGetLastSnapshotID(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "table.snapshot:v1"
	orderType := "table.order-placed:v1"

	// No snapshot exists yet
	id, err := repo.GetLastSnapshotID(context.Background(), "table:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != 0 {
		t.Fatalf("Expected 0 when no snapshot exists, got %d", id)
	}

	// Add some events
	order1, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 1})
	snapshot1, _ := event.New(userID, snapshotType, "table:1", map[string]any{"balance": 100})
	order2, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 2})
	snapshot2, _ := event.New(userID, snapshotType, "table:1", map[string]any{"balance": 200})
	order3, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 3})

	_, _ = repo.WriteEvent(context.Background(), order1)
	snapshotID1, _ := repo.WriteEvent(context.Background(), snapshot1)
	_, _ = repo.WriteEvent(context.Background(), order2)
	snapshotID2, _ := repo.WriteEvent(context.Background(), snapshot2)
	_, _ = repo.WriteEvent(context.Background(), order3)

	// Should return the ID of the most recent snapshot
	id, err = repo.GetLastSnapshotID(context.Background(), "table:1", snapshotType)
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

	snapshotType := "table.snapshot:v1"

	// Add snapshots for different tables
	snapshot1, _ := event.New(userID, snapshotType, "table:1", map[string]any{"balance": 100})
	snapshot2, _ := event.New(userID, snapshotType, "table:2", map[string]any{"balance": 200})

	snapshotID1, _ := repo.WriteEvent(context.Background(), snapshot1)
	snapshotID2, _ := repo.WriteEvent(context.Background(), snapshot2)

	// Get snapshot for table:1
	id, err := repo.GetLastSnapshotID(context.Background(), "table:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != snapshotID1 {
		t.Fatalf("Expected snapshot ID %d for table:1, got %d", snapshotID1, id)
	}

	// Get snapshot for table:2
	id, err = repo.GetLastSnapshotID(context.Background(), "table:2", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != snapshotID2 {
		t.Fatalf("Expected snapshot ID %d for table:2, got %d", snapshotID2, id)
	}
}

func TestReadEventsWithSnapshot(t *testing.T) {
	userID, repo, teardown := setup(t)
	defer teardown(t)

	snapshotType := "table.snapshot:v1"
	orderType := "table.order-placed:v1"

	// Add events: order -> snapshot -> order -> order
	order1, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 1})
	snapshot, _ := event.New(userID, snapshotType, "table:1", map[string]any{"balance": 100})
	order2, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 2})
	order3, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 3})

	_, _ = repo.WriteEvent(context.Background(), order1)
	snapshotID, _ := repo.WriteEvent(context.Background(), snapshot)
	_, _ = repo.WriteEvent(context.Background(), order2)
	_, _ = repo.WriteEvent(context.Background(), order3)

	// Should return snapshot + 2 orders (3 events total, not including order1)
	events, err := repo.ReadEventsWithSnapshot(context.Background(), "table:1", snapshotType)
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

	snapshotType := "table.snapshot:v1"
	orderType := "table.order-placed:v1"

	// Add only orders, no snapshot
	order1, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 1})
	order2, _ := event.New(userID, orderType, "table:1", map[string]any{"order": 2})

	_, _ = repo.WriteEvent(context.Background(), order1)
	_, _ = repo.WriteEvent(context.Background(), order2)

	// Should return all events when no snapshot exists
	events, err := repo.ReadEventsWithSnapshot(context.Background(), "table:1", snapshotType)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
}
