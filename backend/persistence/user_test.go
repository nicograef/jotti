//go:build integration

package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/nicograef/jotti/backend/domain/user"
)

func database() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=admin password=admin dbname=jotti sslmode=disable")
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

	persistence := &UserPersistence{DB: db}
	user, err := persistence.GetUser(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Fatalf("Expected user ID 1, got %d", user.ID)
	}
}

func TestGetUser_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	_, err := persistence.GetUser(100000)

	if err != user.ErrUserNotFound {
		t.Fatalf("Expected user not found error, got %v", err)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	user, err := persistence.GetUserByUsername("admin")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("Expected username 'admin', got %s", user.Username)
	}
}

func TestGetUserByUsername_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	_, err := persistence.GetUserByUsername("nonexistentuser")

	if err != user.ErrUserNotFound {
		t.Fatalf("Expected user not found error, got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	users, err := persistence.GetAllUsers()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("Expected at least one user, got %d", len(users))
	}
}

func TestCreateUserWithoutPassword(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	userID, err := persistence.CreateUserWithoutPassword("Test User", "testuser", user.AdminRole)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if userID != 2 { // 2 because the first user is created in the schema migrations
		t.Fatalf("Expected valid user ID, got %d", userID)
	}

}

func TestUpdateUser(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	err := persistence.UpdateUser(1, "Updated Name", "updatedusername", user.ServiceRole, true)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedUser, err := persistence.GetUser(1)
	if err != nil {
		t.Fatalf("Expected no error retrieving user, got %v", err)
	}
	if updatedUser.Name != "Updated Name" || updatedUser.Username != "updatedusername" || updatedUser.Role != user.ServiceRole || !updatedUser.Locked {
		t.Fatalf("User not updated correctly: %+v", updatedUser)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	err := persistence.UpdateUser(100000, "Updated Name", "updatedusername", user.ServiceRole, true)

	if err != user.ErrUserNotFound {
		t.Fatalf("Expected user not found error, got %v", err)
	}
}
