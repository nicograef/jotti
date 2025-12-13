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

func createUser(DB *sql.DB) (int, error) {
	var userID int
	err := DB.QueryRow("INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at) VALUES ($1, $2, $3, $4, $5, $6, now()) RETURNING id", "nico", "nico", "admin", "active", "hashedpassword", "onetimesethash").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func TestWriteEvent(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	repo := Repository{DB: db}
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

	// Cleanup
	_, _ = db.ExecContext(context.Background(), "DELETE FROM events")
	_, _ = db.ExecContext(context.Background(), "DELETE FROM users")
}

func TestReadEvent(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	repo := &Repository{DB: db}
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

	// Cleanup
	_, _ = db.ExecContext(context.Background(), "DELETE FROM events")
	_, _ = db.ExecContext(context.Background(), "DELETE FROM users")
}

func TestReadEvent_NotFound(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	repo := &Repository{DB: db}
	_, err := repo.ReadEvent(context.Background(), 999999)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if errors.Is(err, dbpkg.ErrNotFound) == false {
		t.Fatalf("Expected not found error, got %v", err)
	}

	// Cleanup
	_, _ = db.ExecContext(context.Background(), "DELETE FROM events")
	_, _ = db.ExecContext(context.Background(), "DELETE FROM users")
}

func TestReadEventsBySubject(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	repo := &Repository{DB: db}
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

	// Cleanup
	_, _ = db.ExecContext(context.Background(), "DELETE FROM events")
	_, _ = db.ExecContext(context.Background(), "DELETE FROM users")
}
