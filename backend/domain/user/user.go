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
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"-"`
}

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidPassword is returned when a password is invalid.
var ErrInvalidPassword = errors.New("invalid password")

// ErrPasswordHashing is returned when there is an error hashing the password.
var ErrPasswordHashing = errors.New("password hashing error")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

type persistence interface {
	GetUserByUsername(username string) (*User, error)
	GetUser(id int) (*User, error)
	CreateUserWithoutPassword(name, username string, role Role) (int64, error)
	SetPasswordHash(userID int, passwordHash string) error
}

// Service provides user-related operations.
type Service struct {
	DB persistence
}

// CreateUserWithoutPassword creates a new user in the database without setting a password.
func (s *Service) CreateUserWithoutPassword(name, username string, role Role) (*User, error) {
	userID, err := s.DB.CreateUserWithoutPassword(name, username, role)
	if err != nil {
		log.Printf("ERROR Failed to create user: %v", err)
		return nil, ErrDatabase
	}

	return &User{
		ID:       int(userID),
		Name:     name,
		Username: username,
		Role:     role,
	}, nil
}

// LoginUserViaPassword logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) LoginUserViaPassword(username, password string) (*User, error) {
	user, err := s.DB.GetUserByUsername(username)
	if err != nil {
		log.Printf("ERROR Failed to retrieve password hash for user %s: %v", username, err)
		return nil, err
	}

	if user.PasswordHash == "" {
		log.Printf("INFO Setting password for first time login for user %s", username)

		hashedPassword, err := createArgon2idHash(password)
		if err != nil {
			log.Printf("ERROR Failed to hash password for user %s: %v", username, err)
			return nil, err
		}

		if err := s.DB.SetPasswordHash(user.ID, hashedPassword); err != nil {
			log.Printf("ERROR Failed to set password hash in DB for user %s: %v", username, err)
			return nil, err
		}

		user.PasswordHash = hashedPassword
		log.Printf("INFO Password set successfully for user %s", username)
	}

	if err := verifyPassword(user.PasswordHash, password); err != nil {
		log.Printf("ERROR Password validation failed for user %s: %v", username, err)
		return nil, err
	}

	return user, nil
}
