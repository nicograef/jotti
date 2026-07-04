package betreiber_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{q: dbgen.New(database)}
}

// GetBetreiber returns the betreiber data, or db.ErrNotFound if not yet set.
func (r Repository) GetBetreiber(ctx context.Context) (betreiber.Betreiber, error) {
	row, err := r.q.GetBetreiber(ctx)
	if err != nil {
		return betreiber.Betreiber{}, db.Error(err)
	}
	return toDomain(row), nil
}

// UpsertBetreiber creates or updates the betreiber data.
func (r Repository) UpsertBetreiber(ctx context.Context, b betreiber.Betreiber) error {
	params := dbgen.UpsertBetreiberParams{
		Vereinsname: b.Vereinsname,
		Strasse:     b.Strasse,
		Plz:         b.Plz,
		Ort:         b.Ort,
	}
	if b.Steuernummer != nil {
		params.Steuernummer = sql.NullString{String: *b.Steuernummer, Valid: true}
	}
	if b.UstID != nil {
		params.UstID = sql.NullString{String: *b.UstID, Valid: true}
	}
	if err := r.q.UpsertBetreiber(ctx, params); err != nil {
		return db.Error(err)
	}
	return nil
}

func toDomain(row dbgen.GetBetreiberRow) betreiber.Betreiber {
	b := betreiber.Betreiber{
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
