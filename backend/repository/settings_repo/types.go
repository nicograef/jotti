package settings_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{q: dbgen.New(database)}
}

func toDomain(row dbgen.GetBetreiberRow) settings.Betreiber {
	b := settings.Betreiber{
		Vereinsname: row.Vereinsname,
		Strasse:     row.Strasse,
		Plz:         row.Plz,
		Ort:         row.Ort,
	}
	if row.Steuernummer.Valid {
		b.Steuernummer = &row.Steuernummer.String
	}
	if row.UstID.Valid {
		b.UstID = &row.UstID.String
	}
	return b
}
