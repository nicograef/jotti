package persistence

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/user"
)

// UserPersistence implements user persistence layer using a SQL database.
type UserPersistence struct {
	DB *sql.DB
}

// GetUser retrieves a user from the database by their ID.
func (p *UserPersistence) GetUser(id int) (*user.User, error) {
	row := p.DB.QueryRow("SELECT id, name, username, role, password_hash FROM users WHERE id = $1", id)

	var dbUser user.User
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByUsername retrieves a user from the database by their username.
func (p *UserPersistence) GetUserByUsername(username string) (*user.User, error) {
	row := p.DB.QueryRow("SELECT id, name, username, role, password_hash FROM users WHERE username = $1", username)

	var dbUser user.User
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// CreateUserWithoutPassword inserts a new user into the database with the given name, username and role.
// Returns an error if the operation fails, and the row id of the newly created user.
func (p *UserPersistence) CreateUserWithoutPassword(name, username string, role user.Role) (int, error) {
	var userID int
	err := p.DB.QueryRow("INSERT INTO users (name, username, role) VALUES ($1, $2, $3) RETURNING id", name, username, role).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// SetPasswordHash updates the password hash for the user with the given ID.
func (p *UserPersistence) SetPasswordHash(userID int, passwordHash string) error {
	_, err := p.DB.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", passwordHash, userID)
	return err
}
