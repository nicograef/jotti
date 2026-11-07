package tisch_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetTisch(ctx context.Context, id int) (tisch.Tisch, error) {
	row, err := r.q.GetTisch(ctx, id)
	if err != nil {
		return tisch.Tisch{}, db.Error(err)
	}

	return tischRowToDomain(row), nil
}

func (r Repository) GetAlleTische(ctx context.Context) ([]tisch.Tisch, error) {
	rows, err := r.q.GetAlleTische(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tische := make([]tisch.Tisch, 0, len(rows))
	for _, row := range rows {
		tische = append(tische, tischRowToDomain(row))
	}

	return tische, nil
}

// GetAlleTischNamen maps every Tisch ID to its name, deleted Tische included.
func (r Repository) GetAlleTischNamen(ctx context.Context) (map[int]string, error) {
	rows, err := r.q.GetAlleTischNamen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	namen := make(map[int]string, len(rows))
	for _, row := range rows {
		namen[row.ID] = row.Name
	}

	return namen, nil
}

// GetTischSaldiOffeneSitzung maps tischID to saldoCents from the tisch_sessions projection of the
// offene Kassensitzung; without one the map is empty.
func (r Repository) GetTischSaldiOffeneSitzung(ctx context.Context) (map[int]int, error) {
	rows, err := r.q.GetTischSaldiOffeneSitzung(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make(map[int]int, len(rows))
	for _, row := range rows {
		result[row.TischID] = row.SaldoCents
	}

	return result, nil
}

func (r Repository) GetAktiveTischeMitFavoriten(ctx context.Context, userID int, kassensitzungNr int) ([]tisch.AktiverTischMitFavorit, error) {
	rows, err := r.q.GetAktiveTischeMitFavoriten(ctx, dbgen.GetAktiveTischeMitFavoritenParams{
		UserID:          userID,
		KassensitzungNr: kassensitzungNr,
	})
	if err != nil {
		return nil, db.Error(err)
	}

	tische := make([]tisch.AktiverTischMitFavorit, 0, len(rows))
	for _, row := range rows {
		tische = append(tische, tisch.AktiverTischMitFavorit{
			ID:         row.ID,
			Name:       row.Name,
			SaldoCents: row.SaldoCents,
			IstFavorit: row.IstFavorit,
		})
	}

	return tische, nil
}

func (r Repository) CreateTisch(ctx context.Context, t tisch.Tisch) (int, error) {
	id, err := r.q.CreateTisch(ctx, dbgen.CreateTischParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateTisch(ctx context.Context, t tisch.Tisch) error {
	result, err := r.q.UpdateTisch(ctx, updateTischParams(t))
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}

// AendereStatus persists t's new status under the Tisch row lock. With saldoSchutz it refuses an open
// saldo with tisch.ErrOffenerSaldo; a booking raising the saldo waits for this lock, so none slips past
// the check. A deleted Tisch loses its Favoriten in the same transaction: it leaves the Tischauswahl,
// so they could no longer be removed.
func (r Repository) AendereStatus(ctx context.Context, t tisch.Tisch, saldoSchutz bool) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		if _, err := qtx.GetTischStatusForUpdate(ctx, t.ID); err != nil {
			return db.Error(err)
		}
		if saldoSchutz {
			hatSaldo, err := qtx.TischHatOffenenSaldo(ctx, t.ID)
			if err != nil {
				return db.Error(err)
			}
			if hatSaldo {
				return tisch.ErrOffenerSaldo
			}
		}
		if t.Status == tisch.DeletedStatus {
			if err := qtx.RemoveFavoritenByTisch(ctx, t.ID); err != nil {
				return db.Error(err)
			}
		}
		result, err := qtx.UpdateTisch(ctx, updateTischParams(t))
		if err != nil {
			return db.Error(err)
		}
		return db.ResultError(result)
	})
}

func updateTischParams(t tisch.Tisch) dbgen.UpdateTischParams {
	return dbgen.UpdateTischParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		UpdatedAt: t.UpdatedAt,
		ID:        t.ID,
	}
}
