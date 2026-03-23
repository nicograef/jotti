//go:build unit

package kassenjournal_repo

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
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
	events            map[int]event.Event
	err               error
	writeErr          error // separate error for WriteEvent
	tischSessions     map[string]kasse.TischSession
	tischSessionErr   error
	offeneKS          *kasse.KassensitzungState
	offeneKSErr       error
}

func (m *mockRepo) WriteEvent(_ context.Context, e event.Event, _ kasse.StreamType, _ int) (int, error) {
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

func (m *mockRepo) GetMaxVersion(_ context.Context, subject string) (int, error) {
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

func (m *mockRepo) ReadEvent(_ context.Context, eventID int) (event.Event, error) {
	e, ok := m.events[eventID]
	if !ok {
		return event.Event{}, m.err
	}
	return e, m.err
}

func (m *mockRepo) ReadEventsBySubject(_ context.Context, subject string) ([]event.Event, error) {
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

func (m *mockRepo) ReadTischSession(_ context.Context, subject string) (kasse.TischSession, error) {
	if m.tischSessionErr != nil {
		return kasse.TischSession{}, m.tischSessionErr
	}
	if m.err != nil {
		return kasse.TischSession{}, m.err
	}
	if m.tischSessions != nil {
		if state, ok := m.tischSessions[subject]; ok {
			return state, nil
		}
	}
	return kasse.TischSession{}, nil
}

// SetTischSession sets the projected state for a given subject in the mock.
func (m *mockRepo) SetTischSession(subject string, state kasse.TischSession) {
	if m.tischSessions == nil {
		m.tischSessions = make(map[string]kasse.TischSession)
	}
	m.tischSessions[subject] = state
}

func (m *mockRepo) GetOffeneKassensitzung(_ context.Context) (*kasse.KassensitzungState, error) {
	if m.offeneKSErr != nil {
		return nil, m.offeneKSErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.offeneKS, nil
}

// SetOffeneKassensitzung sets the open Kassensitzung for the mock.
func (m *mockRepo) SetOffeneKassensitzung(ks *kasse.KassensitzungState) {
	m.offeneKS = ks
}

// AddEvent adds an event to the mock for ReadEventsBySubject.
func (m *mockRepo) AddEvent(e event.Event) {
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
}

func (m *mockRepo) GetBestellungEventsSinceCursor(_ context.Context, cursor int) ([]event.Event, error) {
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
