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

// scanEvents reads all rows from a query result and returns a slice of events.
func scanEvents(rows *sql.Rows) ([]event.Event, error) {
	events := []event.Event{}
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.Type, &e.Subject, &e.Data, &e.Time); err != nil {
			return nil, db.Error(err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}
	return events, nil
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
		`SELECT id, user_id, type, subject, data, timestamp FROM events WHERE id = $1`,
		eventID,
	)

	var e event.Event
	if err := row.Scan(&e.ID, &e.UserID, &e.Type, &e.Subject, &e.Data, &e.Time); err != nil {
		return e, db.Error(err)
	}

	return e, nil
}

// ReadEventsBySubject retrieves all events of the given subject.
// Events are ordered by ID ascending (first element in slice is first event).
func (r Repository) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp 
		 FROM events WHERE subject = $1 ORDER BY id ASC`,
		subject,
	)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "events")

	return scanEvents(rows)
}

// ReadEventsSinceID retrieves events for a subject starting from a given ID (inclusive).
// This is useful for reading events since the last snapshot.
func (r Repository) ReadEventsSinceID(ctx context.Context, subject string, fromID int) ([]event.Event, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp 
		 FROM events WHERE subject = $1 AND id >= $2 ORDER BY id ASC`,
		subject, fromID,
	)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "events")

	return scanEvents(rows)
}

// GetLastSnapshotID returns the ID of the most recent snapshot for a subject.
// Returns 0 if no snapshot exists.
func (r Repository) GetLastSnapshotID(ctx context.Context, subject string, snapshotEventType string) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(id), 0) FROM events WHERE subject = $1 AND type = $2`,
		subject, snapshotEventType,
	).Scan(&id)
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

// ReadEventsWithSnapshot finds the last snapshot for the subject and returns
// all events starting from that snapshot (inclusive). If no snapshot exists,
// all events for the subject are returned.
func (r Repository) ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error) {
	query := `
		WITH last_snapshot AS (
			SELECT COALESCE(MAX(id), 0) AS id 
			FROM events 
			WHERE subject = $1 AND type = $2
		)
		SELECT e.id, e.user_id, e.type, e.subject, e.data, e.timestamp
		FROM events e, last_snapshot ls
		WHERE e.subject = $1 AND e.id >= ls.id
		ORDER BY e.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, subject, snapshotEventType)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "events with snapshot")

	return scanEvents(rows)
}
