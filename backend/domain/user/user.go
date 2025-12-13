package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/jwt"
)

type Role string

const (
	// AdminRole: can do everything.
	AdminRole Role = "admin"
	// ServiceRole: can only see active tables and products and .
	ServiceRole Role = "service"
)

type Status string

const (
	// ActiveStatus: user can authenticate and use the system.
	ActiveStatus Status = "active"
	// InactiveStatus: user is disabled and cannot authenticate.
	InactiveStatus Status = "inactive"
)

type User struct {
	ID                  int
	Name                string
	Username            string
	Role                Role
	Status              Status
	PasswordHash        string
	OnetimePasswordHash string
	CreatedAt           time.Time
}

var IDSchema = z.Int().GTE(1, z.Message("Invalid user ID"))

var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(50, z.Message("Name too long"))

var UsernameSchema = z.String().Trim().Min(3, z.Message("Username too short")).Max(20, z.Message("Username too long")).Match(
	regexp.MustCompile(`^[a-z0-9]+$`),
	z.Message("Only lowercase alphanumerical usernames allowed"),
)

var RoleSchema = z.StringLike[Role]().OneOf(
	[]Role{AdminRole, ServiceRole},
	z.Message("Invalid role"),
)

var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus},
	z.Message("Invalid status"),
)

var UserSchema = z.Struct(z.Shape{
	"ID":                  IDSchema.Required(),
	"Name":                NameSchema.Required(),
	"Username":            UsernameSchema.Required(),
	"Role":                RoleSchema.Required(),
	"Status":              StatusSchema.Required(),
	"PasswordHash":        z.String(),
	"OnetimePasswordHash": z.String(),
	"CreatedAt":           z.Time().Required(),
})

var ErrNotActive = fmt.Errorf("user is not active")

func (u User) Validate() error {
	if errsMap := UserSchema.Validate(&u); errsMap != nil {
		issues := z.Issues.SanitizeMapAndCollect(errsMap)
		return fmt.Errorf("invalid user: %v", issues)
	}
	return nil
}

func NewUser(name, username string, role Role) (User, string, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return User{}, "", fmt.Errorf("invalid name")
	}

	if issue := UsernameSchema.Validate(&username); issue != nil {
		return User{}, "", fmt.Errorf("invalid username")
	}

	if issue := RoleSchema.Validate(&role); issue != nil {
		return User{}, "", fmt.Errorf("invalid role")
	}

	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		return User{}, "", fmt.Errorf("failed to generate one-time password: %w", err)
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to hash one-time password: %w", err)
	}

	user := User{
		Name:                name,
		Username:            strings.ToLower(username),
		Role:                role,
		Status:              InactiveStatus,
		PasswordHash:        "",
		OnetimePasswordHash: onetimePasswordHash,
		CreatedAt:           time.Now().UTC(),
	}

	return user, onetimePassword, nil
}

func (u *User) Activate() {
	u.Status = ActiveStatus
}

func (u *User) Deactivate() {
	u.Status = InactiveStatus
}

func (u *User) UpdateDetails(name, username string, role Role) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := UsernameSchema.Validate(&username); issue != nil {
		return fmt.Errorf("invalid username")
	}

	if issue := RoleSchema.Validate(&role); issue != nil {
		return fmt.Errorf("invalid role")
	}

	u.Name = name
	u.Username = username
	u.Role = role

	return nil
}

func (u *User) ResetPassword() (string, error) {
	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		return "", fmt.Errorf("failed to generate one-time password: %w", err)
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		return "", fmt.Errorf("failed to hash one-time password: %w", err)
	}

	u.OnetimePasswordHash = onetimePasswordHash
	u.PasswordHash = ""

	return onetimePassword, nil
}

func (u *User) SetPassword(onetimePassword, newPassword string) error {
	if u.OnetimePasswordHash == "" {
		return ErrNoPassword
	}

	if err := verifyPassword(u.OnetimePasswordHash, onetimePassword); err != nil {
		return err
	}

	passwordHash, err := createArgon2idHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	u.PasswordHash = passwordHash
	u.OnetimePasswordHash = ""

	return nil
}

func (u *User) GenerateJWTToken(password, secret string) (string, error) {
	if u.Status != ActiveStatus {
		return "", ErrNotActive
	}

	if u.PasswordHash == "" {
		return "", ErrNoPassword
	}

	if err := verifyPassword(u.PasswordHash, password); err != nil {
		return "", err
	}

	return jwt.GenerateJWTTokenForUser(u.ID, string(u.Role), secret)
}
