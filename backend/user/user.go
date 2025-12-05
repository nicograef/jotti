package user

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog"
)

// Role represents the role of a user.
type Role string

const (
	// AdminRole represents the admin role. Users with this role have elevated privileges.
	AdminRole Role = "admin"
	// ServiceRole represents the service role. Users with this role have limited privileges.
	ServiceRole Role = "service"
)

// Status represents the status of a user.
type Status string

const (
	// ActiveStatus indicates the user can authenticate and use the system.
	ActiveStatus Status = "active"
	// InactiveStatus indicates the user is disabled and cannot authenticate.
	InactiveStatus Status = "inactive"
	// DeletedStatus indicates the user is deleted/archived.
	DeletedStatus Status = "deleted"
)

// User represents a user in the system.
type User struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Username            string    `json:"username"`
	Role                Role      `json:"role"`
	Status              Status    `json:"status"`
	PasswordHash        string    `json:"-"`
	OnetimePasswordHash string    `json:"-"`
	CreatedAt           time.Time `json:"createdAt"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid user ID"))

// NameSchema defines the schema for a user's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(50, z.Message("Name too long"))

// UsernameSchema defines the schema for a username.
var UsernameSchema = z.String().Trim().Min(3, z.Message("Username too short")).Max(20, z.Message("Username too long")).Match(
	regexp.MustCompile(`^[a-z0-9]+$`),
	z.Message("Only lowercase alphanumerical usernames allowed"),
)

// RoleSchema defines the schema for a user role.
var RoleSchema = z.StringLike[Role]().OneOf(
	[]Role{AdminRole, ServiceRole},
	z.Message("Invalid role"),
)

// StatusSchema defines the schema for a user status.
var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Invalid status"),
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrUserNotFound is returned when a user is not found.
var ErrUsernameAlreadyExists = errors.New("username already exists")

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
	GetUserID(ctx context.Context, username string) (int, error)
	GetUser(ctx context.Context, id int) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error)
	UpdateUser(ctx context.Context, id int, name, username string, role Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	SetPasswordHash(ctx context.Context, id int, passwordHash string) error
	SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error
}

// Service provides user-related operations.
type Service struct {
	Persistence persistence
}

// CreateUser creates a new user in the database without setting a password.
func (s *Service) CreateUser(ctx context.Context, name, username string, role Role) (*User, string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return nil, "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return nil, "", ErrPasswordHashing
	}

	// usernames are always lowercase in jotti
	lowerCaseUsername := strings.ToLower(username)

	id, err := s.Persistence.CreateUser(ctx, name, lowerCaseUsername, onetimePasswordHash, role)
	if err != nil {
		log.Error().Err(err).Str("username", lowerCaseUsername).Msg("Failed to create user")
		return nil, "", ErrDatabase
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to retrieve user after creation")
		return nil, "", ErrDatabase
	}

	return user, onetimePassword, nil
}

// VerifyPasswordAndGetUser logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*User, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve user")
		return nil, ErrDatabase
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve password hash")
		return nil, ErrDatabase
	}
	if user.PasswordHash == "" {
		log.Warn().Str("username", username).Msg("No password set for user")
		return nil, ErrNoPassword
	}

	if err := verifyPassword(user.PasswordHash, password); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("Password validation failed")
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// SetNewPassword logs in a user by validating the provided one-time password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Service) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) (*User, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during password validation")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve user")
		return nil, ErrDatabase
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve one-time password hash")
		return nil, ErrDatabase
	}
	if user.OnetimePasswordHash == "" {
		log.Warn().Str("username", username).Msg("No one-time password set for user")
		return nil, ErrNoOnetimePassword
	}
	if err := verifyPassword(user.OnetimePasswordHash, onetimePassword); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("One-time password validation failed")
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := createArgon2idHash(newPassword)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to hash password")
		return nil, ErrPasswordHashing
	}

	if err := s.Persistence.SetPasswordHash(ctx, user.ID, hashedPassword); err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to set password hash in persistence")
		return nil, ErrDatabase
	}

	log.Info().Str("username", username).Msg("Password set successfully")

	return user, nil
}

// ResetPassword resets the password for the user with the given user ID and returns a new one-time password.
func (s *Service) ResetPassword(ctx context.Context, userID int) (string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return "", ErrPasswordHashing
	}

	err = s.Persistence.SetOnetimePasswordHash(ctx, userID, onetimePasswordHash)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", userID).Msg("User not found for password reset")
		return "", ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to set one-time password hash in persistence")
		return "", ErrDatabase
	}

	return onetimePassword, nil
}

// UpdateUser updates the user's details in the database.
func (s *Service) UpdateUser(ctx context.Context, id int, name, username string, role Role) (*User, error) {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.UpdateUser(ctx, id, name, username, role)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for update")
		return nil, ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to update user")
		return nil, ErrDatabase
	}

	updatedUser, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to retrieve updated user")
		return nil, ErrDatabase
	}

	return updatedUser, nil
}

// GetAllUsers retrieves all users from the database.
func (s *Service) GetAllUsers(ctx context.Context) ([]*User, error) {
	log := zerolog.Ctx(ctx)

	users, err := s.Persistence.GetAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all users")
		return nil, ErrDatabase
	}

	return users, nil
}

// ActivateUser sets the status of the user with the given user ID to 'active'.
func (s *Service) ActivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.ActivateUser(ctx, id)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for activation")
		return ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to activate user")
		return ErrDatabase
	}

	return nil
}

// DeactivateUser sets the status of the user with the given user ID to 'inactive'.
func (s *Service) DeactivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.DeactivateUser(ctx, id)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for deactivation")
		return ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to deactivate user")
		return ErrDatabase
	}

	return nil
}
