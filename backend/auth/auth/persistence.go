package auth

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
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
		log.Error().Err(err).Int("user_id", id).Msg("DB Error scanning user row")
		return nil, db.Error(err)
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
func (p *Persistence) GetUserID(ctx context.Context, username string) (int, error) {
	log := zerolog.Ctx(ctx)

	row := p.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1 AND status != 'deleted'", username)

	var userID int
	if err := row.Scan(&userID); err != nil {
		log.Error().Err(err).Str("username", username).Msg("DB Error scanning user ID row")
		return 0, db.Error(err)
	}

	return userID, nil
}

// SetPasswordHash updates the password hash for the user with the given user ID.
func (p *Persistence) SetPasswordHash(ctx context.Context, id int, passwordHash string) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET password_hash = $1, onetime_password_hash = NULL WHERE id = $2", passwordHash, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("DB Error setting password hash")
		return db.Error(err)
	}

	return db.ResultError(result)
}

// SetOnetimePasswordHash updates the one-time password hash for the user with the given user ID.
func (p *Persistence) SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE users SET onetime_password_hash = $1, password_hash = NULL WHERE id = $2", onetimePasswordHash, id)
	if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("DB Error setting onetime password hash")
		return db.Error(err)
	}

	return db.ResultError(result)
}
