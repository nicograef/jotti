//go:build integration

package user_repo

import (
	"context"
	"errors"
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

	repo := NewRepository(db)
	user, err := createTestUser(t, repo)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	return user, repo, func(t *testing.T) {
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

	u, _, err := user.NewUser("nico2", "nicousername2", user.AdminRole)
	if err != nil {
		t.Fatalf("Failed to create user user object: %v", err)
	}

	userID, err := repo.CreateUser(context.Background(), u)
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
	u, repo, teardown := setup(t)
	defer teardown(t)

	u.ID = 99999 // Non-existent user ID
	err := repo.UpdateUser(context.Background(), u)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestCreateUser_DuplicateUsernameRejected(t *testing.T) {
	seeded, repo, teardown := setup(t)
	defer teardown(t)

	duplicate, _, err := user.NewUser("someone else", seeded.Username, user.ServiceRole)
	if err != nil {
		t.Fatalf("failed to build user object: %v", err)
	}

	_, err = repo.CreateUser(context.Background(), duplicate)
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists for duplicate username, got %v", err)
	}
}

func TestCreateUser_UsernameNotRecycledAfterSoftDelete(t *testing.T) {
	seeded, repo, teardown := setup(t)
	defer teardown(t)

	seeded.Delete()
	if err := repo.UpdateUser(context.Background(), seeded); err != nil {
		t.Fatalf("expected no error soft-deleting user, got %v", err)
	}

	reused, _, err := user.NewUser("someone else", seeded.Username, user.ServiceRole)
	if err != nil {
		t.Fatalf("failed to build user object: %v", err)
	}

	_, err = repo.CreateUser(context.Background(), reused)
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists; a soft-deleted user's username must not be recycled, got %v", err)
	}
}

func TestGetAllUsers_ExcludesDeletedUsers(t *testing.T) {
	u, repo, teardown := setup(t)
	defer teardown(t)

	u.Delete()
	err := repo.UpdateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("expected no error deleting user, got %v", err)
	}

	users, err := repo.GetAllUsers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users (deleted excluded), got %d", len(users))
	}
}
