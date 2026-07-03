package user_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

func userRowToDomain(row dbgen.GetUserRow) user.User {
	return user.User{
		ID:                      row.ID,
		Name:                    row.Name,
		Username:                row.Username,
		Role:                    user.Role(row.Role),
		Status:                  user.Status(row.Status),
		PasswordHash:            row.PasswordHash.String,
		OnetimePasswordHash:     row.OnetimePasswordHash.String,
		OnetimePasswordAttempts: int(row.OnetimePasswordAttempts),
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

func userByUsernameRowToDomain(row dbgen.GetUserByUsernameRow) user.User {
	return user.User{
		ID:                      row.ID,
		Name:                    row.Name,
		Username:                row.Username,
		Role:                    user.Role(row.Role),
		Status:                  user.Status(row.Status),
		PasswordHash:            row.PasswordHash.String,
		OnetimePasswordHash:     row.OnetimePasswordHash.String,
		OnetimePasswordAttempts: int(row.OnetimePasswordAttempts),
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

func allUsersRowToDomain(row dbgen.GetAllUsersRow) user.User {
	return user.User{
		ID:        row.ID,
		Name:      row.Name,
		Username:  row.Username,
		Role:      user.Role(row.Role),
		Status:    user.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
