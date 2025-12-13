//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func TestGenerateJWTToken_NotFound(t *testing.T) {
	repo := user_repo.NewMock([]user.User{}, db.ErrNotFound)
	command := Command{UserRepo: repo, JWTSecret: "test-secret"}

	_, err := command.GenerateJWTToken(context.Background(), "nonexistent", "password")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGenerateJWTToken_Success(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret"}

	token, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected a token, got empty string")
	}
}

func TestGenerateJWTToken_InvalidPassword(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret"}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestGenerateJWTToken_HashError(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "invalidhashformat"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret"}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "somepassword")

	if err != ErrTokenGeneration {
		t.Fatalf("expected token generation error, got %v", err)
	}
}

func TestGenerateJWTToken_UserInactive(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.InactiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret"}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if err != ErrNotActive {
		t.Fatalf("expected user not active error, got %v", err)
	}
}
