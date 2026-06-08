package druckauftrag_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type NeuerDruckauftrag struct {
	ZielIP   string
	Payload  string
	BonArt   string
	Referenz string
}

type OffenerDruckauftrag struct {
	ID      int
	ZielIP  string
	Payload string
}

type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{DB: database, q: dbgen.New(database)}
}

func (r Repository) EnqueueDruckauftraege(ctx context.Context, auftraege []NeuerDruckauftrag) error {
	if len(auftraege) == 0 {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if err := InsertDruckauftraege(ctx, r.q.WithTx(tx), auftraege); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return db.Error(err)
	}

	return nil
}

// InsertDruckauftraege inserts the given print jobs using the provided
// transaction-bound queries. The caller owns the transaction, which enables a
// transactional outbox: writing an event and its resulting print jobs atomically
// (see kassenjournal_repo.WriteEventWithDruckauftraege).
func InsertDruckauftraege(ctx context.Context, qtx *dbgen.Queries, auftraege []NeuerDruckauftrag) error {
	for _, auftrag := range auftraege {
		err := qtx.InsertDruckauftrag(ctx, dbgen.InsertDruckauftragParams{
			ZielIp:   auftrag.ZielIP,
			Payload:  auftrag.Payload,
			BonArt:   auftrag.BonArt,
			Referenz: auftrag.Referenz,
		})
		if err != nil {
			return db.Error(err)
		}
	}

	return nil
}

func (r Repository) GetOffeneDruckauftraege(ctx context.Context) ([]OffenerDruckauftrag, error) {
	rows, err := r.q.GetOffeneDruckauftraege(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]OffenerDruckauftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, OffenerDruckauftrag{
			ID:      int(row.ID),
			ZielIP:  row.ZielIp,
			Payload: row.Payload,
		})
	}

	return result, nil
}

func (r Repository) QuittiereGedruckteAuftraege(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)
	for _, id := range ids {
		err := qtx.MarkDruckauftragGedruckt(ctx, int32(id))
		if err != nil {
			return db.Error(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return db.Error(err)
	}

	return nil
}
