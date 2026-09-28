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

// GetAlleTischNamen includes deleted Tische: the DSFinV-K export of past Kassensitzungen
// resolves their names, where GetAlleTische would hide them.
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

// TischHatOffenenSaldo meldet, ob der Tisch in der offenen Kassensitzung einen
// offenen Saldo trägt (Schutz-Guard). Ohne offene Sitzung immer false.
func (r Repository) TischHatOffenenSaldo(ctx context.Context, tischID int) (bool, error) {
	hat, err := r.q.TischHatOffenenSaldo(ctx, tischID)
	if err != nil {
		return false, db.Error(err)
	}

	return hat, nil
}

func (r Repository) GetAktiveTische(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error) {
	rows, err := r.q.GetAktiveTische(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	tische := make([]tisch.AktiverTisch, 0, len(rows))
	for _, row := range rows {
		tische = append(tische, tisch.AktiverTisch{
			ID:         row.ID,
			Name:       row.Name,
			SaldoCents: row.SaldoCents,
		})
	}

	return tische, nil
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
	result, err := r.q.UpdateTisch(ctx, dbgen.UpdateTischParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		UpdatedAt: t.UpdatedAt,
		ID:        t.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}

// DeleteTischMitFavoriten shares one transaction because Favoriten of a deleted tisch
// would be invisible and unremovable. The caller applies Delete() first.
func (r Repository) DeleteTischMitFavoriten(ctx context.Context, t tisch.Tisch) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		if err := qtx.RemoveFavoritenByTisch(ctx, t.ID); err != nil {
			return db.Error(err)
		}

		result, err := qtx.UpdateTisch(ctx, dbgen.UpdateTischParams{
			Name:      t.Name,
			Status:    dbgen.Entitystatus(t.Status),
			UpdatedAt: t.UpdatedAt,
			ID:        t.ID,
		})
		if err != nil {
			return db.Error(err)
		}
		return db.ResultError(result)
	})
}
