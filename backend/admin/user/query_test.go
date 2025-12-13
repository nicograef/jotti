//go:build unit

package user

import (
	"context"
	"testing"
)

type mockQueryPersistence struct {
	ShouldFail bool
	User       *User
}

func (m *mockQueryPersistence) GetUserID(ctx context.Context, username string) (int, error) {
	if m.ShouldFail {
		return 0, ErrUserNotFound
	}
	return m.User.ID, nil
}

func (m *mockQueryPersistence) GetUser(ctx context.Context, id int) (*User, error) {
	if m.ShouldFail {
		return nil, ErrUserNotFound
	}
	return m.User, nil
}

func (m *mockQueryPersistence) GetAllUsers(ctx context.Context) ([]User, error) {
	if m.ShouldFail {
		return nil, ErrDatabase
	}
	return []User{*m.User}, nil
}

func TestGetAllUsers_Success(t *testing.T) {
	mockUser := &User{ID: 1, Name: "Test User", Username: "testuser", Role: ServiceRole}
	userQuery := Query{Persistence: &mockQueryPersistence{User: mockUser}}

	users, err := userQuery.GetAllUsers(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].ID != mockUser.ID {
		t.Errorf("expected user ID %d, got %d", mockUser.ID, users[0].ID)
	}
}

func TestGetAllUsers_Error(t *testing.T) {
	userQuery := Query{Persistence: &mockQueryPersistence{ShouldFail: true}}

	_, err := userQuery.GetAllUsers(context.Background())

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}
}
