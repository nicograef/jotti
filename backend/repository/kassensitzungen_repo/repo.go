package kassensitzungen_repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
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

// GetAbgeschlosseneKassensitzungen reads only closed Kassensitzungen (status 'abgeschlossen')
// for the Kassenberichte page; the transient 'wird_abgeschlossen' status never appears there.
func (r Repository) GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error) {
	rows, err := r.q.GetAbgeschlosseneKassensitzungen(ctx)
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
		Status:      kasse.KassensitzungStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

// GetAktiveKassensitzung reads the active (not yet closed) Kassensitzung, i.e. one with status
// 'offen' or 'wird_abgeschlossen'. Returns nil if no such Kassensitzung exists. There is at most
// one active Kassensitzung (enforced by idx_kassensitzungen_eine_aktiv).
func (r Repository) GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error) {
	row, err := r.q.GetAktiveKassensitzung(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}

	ks := kassensitzungRowToDomain(row)
	return &ks, nil
}

// SetKassensitzungWirdAbgeschlossen sets the barrier status: it moves the Kassensitzung from
// 'offen' (or keeps it at 'wird_abgeschlossen' for a resumed close) to 'wird_abgeschlossen'.
// The UPDATE waits for in-flight booking transactions holding FOR SHARE. Returns the number of
// affected rows (0 when the Kassensitzung is already closed).
func (r Repository) SetKassensitzungWirdAbgeschlossen(ctx context.Context, zNr int) (int64, error) {
	rows, err := r.q.SetKassensitzungWirdAbgeschlossen(ctx, zNr)
	if err != nil {
		return 0, db.Error(err)
	}
	return rows, nil
}

// SetKassensitzungOffen resets the barrier status back to 'offen' after a failed close (best effort).
// It only affects a Kassensitzung still in 'wird_abgeschlossen'. Returns the number of affected rows.
func (r Repository) SetKassensitzungOffen(ctx context.Context, zNr int) (int64, error) {
	rows, err := r.q.SetKassensitzungOffen(ctx, zNr)
	if err != nil {
		return 0, db.Error(err)
	}
	return rows, nil
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
