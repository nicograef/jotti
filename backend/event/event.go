package event

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Event represents an event (message) following the CNCF Cloudevents specification.
type Event struct {
	// Identifies the event. Must be unique within the scope of the producer/source.
	ID uuid.UUID `json:"id"`
	// The ID of the user associated with the event.
	UserID int `json:"userId"`
	// The type of event related to the source system and subject. E.g. com.library.book.borrowed:v1
	Type string `json:"type"`
	// The timestamp of when the event occurred.
	Time time.Time `json:"time"`
	// The subject of the event in the context of the event producer (identified by source). E.g. the entity to which the event is primarily related. E.g. /users/12345
	Subject string `json:"subject"`
	// The event payload.
	Data any `json:"data"`
}

// Candidate represents the input required to create a new Event.
type Candidate struct {
	// The ID of the user associated with the event.
	UserID int `json:"userId"`
	// The type of event related to the source system and subject. E.g. com.library.book.borrowed:v1
	Type string `json:"type"`
	// The subject of the event in the context of the event producer (identified by source). E.g. the entity to which the event is primarily related. E.g. /users/12345"
	Subject string `json:"subject"`
	// The event payload.
	Data any `json:"data"`
}

// New creates a new Event with the given parameters and automatically sets the ID and Time fields.
// It returns an error if any of the required fields are invalid.
func New(candidate Candidate) (*Event, error) {
	event := Event{
		ID:      uuid.New(),
		UserID:  candidate.UserID,
		Type:    candidate.Type,
		Time:    time.Now().UTC(),
		Subject: candidate.Subject,
		Data:    candidate.Data,
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	return &event, nil
}

// Validate checks the Event fields for validity according to the CNCF Cloudevents specification.
func (e *Event) Validate() error {
	if e.ID == uuid.Nil {
		return errors.New("event ID cannot be nil")
	}

	if e.UserID <= 0 {
		return errors.New("user ID must be a positive integer")
	}

	if len(strings.TrimSpace(e.Type)) < 5 {
		return errors.New("event type must be at least 5 characters long")
	}

	if e.Time.IsZero() {
		return errors.New("event time cannot be zero")
	}

	if len(strings.TrimSpace(e.Subject)) < 5 {
		return errors.New("event subject must be at least 5 characters long")
	}

	if e.Data == nil {
		return errors.New("event data cannot be nil")
	}

	return nil
}
