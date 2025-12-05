package user

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"
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
func (p *Persistence) GetUser(ctx context.Context, id int) (*User, error) {
	log := zerolog.Ctx(ctx)

	row := p.DB.QueryRowContext(ctx, "SELECT id, name, username, role, status, password_hash, onetime_password_hash, created_at FROM users WHERE id = $1 AND status != 'deleted'", id)

	var dbUser dbuser
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Username, &dbUser.Role, &dbUser.Status, &dbUser.PasswordHash, &dbUser.OnetimePasswordHash, &dbUser.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Int("user_id", id).Msg("User not found")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Int("user_id", id).Msg("Failed to scan user")
		return nil, err
	}

	log.Debug().Int("user_id", id).Msg("User retrieved")
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
func (p *Persistence) GetUserID(ctx context.Context, username string) (int, error) {
	log := zerolog.Ctx(ctx)

	row := p.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1 AND status != 'deleted'", username)

	var userID int
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Str("username", username).Msg("User not found by username")
			return 0, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to get user ID")
		return 0, err
	}

	log.Debug().Str("username", username).Int("user_id", userID).Msg("User ID retrieved by username")
	return userID, nil
}

// GetAllUsers retrieves all users from the database.
func (p *Persistence) GetAllUsers(ctx context.Context) ([]*User, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, username, role, status, created_at FROM users WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		log.Error().Err(err).Msg("Failed to query all users")
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
		log.Error().Err(err).Msg("Error iterating over user rows")
		return nil, err
	}

	log.Debug().Int("count", len(users)).Msg("Retrieved all users")
	return users, nil
}

// CreateUser inserts a new user into the database with the given name, username and role.
// Returns an error if the operation fails, and the row id of the newly created
func (p *Persistence) CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error) {
	log := zerolog.Ctx(ctx)

	var userID int
	err := p.DB.QueryRowContext(ctx, "INSERT INTO users (name, username, onetime_password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", name, username, onetimePasswordHash, role).Scan(&userID)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to create user")
		return 0, err
	}

	log.Info().Int("user_id", userID).Str("username", username).Msg("User created")
	return userID, nil
}

// UpdateUser updates the user's information in the database.
func (p *Persistence) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET name = $1, username = $2, role = $3 WHERE id = $4", name, username, role, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to update user")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Int("user_id", id).Msg("User not found for update")
		return ErrUserNotFound
	}

	log.Info().Int("user_id", id).Msg("User updated")
	return nil
}

// ActivateUser sets the status of the user with the given user ID to 'active'.
func (p *Persistence) ActivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET status = 'active' WHERE id = $1", id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to activate user")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Int("user_id", id).Msg("User not found for activation")
		return ErrUserNotFound
	}

	log.Info().Int("user_id", id).Msg("User activated")
	return nil
}

// DeactivateUser sets the status of the user with the given user ID to 'inactive'.
func (p *Persistence) DeactivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET status = 'inactive' WHERE id = $1", id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to deactivate user")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Int("user_id", id).Msg("User not found for deactivation")
		return ErrUserNotFound
	}

	log.Info().Int("user_id", id).Msg("User deactivated")
	return nil
}

// SetPasswordHash updates the password hash for the user with the given user ID.
func (p *Persistence) SetPasswordHash(ctx context.Context, id int, passwordHash string) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET password_hash = $1, onetime_password_hash = NULL WHERE id = $2", passwordHash, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to set password hash")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Int("user_id", id).Msg("User not found for setting password hash")
		return ErrUserNotFound
	}

	log.Info().Int("user_id", id).Msg("Password hash set")
	return nil
}

// SetOnetimePasswordHash updates the one-time password hash for the user with the given user ID.
func (p *Persistence) SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET onetime_password_hash = $1, password_hash = NULL WHERE id = $2", onetimePasswordHash, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to set onetime password hash")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Int("user_id", id).Msg("User not found for setting onetime password hash")
		return ErrUserNotFound
	}

	log.Info().Int("user_id", id).Msg("Onetime password hash set")
	return nil
}
