package event_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/event"
)

// NewMock creates a new mock repository with the given events and error.
func NewMock(events []event.Event, err error) *mockRepo {
	eventMap := make(map[int]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	return &mockRepo{
		events: eventMap,
		err:    err,
	}
}

// NewMockWithWriteErr creates a mock that returns writeErr on the first failCount
// WriteEvent calls, then succeeds. Useful for testing OCC retry logic.
func NewMockWithWriteErr(events []event.Event, writeErr error, failCount int) *mockRepo {
	eventMap := make(map[int]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	return &mockRepo{
		events:       eventMap,
		writeErr:     writeErr,
		writeErrLeft: failCount,
	}
}

type mockRepo struct {
	events       map[int]event.Event
	err          error
	writeErr     error // separate error for WriteEvent (used in OCC tests)
	writeErrLeft int   // number of times WriteEvent should return writeErr before succeeding
}

func (m *mockRepo) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	if m.writeErr != nil && m.writeErrLeft > 0 {
		m.writeErrLeft--
		return 0, m.writeErr
	}
	if m.err != nil {
		return 0, m.err
	}
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
	return newID, nil
}

func (m *mockRepo) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	maxVersion := 0
	for _, e := range m.events {
		if e.Subject == subject && e.Version > maxVersion {
			maxVersion = e.Version
		}
	}
	return maxVersion, nil
}

func (m *mockRepo) ReadEvent(ctx context.Context, eventID int) (event.Event, error) {
	e, ok := m.events[eventID]
	if !ok {
		return event.Event{}, m.err
	}
	return e, m.err
}

func (m *mockRepo) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	events := []event.Event{}
	for _, e := range m.events {
		if e.Subject == subject {
			events = append(events, e)
		}
	}
	return events, m.err
}

func (m *mockRepo) ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error) {
	return m.ReadEventsBySubject(ctx, subject)
}
