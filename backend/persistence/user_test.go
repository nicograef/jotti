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

func TestUserPersistence_CreateUserWithoutPassword(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &UserPersistence{DB: db}
	userID, err := persistence.CreateUserWithoutPassword("Test User", "testuser", user.AdminRole)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if userID != 1 { // 1 because we reset the database before each test run
		t.Fatalf("Expected valid user ID, got %d", userID)
	}

}
