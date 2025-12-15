//go:build integration

package user_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

func createTestUser(t *testing.T, repo Repository) (user.User, error) {
	u, _, err := user.NewUser("nico", "nicousername", user.AdminRole)
	if err != nil {
		t.Fatalf("Failed to create user user object: %v", err)
	}

	userID, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	u.ID = userID

	return u, nil
}

func setup(t *testing.T) (user.User, Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}

	user, err := createTestUser(t, Repository{DB: db})
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	return user, Repository{DB: db}, func(t *testing.T) {
		_, err := db.Exec("DELETE FROM users")
		if err != nil {
			t.Fatalf("Failed to clean users table: %v", err)
		}

		db.Close()
	}
}

func TestGetUser(t *testing.T) {
	user, repo, teardown := setup(t)
	defer teardown(t)

	retrievedUser, err := repo.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrievedUser.ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}
	if retrievedUser.Username != user.Username {
		t.Fatalf("expected username %s, got %s", user.Username, retrievedUser.Username)
	}
	if retrievedUser.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero created_at, got %v", retrievedUser.CreatedAt)
	}
	if retrievedUser.Status != user.Status {
		t.Fatalf("expected user to be active, got %s", retrievedUser.Status)
	}
	if retrievedUser.Role != user.Role {
		t.Fatalf("expected user role %s, got %s", user.Role, retrievedUser.Role)
	}
}

func TestGetUser_Error(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	_, err := repo.GetUser(context.Background(), 100000)
	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGetUserByUsername(t *testing.T) {
	user, repo, teardown := setup(t)
	defer teardown(t)

	retrievedUser, err := repo.GetUserByUsername(context.Background(), user.Username)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrievedUser.ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}
}

func TestGetUserByUsername_Error(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	_, err := repo.GetUserByUsername(context.Background(), "nonexistentuser")

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	user, repo, teardown := setup(t)
	defer teardown(t)

	users, err := repo.GetAllUsers(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 users, got %d", len(users))
	}
	if users[0].ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, users[0].ID)
	}
}

func TestCreateUserInDB(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	userID, err := repo.CreateUser(context.Background(), user.User{Name: "Test User", Username: "testuser", Role: user.AdminRole})
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	if userID < 1 {
		t.Fatalf("expected valid user ID, got %d", userID)
	}
}

func TestUpdateUser(t *testing.T) {
	u, repo, teardown := setup(t)
	defer teardown(t)

	u.Name = "Updated Name"
	u.Username = "updatedusername"
	u.Role = user.ServiceRole

	err := repo.UpdateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedUser, err := repo.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("expected no error retrieving user, got %v", err)
	}
	if updatedUser.Name != "Updated Name" || updatedUser.Username != "updatedusername" || updatedUser.Role != user.ServiceRole {
		t.Fatalf("user not updated correctly: %+v", updatedUser)
	}
}

func TestUpdateUserInDB_Error(t *testing.T) {
	_, repo, teardown := setup(t)
	defer teardown(t)

	err := repo.UpdateUser(context.Background(), user.User{ID: 99999})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}
