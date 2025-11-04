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
	Locked       bool   `json:"locked"`
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
	GetAllUsers() ([]*User, error)
	CreateUserWithoutPassword(name, username string, role Role) (int, error)
	SetPasswordHash(userID int, passwordHash string) error
	UpdateUser(id int, name, username string, role Role, locked bool) error
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
		ID:       userID,
		Name:     name,
		Username: username,
		Role:     role,
		Locked:   false,
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
