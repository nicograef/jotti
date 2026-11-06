package kassenjournal_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// ReadPfandrueckgabeEvents returns the Kassensitzung's Pfandrückgaben and their cancellations at
// one table, ordered by ID ascending; tischID 0 is the Direktverkauf.
func (r Repository) ReadPfandrueckgabeEvents(ctx context.Context, kassensitzungNr int, tischID int) ([]event.Event, error) {
	rows, err := r.q.ReadPfandrueckgabeEvents(ctx, dbgen.ReadPfandrueckgabeEventsParams{KassensitzungNr: kassensitzungNr, TischID: tischID})
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for i := range rows {
		events = append(events, eventFromReadRow(dbgen.ReadEventsBySubjectRow(rows[i])))
	}

	return events, nil
}
