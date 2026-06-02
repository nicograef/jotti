package kassensitzungen_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error) {
	rows, err := r.q.GetAllKassensitzungen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	kassensitzungen := make([]kasse.Kassensitzung, 0, len(rows))
	for _, row := range rows {
		kassensitzungen = append(kassensitzungen, kassensitzungRowToDomain(row))
	}

	return kassensitzungen, nil
}

// GetOffeneKassensitzung reads the currently open Kassensitzung from the kassensitzungen CRUD entity.
// Returns nil if no open Kassensitzung exists.
func (r Repository) GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error) {
	row, err := r.q.GetOffeneKassensitzung(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}

	return &kasse.Kassensitzung{
		ZNr:         row.ZNr,
		Datum:       row.Datum,
		Bezeichnung: row.Bezeichnung,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

// GetOffeneKassensitzungNr returns the z_nr of the currently open Kassensitzung, or 0 if none exists.
func (r Repository) GetOffeneKassensitzungNr(ctx context.Context) (int, error) {
	ks, err := r.GetOffeneKassensitzung(ctx)
	if err != nil {
		return 0, err
	}
	if ks == nil {
		return 0, nil
	}
	return ks.ZNr, nil
}

// InsertKassensitzung creates a new Kassensitzung CRUD entity with status 'offen'.
// Returns the generated z_nr.
func (r Repository) InsertKassensitzung(ctx context.Context, datum time.Time, bezeichnung string) (int, error) {
	zNr, err := r.q.InsertKassensitzung(ctx, dbgen.InsertKassensitzungParams{
		Datum:       datum,
		Bezeichnung: bezeichnung,
		Status:      kasse.KassensitzungOffen,
	})
	if err != nil {
		return 0, db.Error(err)
	}
	return zNr, nil
}
