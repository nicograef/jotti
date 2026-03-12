package event_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{DB: db, q: dbgen.New(db)}
}

// WriteEvent stores a new event in the database.
func (r Repository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	id, err := r.q.WriteEvent(ctx, dbgen.WriteEventParams{
		UserID:    e.UserID,
		UserName:  e.UserName,
		Type:      e.Type,
		Subject:   e.Subject,
		Version:   e.Version,
		Data:      e.Data,
		Timestamp: e.Time,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) ReadEvent(ctx context.Context, eventID int) (event.Event, error) {
	row, err := r.q.ReadEvent(ctx, eventID)
	if err != nil {
		return event.Event{}, db.Error(err)
	}

	return event.Event{
		ID:       row.ID,
		UserID:   row.UserID,
		UserName: row.UserName,
		Version:  row.Version,
		Type:     row.Type,
		Subject:  row.Subject,
		Data:     row.Data,
		Time:     row.Timestamp,
	}, nil
}

// ReadEventsBySubject retrieves all events of the given subject.
// Events are ordered by ID ascending (first element in slice is first event).
func (r Repository) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	rows, err := r.q.ReadEventsBySubject(ctx, subject)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, event.Event{
			ID:       row.ID,
			UserID:   row.UserID,
			UserName: row.UserName,
			Version:  row.Version,
			Type:     row.Type,
			Subject:  row.Subject,
			Data:     row.Data,
			Time:     row.Timestamp,
		})
	}

	return events, nil
}

// ReadEventsSinceID retrieves events for a subject starting from a given ID (inclusive).
// This is useful for reading events since the last snapshot.
func (r Repository) ReadEventsSinceID(ctx context.Context, subject string, fromID int) ([]event.Event, error) {
	rows, err := r.q.ReadEventsSinceID(ctx, dbgen.ReadEventsSinceIDParams{
		Subject: subject,
		ID:      fromID,
	})
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, event.Event{
			ID:       row.ID,
			UserID:   row.UserID,
			UserName: row.UserName,
			Version:  row.Version,
			Type:     row.Type,
			Subject:  row.Subject,
			Data:     row.Data,
			Time:     row.Timestamp,
		})
	}

	return events, nil
}

// GetLastSnapshotID returns the ID of the most recent snapshot for a subject.
// Returns 0 if no snapshot exists.
func (r Repository) GetLastSnapshotID(ctx context.Context, subject string, snapshotEventType string) (int, error) {
	id, err := r.q.GetLastSnapshotID(ctx, dbgen.GetLastSnapshotIDParams{
		Subject: subject,
		Type:    snapshotEventType,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

// ReadEventsWithSnapshot finds the last snapshot for the subject and returns
// all events starting from that snapshot (inclusive). If no snapshot exists,
// all events for the subject are returned.
func (r Repository) ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error) {
	rows, err := r.q.ReadEventsWithSnapshot(ctx, dbgen.ReadEventsWithSnapshotParams{
		Subject: subject,
		Type:    snapshotEventType,
	})
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, event.Event{
			ID:       row.ID,
			UserID:   row.UserID,
			UserName: row.UserName,
			Version:  row.Version,
			Type:     row.Type,
			Subject:  row.Subject,
			Data:     row.Data,
			Time:     row.Timestamp,
		})
	}

	return events, nil
}

// GetMaxVersion returns the highest event version for the given subject.
// Returns 0 if no events exist for the subject.
func (r Repository) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	version, err := r.q.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, db.Error(err)
	}

	return version, nil
}
