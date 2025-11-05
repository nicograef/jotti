//go:build unit

package user

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type MockUserPersistence struct {
	ShouldFail          bool
	MockUser            *User
	PasswordHash        string
	OnetimePasswordHash string
}

func (m *MockUserPersistence) CreateUser(name, username, onetimePasswordHash string, role Role) (int, error) {
	if m.ShouldFail {
		return 0, ErrDatabase
	}
	m.MockUser = &User{
		ID:       1,
		Name:     name,
		Username: username,
		Role:     role,
		Locked:   false,
	}
	return m.MockUser.ID, nil
}

func (m *MockUserPersistence) GetUserByUsername(username string) (*User, error) {
	if m.ShouldFail {
		return nil, ErrUserNotFound
	}
	return m.MockUser, nil
}

func (m *MockUserPersistence) GetUser(id int) (*User, error) {
	if m.ShouldFail {
		return nil, ErrUserNotFound
	}
	return m.MockUser, nil
}

func (m *MockUserPersistence) GetAllUsers() ([]*User, error) {
	if m.ShouldFail {
		return nil, ErrDatabase
	}
	return []*User{m.MockUser}, nil
}

func (m *MockUserPersistence) UpdateUser(id int, name, username string, role Role, locked bool) error {
	if m.ShouldFail {
		return ErrDatabase
	}
	m.MockUser = &User{
		ID:       id,
		Name:     name,
		Username: username,
		Role:     role,
		Locked:   locked,
	}
	return nil
}

func (m *MockUserPersistence) SetPasswordHash(username, passwordHash string) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to set password hash")
	}
	m.PasswordHash = passwordHash
	return nil
}

func (m *MockUserPersistence) GetPasswordHash(username string) (string, error) {
	if m.ShouldFail {
		return "", ErrUserNotFound
	}
	return m.PasswordHash, nil
}

func (m *MockUserPersistence) SetOnetimePasswordHash(username, passwordHash string) error {
	if m.ShouldFail {
		return fmt.Errorf("failed to set one-time password hash")
	}
	m.OnetimePasswordHash = passwordHash
	return nil
}

func (m *MockUserPersistence) GetOnetimePasswordHash(username string) (string, error) {
	if m.ShouldFail {
		return "", ErrUserNotFound
	}
	return m.OnetimePasswordHash, nil
}

func TestCreateUser(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{MockUser: &User{ID: 1}}}

	user, onetimePassword, err := userService.CreateUser("Test User", "testuser", ServiceRole)

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
	userService := Service{DB: &MockUserPersistence{ShouldFail: true}}

	_, _, err := userService.CreateUser("Test User", "testuser", ServiceRole)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrDatabase {
		t.Errorf("expected error %v, got %v", ErrDatabase, err)
	}
}

func TestVerifyPasswordAndGetUser_NotFound(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{ShouldFail: true}}

	_, err := userService.VerifyPasswordAndGetUser("nonexistent", "password")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_Success(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}

	user, err := userService.VerifyPasswordAndGetUser("testuser", "testpassword")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected user ID %d, got %d", 1, user.ID)
	}
}

func TestVerifyPasswordAndGetUser_InvalidPassword(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}

	_, err := userService.VerifyPasswordAndGetUser("testuser", "wrongpassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestVerifyPasswordAndGetUser_HashError(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{PasswordHash: "invalidhashformat"}}

	_, err := userService.VerifyPasswordAndGetUser("testuser", "somepassword")

	if err == nil || strings.Contains(err.Error(), "hash parsing failed") == false {
		t.Fatalf("expected hash parsing error, got %v", err)
	}
}

func TestGetAllUsers_Success(t *testing.T) {
	mockUser := &User{ID: 1, Name: "Test User", Username: "testuser", Role: ServiceRole}
	userService := Service{DB: &MockUserPersistence{MockUser: mockUser}}

	users, err := userService.GetAllUsers()

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
	userService := Service{DB: &MockUserPersistence{ShouldFail: true}}

	_, err := userService.GetAllUsers()

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{}}

	user, err := userService.UpdateUser(1, "Updated User", "updateduser", AdminRole, false)

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
	if user.Locked != false {
		t.Errorf("expected locked false, got %v", user.Locked)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	userService := Service{DB: &MockUserPersistence{ShouldFail: true}}

	user, err := userService.UpdateUser(1, "Updated User", "updateduser", AdminRole, false)

	if err != ErrDatabase {
		t.Fatalf("expected database error, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user on error, got %v", user)
	}
}
