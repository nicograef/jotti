package user

import (
	"errors"
	"log"

	"github.com/nicograef/jotti/backend/config"
)

type Role string

const (
	AdminRole   Role = "admin"
	ServiceRole Role = "service"
)

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"-"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")
var ErrPasswordHashing = errors.New("password hashing error")
var ErrDatabase = errors.New("database error")

type Persistence interface {
	GetUserByUsername(username string) (*User, error)
	GetUser(id int) (*User, error)
	CreateUserWithoutPassword(name, username string, role Role) (int64, error)
	SetPasswordHash(userID int, passwordHash string) error
}

type Service struct {
	DB  Persistence
	Cfg config.Config
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
