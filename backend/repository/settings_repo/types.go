package settings_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

func toDomain(row dbgen.GetBetreiberRow) settings.Betreiber {
	b := settings.Betreiber{
		Vereinsname: row.Vereinsname,
		Strasse:     row.Strasse,
		Plz:         row.Plz,
		Ort:         row.Ort,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.Steuernummer.Valid {
		b.Steuernummer = &row.Steuernummer.String
	}
	if row.UstID.Valid {
		b.UstID = &row.UstID.String
	}
	return b
}

func toTSEKonfiguration(row dbgen.GetTSEKonfigurationRow) settings.TSEKonfiguration {
	return settings.TSEKonfiguration{
		ApiKey:    row.ApiKey,
		ApiSecret: row.ApiSecret,
		TssID:     row.TssID,
		ClientID:  row.ClientID,
		UpdatedAt: row.UpdatedAt,
	}
}
