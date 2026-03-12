package table_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements table persistence layer using sqlc-generated queries.
type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{DB: db, q: dbgen.New(db)}
}

func tischRowToDomain(row dbgen.Table) table.Tisch {
	return table.Tisch{
		ID:        row.ID,
		Name:      row.Name,
		Status:    table.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
