package auth

import (
	"errors"
	"regexp"
	"time"

	z "github.com/Oudwins/zog"
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
