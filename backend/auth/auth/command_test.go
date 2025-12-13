//go:build unit

package auth

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
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

func (m *mockCommandPersistence) SetPasswordHash(ctx context.Context, id int, passwordHash string) error {
	return m.err
}

func TestVerifyPasswordAndGetUser_NotFound(t *testing.T) {
	userCommand := Command{Persistence: &mockCommandPersistence{err: db.ErrNotFound}}

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
