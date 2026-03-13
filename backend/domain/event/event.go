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
	ID int
	// The ID of the user associated with the event.
	UserID int
	// The name of the user who triggered the event.
	UserName string
	// The type of event related to the source system and subject. E.g. com.library.book.borrowed:v1
	Type string
	// The timestamp of when the event occurred.
	Time time.Time
	// The subject of the event in the context of the event producer (identified by source). E.g. the entity to which the event is primarily related. E.g. /users/12345
	Subject string
	// The version of the event for optimistic concurrency control.
	Version int
	// The event payload.
	Data json.RawMessage
}

// New creates a new Event with the given parameters and automatically sets the Time field.
// Version is NOT set here — it is assigned by the OCC mechanism in the application layer.
// It returns an error if any of the required fields are invalid.
func New(userID int, userName string, eventType string, subject string, data any) (Event, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	if userID <= 0 {
		return Event{}, errors.New("user ID must be a positive integer")
	}
	if len(strings.TrimSpace(userName)) == 0 {
		return Event{}, errors.New("user name must be a non-empty string")
	}
	if len(strings.TrimSpace(eventType)) < 5 {
		return Event{}, errors.New("event type must be at least 5 characters long")
	}
	if len(strings.TrimSpace(subject)) < 3 {
		return Event{}, errors.New("event subject must be a non-empty string")
	}
	if len(dataJSON) == 0 {
		return Event{}, errors.New("event data cannot be empty")
	}

	event := Event{
		UserID:   userID,
		UserName: userName,
		Type:     eventType,
		Time:     time.Now().UTC(),
		Subject:  subject,
		Data:     dataJSON,
	}

	return event, nil
}

// Validate checks the Event fields for validity according to the CNCF Cloudevents specification.
func (e *Event) Validate() error {
	if e.UserID <= 0 {
		return errors.New("user ID must be a positive integer")
	}

	if len(strings.TrimSpace(e.UserName)) == 0 {
		return errors.New("user name must be a non-empty string")
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

	if e.Version < 1 {
		return errors.New("event version must be >= 1")
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

	if errs := schema.Validate(dest); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("validation failed: %v", issues)
	}

	return nil
}
