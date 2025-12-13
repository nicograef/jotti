//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func TestGetAllUsers_Success(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Name: "Test User", Username: "testuser", Role: user.ServiceRole}}, nil)
	userQuery := Query{UserRepo: repo}

	users, err := userQuery.GetAllUsers(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].ID != 1 {
		t.Errorf("expected user ID %d, got %d", 1, users[0].ID)
	}
}

func TestGetAllUsers_Error(t *testing.T) {
	userQuery := Query{UserRepo: user_repo.NewMock([]user.User{}, db.ErrDatabase)}

	_, err := userQuery.GetAllUsers(context.Background())

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}
}
