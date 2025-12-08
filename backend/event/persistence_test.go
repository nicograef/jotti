//go:build integration

package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func database() *sql.DB {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=admin password=admin dbname=jotti sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping Postgres: %v\n", err)
		os.Exit(1)
	}

	_, _ = db.Exec("DELETE FROM events")
	_, _ = db.Exec("DELETE FROM users")

	return db
}

func createUser(DB *sql.DB) (int, error) {
	var userID int
	err := DB.QueryRow("INSERT INTO users (name, username, role) VALUES ($1, $2, $3) RETURNING id", "nico", "nico", "admin").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func TestWriteEvent(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	event, err := New(Candidate{
		UserID:  userID,
		Type:    "jotti.order.placed:v1",
		Subject: "table:42",
		Data:    map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventID, err := persistence.WriteEvent(ctx, *event)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM events")
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
}

func TestReadEvent(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	event, err := New(Candidate{
		UserID:  userID,
		Type:    "jotti.order.placed:v1",
		Subject: "table:42",
		Data:    map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventID, err := persistence.WriteEvent(ctx, *event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	readEvent, err := persistence.ReadEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if readEvent == nil {
		t.Fatalf("Expected event, got nil")
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
	_, _ = db.ExecContext(ctx, "DELETE FROM events")
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
}

func TestReadEventsBySubject(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	event1, err := New(Candidate{
		UserID:  userID,
		Type:    "jotti.order.placed:v1",
		Subject: "table:42",
		Data:    map[string]any{"order": 1},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	event2, err := New(Candidate{
		UserID:  userID,
		Type:    "jotti.order.placed:v1",
		Subject: "table:42",
		Data:    map[string]any{"order": 2},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	_, _ = persistence.WriteEvent(ctx, *event1)
	_, _ = persistence.WriteEvent(ctx, *event2)

	events, err := persistence.ReadEventsBySubject(ctx, "table:42", []string{"jotti.order.placed:v1"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM events")
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
}
