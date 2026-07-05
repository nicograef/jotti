package tisch_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements table persistence layer using sqlc-generated queries.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

func tischRowToDomain(row dbgen.Tische) tisch.Tisch {
	return tisch.Tisch{
		ID:        row.ID,
		Name:      row.Name,
		Status:    tisch.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
