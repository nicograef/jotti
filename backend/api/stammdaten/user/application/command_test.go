package application

import (
	"context"
	"regexp"
	"testing"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestCreateUser(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{}, nil)
	userCommand := Command{UserRepo: repo}

	userID, onetimePassword, err := userCommand.CreateUser(context.Background(), "Test User", "testuser", user.ServiceRole)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 1 {
		t.Errorf("expected user ID 1, got %d", userID)
	}
	if !regexp.MustCompile(`^\d{6}$`).MatchString(onetimePassword) {
		t.Errorf("Expected exactly 6 digits, got %s", onetimePassword)
	}
}
