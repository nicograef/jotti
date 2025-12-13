package event_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
)

type Repository struct {
	DB *sql.DB
}

// WriteEvent stores a new event in the database.
func (r Repository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		`INSERT INTO events (user_id, type, subject, data, timestamp)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		e.UserID,
		e.Type,
		e.Subject,
		e.Data,
		e.Time,
	).Scan(&id)

	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) ReadEvent(ctx context.Context, eventID int) (event.Event, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp	FROM events	WHERE id = $1`,
		eventID,
	)

	var e event.Event
	if err := row.Scan(&e.ID, &e.UserID, &e.Type, &e.Subject, &e.Data, &e.Time); err != nil {
		return e, db.Error(err)
	}

	return e, nil
}

// ReadEventsBySubject retrieves all events of the given subject.
// Events are ordered by their sequence number ascending (first element in slice is first event).
func (r Repository) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, user_id, type, subject, data, timestamp FROM events WHERE subject = $1 ORDER BY id ASC`, subject)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "events")

	events := []event.Event{}
	for rows.Next() {
		var event event.Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &event.Data, &event.Time); err != nil {
			return nil, db.Error(err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return events, nil
}
