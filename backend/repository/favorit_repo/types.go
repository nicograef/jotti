package favorit_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements favorit persistence layer using sqlc-generated queries.
type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{DB: db, q: dbgen.New(db)}
}
