//go:build integration

package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
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

	_, _ = db.Exec("DELETE FROM users")

	return db
}

func createTestUser(DB *sql.DB) (int, error) {
	var userID int
	err := DB.QueryRow("INSERT INTO users (name, username, role, status) VALUES ($1, $2, $3, $4) RETURNING id", "nico", "nico", "admin", "active").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func TestGetUser(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	userID, err := createTestUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	user, err := persistence.GetUser(ctx, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != userID {
		t.Fatalf("expected user ID 1, got %d", user.ID)
	}
	if user.Username != "nico" {
		t.Fatalf("expected username 'nico', got %s", user.Username)
	}
	if user.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero created_at, got %v", user.CreatedAt)
	}
	if user.Status != ActiveStatus {
		t.Fatalf("expected user to be active, got %s", user.Status)
	}
	if user.Role != AdminRole {
		t.Fatalf("expected user role 'admin', got %s", user.Role)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM users")

}

func TestGetUser_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	_, err := persistence.GetUser(ctx, 100000)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGetUserID(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	newUserID, err := createTestUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	userID, err := persistence.GetUserID(ctx, "nico")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != newUserID {
		t.Fatalf("expected user ID %d, got %d", newUserID, userID)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
}

func TestGetUserID_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	_, err := persistence.GetUserID(ctx, "nonexistentuser")

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestSetPasswordHash(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	userID, err := createTestUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	persistence := &Persistence{DB: db}
	err = persistence.SetPasswordHash(ctx, userID, "hashedpassword123")

	if err != nil {
		t.Fatalf("expected no error setting password hash, got %v", err)
	}

	user, err := persistence.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("expected no error getting user, got %v", err)
	}
	if user.PasswordHash != "hashedpassword123" {
		t.Fatalf("expected password hash 'hashedpassword123', got %s", user.PasswordHash)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
}

func TestSetPasswordHash_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.SetPasswordHash(ctx, 100000, "somehash")

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}
