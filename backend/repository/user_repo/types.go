package user_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/user"
)

type Repository struct {
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

func (dp *dbuser) toDomain() user.User {
	return user.User{
		ID:                  dp.ID,
		Name:                dp.Name,
		Username:            dp.Username,
		Role:                user.Role(dp.Role),
		Status:              user.Status(dp.Status),
		PasswordHash:        dp.PasswordHash.String,
		OnetimePasswordHash: dp.OnetimePasswordHash.String,
		CreatedAt:           dp.CreatedAt.Time,
	}
}
