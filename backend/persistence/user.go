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
	row := p.DB.QueryRow("SELECT id, name, username, role, locked FROM users WHERE id = $1", id)

	var dbUser user.User
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Locked); err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByUsername retrieves a user from the database by their username.
func (p *UserPersistence) GetUserByUsername(username string) (*user.User, error) {
	row := p.DB.QueryRow("SELECT id, name, username, role, locked FROM users WHERE username = $1", username)

	var dbUser user.User
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Locked); err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

func (p *UserPersistence) GetAllUsers() ([]*user.User, error) {
	rows, err := p.DB.Query("SELECT id, name, username, role, locked FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dbUser user.User
		if err := rows.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Locked); err != nil {
			return nil, err
		}
		users = append(users, &dbUser)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser inserts a new user into the database with the given name, username and role.
// Returns an error if the operation fails, and the row id of the newly created user.
func (p *UserPersistence) CreateUser(name, username, onetimePasswordHash string, role user.Role) (int, error) {
	var userID int
	err := p.DB.QueryRow("INSERT INTO users (name, username, onetime_password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", name, username, onetimePasswordHash, role).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (p *UserPersistence) UpdateUser(id int, name, username string, role user.Role, locked bool) error {
	result, err := p.DB.Exec("UPDATE users SET name = $1, username = $2, role = $3, locked = $4 WHERE id = $5", name, username, role, locked, id)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// GetPasswordHash retrieves the password hash for the user with the given username.
func (p *UserPersistence) GetPasswordHash(username string) (string, error) {
	row := p.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username)

	var passwordHash sql.NullString
	if err := row.Scan(&passwordHash); err != nil {
		if err == sql.ErrNoRows {
			return "", user.ErrUserNotFound
		}
		return "", err
	}

	if !passwordHash.Valid {
		return "", nil
	}

	return passwordHash.String, nil
}

// GetOnetimePasswordHash retrieves the one-time password hash for the user with the given username.
func (p *UserPersistence) GetOnetimePasswordHash(username string) (string, error) {
	row := p.DB.QueryRow("SELECT onetime_password_hash FROM users WHERE username = $1", username)

	var passwordHash sql.NullString
	if err := row.Scan(&passwordHash); err != nil {
		if err == sql.ErrNoRows {
			return "", user.ErrUserNotFound
		}
		return "", err
	}

	if !passwordHash.Valid {
		return "", nil
	}

	return passwordHash.String, nil
}

// SetPasswordHash updates the password hash for the user with the given username.
func (p *UserPersistence) SetPasswordHash(username, passwordHash string) error {
	_, err := p.DB.Exec("UPDATE users SET password_hash = $1, onetime_password_hash = NULL WHERE username = $2", passwordHash, username)
	return err
}

// SetOnetimePasswordHash updates the one-time password hash for the user with the given username.
func (p *UserPersistence) SetOnetimePasswordHash(username, onetimePasswordHash string) error {
	_, err := p.DB.Exec("UPDATE users SET onetime_password_hash = $1, password_hash = NULL WHERE username = $2", onetimePasswordHash, username)
	return err
}
