//go:build integration

package user

import (
	"context"
	"database/sql"
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

	return db
}

func TestGetUser(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	user, err := persistence.GetUser(ctx, 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
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
}

func TestGetUser_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	_, err := persistence.GetUser(ctx, 100000)

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGetUserID(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	userID, err := persistence.GetUserID(ctx, "nico")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 1 {
		t.Fatalf("expected user ID 1, got %d", userID)
	}
}

func TestGetUserID_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	_, err := persistence.GetUserID(ctx, "nonexistentuser")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	users, err := persistence.GetAllUsers(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("expected at least one user, got %d", len(users))
	}
}

func TestSetPasswordHash(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.SetPasswordHash(ctx, 1, "hashedpassword123")

	if err != nil {
		t.Fatalf("expected no error setting password hash, got %v", err)
	}

	user, err := persistence.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error getting user, got %v", err)
	}
	if user.PasswordHash != "hashedpassword123" {
		t.Fatalf("expected password hash 'hashedpassword123', got %s", user.PasswordHash)
	}
}

func TestSetPasswordHash_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.SetPasswordHash(ctx, 100000, "somehash")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestCreateUserInDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	userID, err := persistence.CreateUser(ctx, "Test User", "testuser", "onetimepasswordhash", AdminRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID < 1 {
		t.Fatalf("expected valid user ID, got %d", userID)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
}

func TestUpdateUser(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// First create a user to update
	userID, err := persistence.CreateUser(ctx, "Update Test", "updatetest", "hash", ServiceRole)
	if err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	err = persistence.UpdateUser(ctx, userID, "Updated Name", "updatedusername", ServiceRole)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedUser, err := persistence.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("expected no error retrieving user, got %v", err)
	}
	if updatedUser.Name != "Updated Name" || updatedUser.Username != "updatedusername" || updatedUser.Role != ServiceRole {
		t.Fatalf("user not updated correctly: %+v", updatedUser)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
}

func TestUpdateUserInDB_Error(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.UpdateUser(ctx, 100000, "Updated Name", "updatedusername", ServiceRole)

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}
