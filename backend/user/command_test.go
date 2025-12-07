//go:build unit

package user

import (
	"context"
	"strconv"
	"testing"
)

type mockCommandPersistence struct {
	user *User
	err  error
}

func (m *mockCommandPersistence) GetUserID(ctx context.Context, username string) (int, error) {
	return 1, m.err
}

func (m *mockCommandPersistence) GetUser(ctx context.Context, id int) (*User, error) {
	return m.user, m.err
}

func (m *mockCommandPersistence) CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error) {
	return 1, m.err
}

func (m *mockCommandPersistence) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	return m.err
}

func (m *mockCommandPersistence) SetPasswordHash(ctx context.Context, id int, passwordHash string) error {
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
	userCommand := Command{Persistence: &mockCommandPersistence{err: ErrDatabase}}

	_, _, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", ServiceRole)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrDatabase {
		t.Errorf("expected error %v, got %v", ErrDatabase, err)
	}
}

func TestVerifyPasswordAndGetUser_NotFound(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{err: ErrUserNotFound}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "nonexistent", "password")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_Success(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{user: &User{ID: 1, Username: "testuser", PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}}

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
	userCommand := Command{Persistence: &mockCommandPersistence{user: &User{PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "testuser", "wrongpassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_HashError(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{user: &User{PasswordHash: "invalidhashformat"}}}

	_, err := userCommand.VerifyPasswordAndGetUser(context.Background(), "testuser", "somepassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected password hashing error, got %v", err)
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
	userCommand := Command{Persistence: &mockCommandPersistence{err: ErrDatabase}}

	err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", AdminRole)

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}

}
