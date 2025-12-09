package event

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// Persistence implements product persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

// WriteEvent stores a new event in the database.
func (p *Persistence) WriteEvent(ctx context.Context, event Event) (int, error) {
	log := zerolog.Ctx(ctx)

	var id int
	err := p.DB.QueryRowContext(ctx,
		`INSERT INTO events (user_id, type, subject, data, timestamp)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		event.UserID,
		event.Type,
		event.Subject,
		event.Data,
		event.Time,
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Str("event_type", event.Type).Msg("DB Error creating event")
		return 0, db.Error(err)
	}

	return id, nil
}

func (p *Persistence) ReadEvent(ctx context.Context, eventID int) (*Event, error) {
	log := zerolog.Ctx(ctx)

	row := p.DB.QueryRowContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		FROM events
		WHERE id = $1`,
		eventID,
	)

	var event Event
	if err := row.Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &event.Data, &event.Time); err != nil {
		log.Error().Err(err).Int("event_id", eventID).Msg("DB Error scanning event row")
		return nil, db.Error(err)
	}

	return &event, nil
}

// ReadEventsBySubject retrieves all events of the specified types from the database for the given subject.
// Events are ordered by their sequence number ascending (first element in slice is first event).
func (p *Persistence) ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]Event, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		 FROM events
		 WHERE subject = $1 AND type = ANY($2)
		 ORDER BY id ASC`,
		subject,
		eventTypes,
	)
	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("DB Error querying events")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "events", log)

	events := []Event{}
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &event.Data, &event.Time); err != nil {
			log.Error().Err(err).Msg("DB Error scanning event row")
			return nil, db.Error(err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over event rows")
		return nil, db.Error(err)
	}

	return events, nil
}
