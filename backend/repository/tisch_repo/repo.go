package tisch_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetTable(ctx context.Context, id int) (tisch.Tisch, error) {
	row, err := r.q.GetTisch(ctx, id)
	if err != nil {
		return tisch.Tisch{}, db.Error(err)
	}

	return tischRowToDomain(row), nil
}

func (r Repository) GetAllTables(ctx context.Context) ([]tisch.Tisch, error) {
	rows, err := r.q.GetAlleTische(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]tisch.Tisch, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tischRowToDomain(row))
	}

	return tables, nil
}

// GetAllTableNames liefert die Namen ALLER Tische als Map tischID → Name,
// inklusive gelöschter. Für die historische Namensauflösung (DSFinV-K-Export
// vergangener Kassensitzungen), wo GetAllTables den gelöschten Tisch verschweigt
// und der Aufrufer sonst auf die Tisch-ID zurückfallen müsste.
func (r Repository) GetAllTableNames(ctx context.Context) (map[int]string, error) {
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

// GetTischSaldiOffeneSitzung liefert je Tisch mit offenem Saldo den Betrag aus
// der tisch_sessions-Projektion der offenen Kassensitzung (Map tischID →
// saldoCents). Ohne offene Sitzung ist die Map leer. Reine Journal-Projektion.
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

func (r Repository) GetActiveTables(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error) {
	rows, err := r.q.GetAktiveTische(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]tisch.AktiverTisch, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tisch.AktiverTisch{
			ID:         row.ID,
			Name:       row.Name,
			SaldoCents: row.SaldoCents,
		})
	}

	return tables, nil
}

func (r Repository) GetActiveTablesWithFavorites(ctx context.Context, userID int, kassensitzungNr int) ([]tisch.AktiverTischMitFavorit, error) {
	rows, err := r.q.GetAktiveTischeMitFavoriten(ctx, dbgen.GetAktiveTischeMitFavoritenParams{
		UserID:          userID,
		KassensitzungNr: kassensitzungNr,
	})
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]tisch.AktiverTischMitFavorit, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tisch.AktiverTischMitFavorit{
			ID:         row.ID,
			Name:       row.Name,
			SaldoCents: row.SaldoCents,
			IstFavorit: row.IstFavorit,
		})
	}

	return tables, nil
}

func (r Repository) CreateTable(ctx context.Context, t tisch.Tisch) (int, error) {
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

func (r Repository) UpdateTable(ctx context.Context, t tisch.Tisch) error {
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
