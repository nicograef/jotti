package favorit_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements favorit persistence layer using sqlc-generated queries.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}
