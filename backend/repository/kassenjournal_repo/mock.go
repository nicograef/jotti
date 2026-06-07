//go:build unit

package kassenjournal_repo

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// NewMock creates a new mock repository with the given events and error.
func NewMock(events []event.Event, err error) *MockRepo {
	eventMap := make(map[int]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	return &MockRepo{
		events: eventMap,
		err:    err,
	}
}

// NewMockWithWriteErr creates a mock that always returns writeErr on WriteEvent calls.
func NewMockWithWriteErr(events []event.Event, writeErr error) *MockRepo {
	eventMap := make(map[int]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	return &MockRepo{
		events:   eventMap,
		writeErr: writeErr,
	}
}

type MockRepo struct {
	events          map[int]event.Event
	err             error
	writeErr        error // separate error for WriteEvent
	tischSessions   map[string]kasse.TischSession
	tischSessionErr error
	kassenbestand   int // configurable return value for GetKassenbestand
}

func (m *MockRepo) WriteEvent(_ context.Context, e event.Event, _ kasse.StreamType, _ int) (int, error) {
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

func (m *MockRepo) GetMaxVersion(_ context.Context, subject string) (int, error) {
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

func (m *MockRepo) ReadEventsBySubject(_ context.Context, subject string) ([]event.Event, error) {
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

func (m *MockRepo) ReadTischSession(_ context.Context, subject string) (kasse.TischSession, error) {
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
func (m *MockRepo) SetTischSession(subject string, state kasse.TischSession) {
	if m.tischSessions == nil {
		m.tischSessions = make(map[string]kasse.TischSession)
	}
	m.tischSessions[subject] = state
}

func (m *MockRepo) GetKassenbestand(_ context.Context, _ int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.kassenbestand, nil
}

// SetKassenbestand sets the return value for GetKassenbestand.
func (m *MockRepo) SetKassenbestand(cents int) {
	m.kassenbestand = cents
}

// AddEvent adds an event to the mock for ReadEventsBySubject.
func (m *MockRepo) AddEvent(e event.Event) {
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
}

func (m *MockRepo) GetTischSessionsByKassensitzungNr(_ context.Context, kassensitzungNr int) ([]kasse.TischSession, error) {
	if m.err != nil {
		return nil, m.err
	}
	var sessions []kasse.TischSession
	for _, s := range m.tischSessions {
		if s.KassensitzungNr == kassensitzungNr {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}
