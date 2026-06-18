package kassensitzungen_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Repository implements kassensitzungen persistence layer using sqlc-generated queries.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

func kassensitzungRowToDomain(row dbgen.Kassensitzungen) kasse.Kassensitzung {
	return kasse.Kassensitzung{
		ZNr:         row.ZNr,
		Datum:       row.Datum,
		Bezeichnung: row.Bezeichnung,
		Status:      kasse.KassensitzungStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
