//go:build unit

package user

import (
	"context"
	"strconv"
	"testing"

	"github.com/nicograef/jotti/backend/db"
)

type mockCommandPersistence struct {
	user *User
	err  error
}

func (m *mockCommandPersistence) CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error) {
	return 1, m.err
}

func (m *mockCommandPersistence) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	return m.err
}

func (m *mockCommandPersistence) SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error {
	return m.err
}

func (m *mockCommandPersistence) ActivateUser(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommandPersistence) DeactivateUser(ctx context.Context, id int) error {
	return m.err
}

func TestCreateUser(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{}}

	userId, onetimePassword, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", ServiceRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userId != 1 {
		t.Errorf("expected user ID 1, got %d", userId)
	}
	if len(onetimePassword) != 6 {
		t.Fatalf("Expected password length 6, got %d", len(onetimePassword))
	}
	if _, err := strconv.Atoi(onetimePassword); err != nil {
		t.Fatalf("Expected numeric password, got %s", onetimePassword)
	}
}

func TestCreateUser_Error(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{err: db.ErrDatabase}}

	_, _, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", ServiceRole)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrDatabase {
		t.Errorf("expected error %v, got %v", ErrDatabase, err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{}}

	err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", AdminRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{err: db.ErrDatabase}}

	err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", AdminRole)

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}

}
