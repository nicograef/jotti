package settings_repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// GetBetreiber returns the betreiber data, or nil + db.ErrNotFound if not yet set.
func (r Repository) GetBetreiber(ctx context.Context) (settings.Betreiber, error) {
	row, err := r.q.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings.Betreiber{}, db.ErrNotFound
		}
		return settings.Betreiber{}, db.ErrDatabase
	}
	return toDomain(row), nil
}

// UpsertBetreiber creates or updates the betreiber data.
func (r Repository) UpsertBetreiber(ctx context.Context, b settings.Betreiber) error {
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
		return db.ErrDatabase
	}
	return nil
}

// GetKassenidentitaet returns the kasse identity, or db.ErrNotFound if not yet initialized.
func (r Repository) GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error) {
	row, err := r.q.GetKassenidentitaet(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings.Kassenidentitaet{}, db.ErrNotFound
		}
		return settings.Kassenidentitaet{}, db.ErrDatabase
	}
	return settings.Kassenidentitaet{
		Seriennummer: row.Seriennummer,
		AngelegtAm:   row.AngelegtAm,
	}, nil
}
