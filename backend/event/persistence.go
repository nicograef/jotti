package event

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api"
	"github.com/rs/zerolog/log"
)

// Persistence implements product persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

// WriteEvent stores a new event in the database.
func (p *Persistence) WriteEvent(ctx context.Context, event Event) (uuid.UUID, error) {
	correlationID, _ := ctx.Value(api.CorrelationIDKey).(string)

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to marshal event data")
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = p.DB.QueryRowContext(ctx,
		`INSERT INTO events (id, correlation_id, user_id, type, subject, data, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		event.ID,
		correlationID,
		event.UserID,
		event.Type,
		event.Subject,
		dataJSON,
		event.Time,
	).Scan(&id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to create event")
		return uuid.Nil, err
	}

	log.Debug().Str("correlation_id", correlationID).Str("event_id", id.String()).Msg("Event created")
	return id, nil
}

func (p *Persistence) ReadEvent(ctx context.Context, eventID uuid.UUID) (*Event, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)

	var event Event
	var dataJSON []byte
	err := p.DB.QueryRowContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		 FROM events
		 WHERE id = $1`,
		eventID,
	).Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &dataJSON, &event.Time)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Str("correlation_id", correlationID).Str("event_id", eventID.String()).Msg("Event not found")
			return nil, nil
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to read event by ID")
		return nil, err
	}

	if err := json.Unmarshal(dataJSON, &event.Data); err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to unmarshal event data")
		return nil, err
	}

	log.Debug().Str("correlation_id", correlationID).Str("event_id", event.ID.String()).Msg("Event read by ID")
	return &event, nil
}

// ReadEventsBySubject retrieves all events of the specified types from the database for the given subject.
// Events are ordered by their sequence number ascending (first element in slice is first event).
func (p *Persistence) ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]Event, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)

	rows, err := p.DB.QueryContext(ctx,
		`SELECT id, user_id, type, subject, data, timestamp
		 FROM events
		 WHERE subject = $1 AND type = ANY($2)
		 ORDER BY sequence ASC`,
		subject,
		eventTypes,
	)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to read events by subject")
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Error().Str("correlation_id", correlationID).Err(closeErr).Msg("Failed to close rows")
		}
	}()

	var events []Event
	for rows.Next() {
		var event Event
		var dataJSON []byte
		if err := rows.Scan(&event.ID, &event.UserID, &event.Type, &event.Subject, &dataJSON, &event.Time); err != nil {
			log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to scan event row")
			return nil, err
		}
		if err := json.Unmarshal(dataJSON, &event.Data); err != nil {
			log.Error().Str("correlation_id", correlationID).Err(err).Msg("Failed to unmarshal event data")
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Msg("Row iteration error")
		return nil, err
	}

	log.Debug().Str("correlation_id", correlationID).Int("event_count", len(events)).Msg("Events read by subject")
	return events, nil
}
