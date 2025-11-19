package user

import (
	"database/sql"
)

// Persistence implements user persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

type dbuser struct {
	ID                  int            `db:"id"`
	Name                string         `db:"name"`
	Username            string         `db:"username"`
	Role                string         `db:"role"`
	Status              string         `db:"status"`
	PasswordHash        sql.NullString `db:"password_hash"`
	OnetimePasswordHash sql.NullString `db:"onetime_password_hash"`
	CreatedAt           sql.NullTime   `db:"created_at"`
}

// GetUser retrieves a user from the database by their ID.
func (p *Persistence) GetUser(id int) (*User, error) {
	row := p.DB.QueryRow("SELECT id, name, username, role, status, password_hash, onetime_password_hash, created_at FROM users WHERE id = $1 AND status != 'deleted'", id)

	var dbUser dbuser
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Status, &dbUser.PasswordHash, &dbUser.OnetimePasswordHash, &dbUser.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &User{
		ID:                  dbUser.ID,
		Name:                dbUser.Name,
		Username:            dbUser.Username,
		Role:                Role(dbUser.Role),
		Status:              Status(dbUser.Status),
		PasswordHash:        dbUser.PasswordHash.String,
		OnetimePasswordHash: dbUser.OnetimePasswordHash.String,
		CreatedAt:           dbUser.CreatedAt.Time,
	}, nil
}

// GetUserID retrieves a user id from the database by their username.
func (p *Persistence) GetUserID(username string) (int, error) {
	row := p.DB.QueryRow("SELECT id FROM users WHERE username = $1 AND status != 'deleted'", username)

	var userID int
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	return userID, nil
}

// GetAllUsers retrieves all users from the database.
func (p *Persistence) GetAllUsers() ([]*User, error) {
	rows, err := p.DB.Query("SELECT id, name, username, role, status, created_at FROM users WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var users []*User
	for rows.Next() {
		var dbUser dbuser
		if err := rows.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Status, &dbUser.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &User{
			ID:        dbUser.ID,
			Name:      dbUser.Name,
			Username:  dbUser.Username,
			Role:      Role(dbUser.Role),
			Status:    Status(dbUser.Status),
			CreatedAt: dbUser.CreatedAt.Time,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser inserts a new user into the database with the given name, username and role.
// Returns an error if the operation fails, and the row id of the newly created
func (p *Persistence) CreateUser(name, username, onetimePasswordHash string, role Role) (int, error) {
	var userID int
	err := p.DB.QueryRow("INSERT INTO users (name, username, onetime_password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", name, username, onetimePasswordHash, role).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// UpdateUser updates the user's information in the database.
func (p *Persistence) UpdateUser(id int, name, username string, role Role) error {
	result, err := p.DB.Exec("UPDATE users SET name = $1, username = $2, role = $3 WHERE id = $4", name, username, role, id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ActivateUser sets the status of the user with the given user ID to 'active'.
func (p *Persistence) ActivateUser(id int) error {
	result, err := p.DB.Exec("UPDATE users SET status = 'active' WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeactivateUser sets the status of the user with the given user ID to 'inactive'.
func (p *Persistence) DeactivateUser(id int) error {
	result, err := p.DB.Exec("UPDATE users SET status = 'inactive' WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetPasswordHash updates the password hash for the user with the given user ID.
func (p *Persistence) SetPasswordHash(id int, passwordHash string) error {
	result, err := p.DB.Exec("UPDATE users SET password_hash = $1, onetime_password_hash = NULL WHERE id = $2", passwordHash, id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetOnetimePasswordHash updates the one-time password hash for the user with the given user ID.
func (p *Persistence) SetOnetimePasswordHash(id int, onetimePasswordHash string) error {
	result, err := p.DB.Exec("UPDATE users SET onetime_password_hash = $1, password_hash = NULL WHERE id = $2", onetimePasswordHash, id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
