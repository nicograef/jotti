//go:build unit

package event_repo

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/table"
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

// NewMockWithWriteErr creates a mock that always returns writeErr on WriteEvent calls.
func NewMockWithWriteErr(events []event.Event, writeErr error) *mockRepo {
	eventMap := make(map[int]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	return &mockRepo{
		events:   eventMap,
		writeErr: writeErr,
	}
}

type mockRepo struct {
	events        map[int]event.Event
	err           error
	writeErr      error // separate error for WriteEvent
	tableState    map[int]table.TischState
	tableStateErr error
}

func (m *mockRepo) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	if m.writeErr != nil {
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
	sort.Slice(events, func(i, j int) bool {
		return events[i].ID < events[j].ID
	})
	return events, m.err
}

func (m *mockRepo) ReadTableState(ctx context.Context, tischID int) (table.TischState, error) {
	if m.tableStateErr != nil {
		return table.TischState{}, m.tableStateErr
	}
	if m.err != nil {
		return table.TischState{}, m.err
	}
	if m.tableState != nil {
		if state, ok := m.tableState[tischID]; ok {
			return state, nil
		}
	}
	return table.TischState{}, nil
}

// SetTableState sets the projected state for a given tisch ID in the mock.
func (m *mockRepo) SetTableState(tischID int, state table.TischState) {
	if m.tableState == nil {
		m.tableState = make(map[int]table.TischState)
	}
	m.tableState[tischID] = state
}

// AddEvent adds an event to the mock for ReadEventsBySubject.
func (m *mockRepo) AddEvent(e event.Event) {
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
}

func (m *mockRepo) GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []event.Event
	for _, e := range m.events {
		if e.ID > cursor {
			result = append(result, e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}
