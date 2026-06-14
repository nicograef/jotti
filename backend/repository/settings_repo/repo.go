package settings_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// GetBetreiber returns the betreiber data, or db.ErrNotFound if not yet set.
func (r Repository) GetBetreiber(ctx context.Context) (settings.Betreiber, error) {
	row, err := r.q.GetBetreiber(ctx)
	if err != nil {
		return settings.Betreiber{}, db.Error(err)
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
		return db.Error(err)
	}
	return nil
}

// GetKassenidentitaet returns the kasse identity, or db.ErrNotFound if not yet initialized.
func (r Repository) GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error) {
	row, err := r.q.GetKassenidentitaet(ctx)
	if err != nil {
		return settings.Kassenidentitaet{}, db.Error(err)
	}
	return settings.Kassenidentitaet{
		Seriennummer: row.Seriennummer,
		AngelegtAm:   row.AngelegtAm,
	}, nil
}

func (r Repository) GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error) {
	row, err := r.q.GetTSEKonfiguration(ctx)
	if err != nil {
		return settings.TSEKonfiguration{}, db.Error(err)
	}
	return toTSEKonfiguration(row), nil
}

func (r Repository) UpsertTSEKonfiguration(ctx context.Context, c settings.TSEKonfiguration) error {
	err := r.q.UpsertTSEKonfiguration(ctx, dbgen.UpsertTSEKonfigurationParams{
		ApiKey:    c.ApiKey,
		ApiSecret: c.ApiSecret,
		TssID:     c.TssID,
		ClientID:  c.ClientID,
	})
	if err != nil {
		return db.Error(err)
	}
	return nil
}
