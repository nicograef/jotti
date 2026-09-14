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
type Event struct {
	ID       int
	UserID   int
	UserName string
	// Event type incl. version, e.g. bestellung-aufgenommen:v1
	Type string
	Time time.Time
	// The entity the event belongs to, e.g. kassensitzung-1/tisch-42
	Subject string
	// The version of the event for optimistic concurrency control.
	Version int
	Data    json.RawMessage
}

// New sets Time; Version is NOT set here — the OCC mechanism in the application
// layer assigns it.
func New(userID int, userName string, eventType string, subject string, data any) (Event, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	if err := validateFields(userID, userName, eventType, subject, dataJSON); err != nil {
		return Event{}, err
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

// validateFields checks the fields shared by New and Validate. Time and Version
// are validated only by Validate: New has not assigned them yet.
func validateFields(userID int, userName, eventType, subject string, data json.RawMessage) error {
	if userID <= 0 {
		return errors.New("user ID must be a positive integer")
	}
	if len(strings.TrimSpace(userName)) == 0 {
		return errors.New("user name must be a non-empty string")
	}
	if len(strings.TrimSpace(eventType)) < 5 {
		return errors.New("event type must be at least 5 characters long")
	}
	if len(strings.TrimSpace(subject)) < 3 {
		return errors.New("event subject must be a non-empty string")
	}
	if len(data) == 0 {
		return errors.New("event data cannot be empty")
	}
	return nil
}

func (e *Event) Validate() error {
	if err := validateFields(e.UserID, e.UserName, e.Type, e.Subject, e.Data); err != nil {
		return err
	}

	if e.Time.IsZero() {
		return errors.New("event time cannot be zero")
	}

	if e.Version < 1 {
		return errors.New("event version must be >= 1")
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
