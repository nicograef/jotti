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

func TestGetUserID(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	user, err := persistence.GetUserID("nico")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user != 1 {
		t.Fatalf("Expected user ID 1, got %d", user)
	}
}

func TestGetUserID_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	_, err := persistence.GetUserID("nonexistentuser")

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

func TestSetPasswordHash(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	err := persistence.SetPasswordHash(1, "hashedpassword123")

	if err != nil {
		t.Fatalf("Expected no error setting password hash, got %v", err)
	}

	user, err := persistence.GetUser(1)
	if err != nil {
		t.Fatalf("Expected no error getting user, got %v", err)
	}
	if user.PasswordHash != "hashedpassword123" {
		t.Fatalf("Expected password hash 'hashedpassword123', got %s", user.PasswordHash)
	}
}

func TestSetPasswordHash_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	err := persistence.SetPasswordHash(100000, "somehash")

	if err != user.ErrUserNotFound {
		t.Fatalf("Expected user not found error, got %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	userID, err := persistence.CreateUser("Test User", "testuser", "onetimepasswordhash", user.AdminRole)

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
	err := persistence.UpdateUser(2, "Updated Name", "updatedusername", user.ServiceRole, true) // needs to run after TestCreateUser

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedUser, err := persistence.GetUser(2)
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
