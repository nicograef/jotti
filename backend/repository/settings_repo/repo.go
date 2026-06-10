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

func (r Repository) GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error) {
	row, err := r.q.GetBondruckEinstellungen(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings.BondruckEinstellungen{}, db.ErrNotFound
		}
		return settings.BondruckEinstellungen{}, db.ErrDatabase
	}
	return toBondruckEinstellungen(row), nil
}

func (r Repository) UpsertBondruckEinstellungen(ctx context.Context, b settings.BondruckEinstellungen) error {
	err := r.q.UpsertBondruckEinstellungen(ctx, dbgen.UpsertBondruckEinstellungenParams{
		KassenbelegDruckerIp: b.KassenbelegDruckerIP,
		DirektverkaufModus:   string(b.DirektverkaufModus),
		AbholbonDruckerIp:    b.AbholbonDruckerIP,
	})
	if err != nil {
		return db.ErrDatabase
	}
	return nil
}

func (r Repository) GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error) {
	row, err := r.q.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings.TSEKonfiguration{}, db.ErrNotFound
		}
		return settings.TSEKonfiguration{}, db.ErrDatabase
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
		return db.ErrDatabase
	}
	return nil
}
