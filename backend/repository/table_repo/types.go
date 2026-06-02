package table_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements table persistence layer using sqlc-generated queries.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

func tischRowToDomain(row dbgen.Tische) table.Tisch {
	return table.Tisch{
		ID:        row.ID,
		Name:      row.Name,
		Status:    table.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
