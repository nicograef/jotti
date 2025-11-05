package user

import (
	"errors"
	"log"
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
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
	Locked   bool   `json:"locked"`
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
	GetUserByUsername(username string) (*User, error)
	GetUser(id int) (*User, error)
	GetAllUsers() ([]*User, error)
	CreateUser(name, username, onetimePasswordHash string, role Role) (int, error)
	UpdateUser(id int, name, username string, role Role, locked bool) error
	SetPasswordHash(username, passwordHash string) error
	GetPasswordHash(username string) (string, error)
	GetOnetimePasswordHash(username string) (string, error)
	SetOnetimePasswordHash(username, onetimePasswordHash string) error
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

	userID, err := s.DB.CreateUser(name, username, onetimePasswordHash, role)
	if err != nil {
		log.Printf("ERROR Failed to create user: %v", err)
		return nil, "", ErrDatabase
	}

	return &User{
		ID:       userID,
		Name:     name,
		Username: username,
		Role:     role,
		Locked:   false,
	}, onetimePassword, nil
}

// VerifyPasswordAndGetUser logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) VerifyPasswordAndGetUser(username, password string) (*User, error) {
	passwordHash, err := s.DB.GetPasswordHash(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("ERROR User %s not found during login", username)
			return nil, ErrUserNotFound
		}
		log.Printf("ERROR Failed to retrieve password hash for user %s: %v", username, err)
		return nil, ErrDatabase
	}
	if passwordHash == "" {
		log.Printf("ERROR No password set for user %s", username)
		return nil, ErrNoPassword
	}

	if err := verifyPassword(passwordHash, password); err != nil {
		log.Printf("ERROR Password validation failed for user %s: %v", username, err)
		return nil, ErrInvalidPassword
	}

	user, err := s.DB.GetUserByUsername(username)
	if err != nil {
		log.Printf("ERROR Failed to retrieve user %s after password validation: %v", username, err)
		return nil, ErrDatabase
	}

	return user, nil
}

// SetNewPassword logs in a user by validating the provided one-time password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) SetNewPassword(username, newPassword, onetimePassword string) (*User, error) {
	onetimePasswordHash, err := s.DB.GetOnetimePasswordHash(username)
	if err != nil {
		log.Printf("ERROR Failed to retrieve one-time password hash for user %s: %v", username, err)
		return nil, ErrDatabase
	}
	if onetimePasswordHash == "" {
		log.Printf("ERROR No one-time password set for user %s", username)
		return nil, ErrNoOnetimePassword
	}
	if err := verifyPassword(onetimePasswordHash, onetimePassword); err != nil {
		log.Printf("ERROR One-time password validation failed for user %s: %v", username, err)
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := createArgon2idHash(newPassword)
	if err != nil {
		log.Printf("ERROR Failed to hash password for user %s: %v", username, err)
		return nil, ErrPasswordHashing
	}

	if err := s.DB.SetPasswordHash(username, hashedPassword); err != nil {
		log.Printf("ERROR Failed to set password hash in DB for user %s: %v", username, err)
		return nil, ErrDatabase
	}

	log.Printf("INFO Password set successfully for user %s", username)

	user, err := s.DB.GetUserByUsername(username)
	if err != nil {
		log.Printf("ERROR Failed to retrieve user %s: %v", username, err)
		return nil, ErrDatabase
	}

	return user, nil
}

// ResetPassword resets the password for the user with the given username and returns a new one-time password.
func (s *Service) ResetPassword(username string) (string, error) {
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

	err = s.DB.SetOnetimePasswordHash(username, onetimePasswordHash)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Printf("ERROR User %s not found for password reset", username)
		return "", ErrUserNotFound
	} else if err != nil {
		log.Printf("ERROR Failed to set one-time password hash in DB for user %s: %v", username, err)
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

func (s *Service) GetAllUsers() ([]*User, error) {
	users, err := s.DB.GetAllUsers()
	if err != nil {
		log.Printf("ERROR Failed to retrieve all users: %v", err)
		return nil, ErrDatabase
	}

	return users, nil
}
