//go:build unit

package user

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

type mockCommandPersistence struct {
	ShouldFail bool
	User       *User
}

func (m *mockCommandPersistence) CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error) {
	if m.ShouldFail {
		return 0, ErrDatabase
	}
	m.User = &User{
		ID:                  1,
		Name:                name,
		Username:            username,
		Role:                role,
		Status:              ActiveStatus,
		OnetimePasswordHash: onetimePasswordHash,
	}
	return m.User.ID, nil
}

func (m *mockCommandPersistence) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	if m.ShouldFail {
		return ErrDatabase
	}
	m.User = &User{
		ID:       id,
		Name:     name,
		Username: username,
		Role:     role,
	}
	return nil
}

func (m *mockCommandPersistence) SetPasswordHash(ctx context.Context, id int, passwordHash string) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to set password hash")
	}
	m.User.PasswordHash = passwordHash
	m.User.OnetimePasswordHash = ""
	return nil
}

func (m *mockCommandPersistence) SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to set one-time password hash")
	}
	m.User.OnetimePasswordHash = onetimePasswordHash
	m.User.PasswordHash = ""
	return nil
}

func (m *mockCommandPersistence) ActivateUser(ctx context.Context, id int) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to activate user")
	}
	m.User.Status = ActiveStatus
	return nil
}

func (m *mockCommandPersistence) DeactivateUser(ctx context.Context, id int) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to deactivate user")
	}
	m.User.Status = InactiveStatus
	return nil
}

func TestCreateUser(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{User: &User{ID: 1}}}

	user, onetimePassword, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", ServiceRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", user.Username)
	}
	if user.Role != ServiceRole {
		t.Errorf("expected role 'service', got %s", user.Role)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", user.Name)
	}
	if len(onetimePassword) != 6 {
		t.Fatalf("Expected password length 6, got %d", len(onetimePassword))
	}
	if _, err := strconv.Atoi(onetimePassword); err != nil {
		t.Fatalf("Expected numeric password, got %s", onetimePassword)
	}
}

func TestCreateUser_Error(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{ShouldFail: true}}

	_, _, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", ServiceRole)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrDatabase {
		t.Errorf("expected error %v, got %v", ErrDatabase, err)
	}
}

func TestVerifyPasswordAndGetUser_NotFound(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{ShouldFail: true}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "nonexistent", "password")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_Success(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{User: &User{ID: 1, Username: "testuser", PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}}

	user, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "testuser", "testpassword")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected user ID %d, got %d", 1, user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", user.Username)
	}
}

func TestVerifyPasswordAndGetUser_InvalidPassword(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{User: &User{PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "testuser", "wrongpassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_HashError(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{User: &User{PasswordHash: "invalidhashformat"}}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "testuser", "somepassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected password hashing error, got %v", err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{}}

	user, err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", AdminRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != "Updated User" {
		t.Errorf("expected name 'Updated User', got %s", user.Name)
	}
	if user.Username != "updateduser" {
		t.Errorf("expected username 'updateduser', got %s", user.Username)
	}
	if user.Role != AdminRole {
		t.Errorf("expected role 'admin', got %s", user.Role)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{ShouldFail: true}}

	user, err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", AdminRole)

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user on error, got %v", user)
	}
}
