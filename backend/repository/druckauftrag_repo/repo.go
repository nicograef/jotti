package druckauftrag_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// MaxDruckversuche ist die Anzahl gemeldeter Fehlversuche, nach der ein
// Druckauftrag als fehlgeschlagen markiert und nicht mehr ausgeliefert wird.
const MaxDruckversuche = 6

// backoffDauer returns the delay after the given 1-based failed attempt; 0 and >= MaxDruckversuche get none.
// See docs/handbuch.md §4.5.
func backoffDauer(versuch int) time.Duration {
	switch versuch {
	case 1:
		return 5 * time.Second
	case 2:
		return 15 * time.Second
	case 3:
		return 30 * time.Second
	case 4:
		return 60 * time.Second
	case 5:
		return 180 * time.Second
	default:
		return 0
	}
}

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

// Fehlversuch meldet einen fehlgeschlagenen Zustellversuch eines Druckauftrags.
type Fehlversuch struct {
	ID     int
	Fehler string
}

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

func (r Repository) EnqueueDruckauftraege(ctx context.Context, auftraege []NeuerDruckauftrag) error {
	if len(auftraege) == 0 {
		return nil
	}

	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		return InsertDruckauftraege(ctx, qtx, auftraege)
	})
}

// InsertDruckauftraege runs in the caller's transaction so an event and its print jobs commit atomically
// (transactional outbox, see kassenjournal_repo.WriteEventWithDruckauftraege).
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
			ID:      row.ID,
			ZielIP:  row.ZielIp,
			Payload: row.Payload,
		})
	}

	return result, nil
}

// ReportDruckergebnis acknowledges idempotently: the status guard 'offen' makes a twice-reported ID a no-op.
// See docs/handbuch.md §4.5.
func (r Repository) ReportDruckergebnis(ctx context.Context, gedruckteIDs []int, fehlversuche []Fehlversuch) error {
	if len(gedruckteIDs) == 0 && len(fehlversuche) == 0 {
		return nil
	}

	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		for _, id := range gedruckteIDs {
			if err := qtx.MarkDruckauftragGedruckt(ctx, id); err != nil {
				return db.Error(err)
			}
		}
		for _, f := range fehlversuche {
			row, err := qtx.IncrementDruckauftragFehlversuch(ctx, dbgen.IncrementDruckauftragFehlversuchParams{
				ID:            f.ID,
				LetzterFehler: sql.NullString{String: f.Fehler, Valid: true},
				MaxVersuche:   MaxDruckversuche,
			})
			if errors.Is(err, sql.ErrNoRows) {
				// Auftrag ist nicht (mehr) offen (z. B. bereits gedruckt oder doppelt
				// gemeldet): idempotenter No-Op, konsistent mit dem Status-Guard von
				// MarkDruckauftragGedruckt.
				continue
			}
			if err != nil {
				return db.Error(err)
			}
			// Solange der Auftrag offen bleibt, die Backoff-Fälligkeit für den
			// nächsten Versuch setzen. Beim MaxDruckversuche-ten Fehlversuch ist er
			// bereits fehlgeschlagen und wird nicht mehr ausgeliefert — kein Backoff.
			if row.Status == "offen" {
				wartezeit := backoffDauer(row.Versuche)
				if err := qtx.SetDruckauftragFaelligkeit(ctx, dbgen.SetDruckauftragFaelligkeitParams{
					ID:       f.ID,
					Sekunden: int(wartezeit / time.Second),
				}); err != nil {
					return db.Error(err)
				}
			}
		}

		return nil
	})
}

// GetFehlgeschlageneDruckauftraege liefert alle nach MaxDruckversuche
// aufgegebenen Aufträge (Status fehlgeschlagen), älteste zuerst.
func (r Repository) GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckstation.FehlgeschlagenerDruckauftrag, error) {
	rows, err := r.q.GetFehlgeschlageneDruckauftraege(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]druckstation.FehlgeschlagenerDruckauftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, druckstation.FehlgeschlagenerDruckauftrag{
			ID:            row.ID,
			BonArt:        row.BonArt,
			ZielIP:        row.ZielIp,
			Referenz:      row.Referenz,
			Versuche:      row.Versuche,
			LetzterFehler: row.LetzterFehler.String,
			ErstelltAm:    row.ErstelltAm,
		})
	}

	return result, nil
}

// RetryDruckauftrag reiht einen fehlgeschlagenen Auftrag wieder ein
// (fehlgeschlagen -> offen, versuche zurück auf 0). Der Status-Guard wirkt nur
// auf fehlgeschlagene Aufträge; andere Status bleiben unberührt.
func (r Repository) RetryDruckauftrag(ctx context.Context, id int) error {
	return db.Error(r.q.RetryDruckauftrag(ctx, id))
}

// DiscardDruckauftrag markiert einen fehlgeschlagenen Auftrag als verworfen
// (fehlgeschlagen -> verworfen). Der Eintrag bleibt in der Datenbank erhalten;
// der Status-Guard wirkt nur auf fehlgeschlagene Aufträge.
func (r Repository) DiscardDruckauftrag(ctx context.Context, id int) error {
	return db.Error(r.q.DiscardDruckauftrag(ctx, id))
}

// DiscardAlleFehlgeschlagenen verwirft alle fehlgeschlagenen Aufträge
// (fehlgeschlagen -> verworfen) und liefert die Anzahl. Der Status-Guard wirkt
// nur auf fehlgeschlagene Aufträge; andere Status bleiben unberührt.
func (r Repository) DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error) {
	n, err := r.q.DiscardAlleFehlgeschlagenenDruckauftraege(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return n, nil
}
