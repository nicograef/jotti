package event

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Persistence implements product persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

// WriteEvent stores a new event in the database.
func (p *Persistence) WriteEvent(ctx context.Context, event Event) (uuid.UUID, error) {
	logger := zerolog.Ctx(ctx)

	var id uuid.UUID
	err := p.DB.QueryRowContext(ctx,
		`INSERT INTO events (id, user_id, type, subject, data, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		event.ID,
		event.UserID,
		event.Type,
		event.Subject,
		event.Data,
		event.Time,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create event")
		return uuid.Nil, err
	}

	logger.Debug().Str("event_id", id.String()).Msg("Event created")
	return id, nil
}

func (p *Persistence) ReadEvent(ctx context.Context, eventID uuid.UUID) (*Event, error) {
	logger := zerolog.Ctx(ctx)

	var event Event
	err := p.DB.QueryRowContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		 FROM events
		 WHERE id = $1`,
		eventID,
	).Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &event.Data, &event.Time)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn().Str("event_id", eventID.String()).Msg("Event not found")
			return nil, nil
		}
		logger.Error().Err(err).Msg("Failed to read event by ID")
		return nil, err
	}

	logger.Debug().Str("event_id", event.ID.String()).Msg("Event read by ID")
	return &event, nil
}

// ReadEventsBySubject retrieves all events of the specified types from the database for the given subject.
// Events are ordered by their sequence number ascending (first element in slice is first event).
func (p *Persistence) ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]Event, error) {
	logger := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		 FROM events
		 WHERE subject = $1 AND type = ANY($2)
		 ORDER BY sequence ASC`,
		subject,
		eventTypes,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read events by subject")
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Error().Err(closeErr).Msg("Failed to close rows")
		}
	}()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &event.Data, &event.Time); err != nil {
			logger.Error().Err(err).Msg("Failed to scan event row")
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Row iteration error")
		return nil, err
	}

	logger.Debug().Int("event_count", len(events)).Msg("Events read by subject")
	return events, nil
}
