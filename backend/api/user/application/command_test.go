//go:build unit

package application

import (
	"context"
	"strconv"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func TestCreateUser(t *testing.T) {
	repo := user_repo.NewMock([]user.User{}, nil)
	userCommand := Command{UserRepo: repo}

	userId, onetimePassword, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", user.ServiceRole)

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
	repo := user_repo.NewMock([]user.User{}, db.ErrDatabase)
	userCommand := Command{UserRepo: repo}

	_, _, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", user.ServiceRole)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrDatabase {
		t.Errorf("expected error %v, got %v", ErrDatabase, err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	repo := user_repo.NewMock([]user.User{user.User{ID: 1}}, nil)
	userCommand := Command{UserRepo: repo}

	err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", user.AdminRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	repo := user_repo.NewMock([]user.User{}, db.ErrDatabase)
	userCommand := Command{UserRepo: repo}

	err := userCommand.UpdateUser(context.Background(), 1, "Updated User", "updateduser", user.AdminRole)

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}

}
