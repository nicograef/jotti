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

func tableRowToDomain(row dbgen.Table) table.Table {
	return table.Table{
		ID:        row.ID,
		Name:      row.Name,
		Status:    table.Status(row.Status),
		CreatedAt: row.CreatedAt,
	}
}
