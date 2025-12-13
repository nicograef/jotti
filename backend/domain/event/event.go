package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
)

// Event represents a CNCF Cloudevent with additional fields for user association.
// Identifies the event. Must be unique within the scope of the producer/source.
type Event struct {
	ID int `json:"id"`
	// The ID of the user associated with the event.
	UserID int `json:"userId"`
	// The type of event related to the source system and subject. E.g. com.library.book.borrowed:v1
	Type string `json:"type"`
	// The timestamp of when the event occurred.
	Time time.Time `json:"time"`
	// The subject of the event in the context of the event producer (identified by source). E.g. the entity to which the event is primarily related. E.g. /users/12345
	Subject string `json:"subject"`
	// The event payload.
	Data json.RawMessage `json:"data"`
}

// New creates a new Event with the given parameters and automatically sets the ID and Time fields.
// It returns an error if any of the required fields are invalid.
func New(userID int, eventType string, subject string, data any) (Event, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	event := Event{
		UserID:  userID,
		Type:    eventType,
		Time:    time.Now().UTC(),
		Subject: subject,
		Data:    dataJSON,
	}

	if err := event.Validate(); err != nil {
		return Event{}, err
	}

	return event, nil
}

// Validate checks the Event fields for validity according to the CNCF Cloudevents specification.
func (e *Event) Validate() error {
	if e.UserID <= 0 {
		return errors.New("user ID must be a positive integer")
	}

	if len(strings.TrimSpace(e.Type)) < 5 {
		return errors.New("event type must be at least 5 characters long")
	}

	if e.Time.IsZero() {
		return errors.New("event time cannot be zero")
	}

	if len(strings.TrimSpace(e.Subject)) < 3 {
		return errors.New("event subject must be a non-empty string")
	}

	if len(e.Data) == 0 {
		return errors.New("event data cannot be empty")
	}

	return nil
}

func ParseData[T any](e Event, dest *T, schema *z.StructSchema) error {
	if err := json.Unmarshal(e.Data, dest); err != nil {
		return err
	}

	if errsMap := schema.Validate(dest); errsMap != nil {
		issues := z.Issues.SanitizeMapAndCollect(errsMap)
		return fmt.Errorf("validation failed: %v", issues)
	}

	return nil
}
