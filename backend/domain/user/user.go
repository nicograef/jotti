package user

import (
	"errors"
	"log"
	"strings"
	"time"
)

// Role represents the role of a user.
type Role string

const (
	// AdminRole represents the admin role. Users with this role have elevated privileges.
	AdminRole Role = "admin"
	// ServiceRole represents the service role. Users with this role have limited privileges.
	ServiceRole Role = "service"
)

// User represents a user in the system.
type User struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Username            string    `json:"username"`
	Role                Role      `json:"role"`
	Locked              bool      `json:"locked"`
	PasswordHash        string    `json:"-"`
	OnetimePasswordHash string    `json:"-"`
	CreatedAt           time.Time `json:"createdAt"`
}

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidPassword is returned when a password is invalid.
var ErrInvalidPassword = errors.New("invalid password")

// ErrNoPassword is returned when there is no password set for the user.
var ErrNoPassword = errors.New("no password set")

// ErrNoOnetimePassword is returned when there is no one-time password set for the user.
var ErrNoOnetimePassword = errors.New("no onetime password set")

// ErrPasswordHashing is returned when there is an error hashing the password.
var ErrPasswordHashing = errors.New("password hashing error")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

type persistence interface {
	GetUserID(username string) (int, error)
	GetUser(id int) (*User, error)
	GetAllUsers() ([]*User, error)
	CreateUser(name, username, onetimePasswordHash string, role Role) (int, error)
	UpdateUser(id int, name, username string, role Role, locked bool) error
	SetPasswordHash(id int, passwordHash string) error
	SetOnetimePasswordHash(id int, onetimePasswordHash string) error
}

// Service provides user-related operations.
type Service struct {
	DB persistence
}

// CreateUser creates a new user in the database without setting a password.
func (s *Service) CreateUser(name, username string, role Role) (*User, string, error) {
	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Printf("ERROR Failed to create one-time password: %v", err)
		return nil, "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Printf("ERROR Failed to hash one-time password: %v", err)
		return nil, "", ErrPasswordHashing
	}

	// usernames are always lowercase in jotti
	lowerCaseUsername := strings.ToLower(username)

	userID, err := s.DB.CreateUser(name, lowerCaseUsername, onetimePasswordHash, role)
	if err != nil {
		log.Printf("ERROR Failed to create user: %v", err)
		return nil, "", ErrDatabase
	}

	return &User{
		ID:        userID,
		Name:      name,
		Username:  lowerCaseUsername,
		Role:      role,
		Locked:    false,
		CreatedAt: time.Now(),
	}, onetimePassword, nil
}

// VerifyPasswordAndGetUser logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) VerifyPasswordAndGetUser(username, password string) (*User, error) {
	userID, err := s.DB.GetUserID(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("ERROR User %s not found during login", username)
			return nil, ErrUserNotFound
		}
		log.Printf("ERROR Failed to retrieve user %s: %v", username, err)
		return nil, ErrDatabase
	}

	user, err := s.DB.GetUser(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("ERROR User %s not found during login", username)
			return nil, ErrUserNotFound
		}
		log.Printf("ERROR Failed to retrieve password hash for user %s: %v", username, err)
		return nil, ErrDatabase
	}
	if user.PasswordHash == "" {
		log.Printf("ERROR No password set for user %s", username)
		return nil, ErrNoPassword
	}

	if err := verifyPassword(user.PasswordHash, password); err != nil {
		log.Printf("ERROR Password validation failed for user %s: %v", username, err)
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// SetNewPassword logs in a user by validating the provided one-time password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) SetNewPassword(username, newPassword, onetimePassword string) (*User, error) {
	userID, err := s.DB.GetUserID(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("ERROR User %s not found during password validation", username)
			return nil, ErrUserNotFound
		}
		log.Printf("ERROR Failed to retrieve user %s: %v", username, err)
		return nil, ErrDatabase
	}

	user, err := s.DB.GetUser(userID)
	if err != nil {
		log.Printf("ERROR Failed to retrieve one-time password hash for user %s: %v", username, err)
		return nil, ErrDatabase
	}
	if user.OnetimePasswordHash == "" {
		log.Printf("ERROR No one-time password set for user %s", username)
		return nil, ErrNoOnetimePassword
	}
	if err := verifyPassword(user.OnetimePasswordHash, onetimePassword); err != nil {
		log.Printf("ERROR One-time password validation failed for user %s: %v", username, err)
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := createArgon2idHash(newPassword)
	if err != nil {
		log.Printf("ERROR Failed to hash password for user %s: %v", username, err)
		return nil, ErrPasswordHashing
	}

	if err := s.DB.SetPasswordHash(user.ID, hashedPassword); err != nil {
		log.Printf("ERROR Failed to set password hash in DB for user %s: %v", username, err)
		return nil, ErrDatabase
	}

	log.Printf("INFO Password set successfully for user %s", username)

	return user, nil
}

// ResetPassword resets the password for the user with the given user ID and returns a new one-time password.
func (s *Service) ResetPassword(userID int) (string, error) {
	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Printf("ERROR Failed to create one-time password: %v", err)
		return "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Printf("ERROR Failed to hash one-time password: %v", err)
		return "", ErrPasswordHashing
	}

	err = s.DB.SetOnetimePasswordHash(userID, onetimePasswordHash)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Printf("ERROR User %d not found for password reset", userID)
		return "", ErrUserNotFound
	} else if err != nil {
		log.Printf("ERROR Failed to set one-time password hash in DB for user %d: %v", userID, err)
		return "", ErrDatabase
	}

	return onetimePassword, nil
}

// UpdateUser updates the user's details in the database.
func (s *Service) UpdateUser(id int, name, username string, role Role, locked bool) (*User, error) {
	err := s.DB.UpdateUser(id, name, username, role, locked)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Printf("ERROR User %d not found for update", id)
		return nil, ErrUserNotFound
	} else if err != nil {
		log.Printf("ERROR Failed to update user %d: %v", id, err)
		return nil, ErrDatabase
	}

	updatedUser, err := s.DB.GetUser(id)
	if err != nil {
		log.Printf("ERROR Failed to retrieve updated user %d: %v", id, err)
		return nil, ErrDatabase
	}

	return updatedUser, nil
}

// GetAllUsers retrieves all users from the database.
func (s *Service) GetAllUsers() ([]*User, error) {
	users, err := s.DB.GetAllUsers()
	if err != nil {
		log.Printf("ERROR Failed to retrieve all users: %v", err)
		return nil, ErrDatabase
	}

	return users, nil
}
