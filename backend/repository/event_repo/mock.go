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

type mockRepo struct {
	events map[int]event.Event
	err    error
}

func (m mockRepo) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
	return newID, m.err
}

func (m mockRepo) ReadEvent(ctx context.Context, eventID int) (event.Event, error) {
	e, ok := m.events[eventID]
	if !ok {
		return event.Event{}, m.err
	}
	return e, m.err
}

func (m mockRepo) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	events := []event.Event{}
	for _, e := range m.events {
		if e.Subject == subject {
			events = append(events, e)
		}
	}
	return events, m.err
}

func (m mockRepo) ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error) {
	return m.ReadEventsBySubject(ctx, subject)
}
