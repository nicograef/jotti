//go:build unit

package kassenjournal_repo

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
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
	NextZNr                int // z_nr, die EroeffneKassensitzung vergibt (0 → 1)
	events                 map[int]event.Event
	err                    error
	writeErr               error // separate error for WriteEvent
	kassensitzungEventsErr error // returned only by ReadKassensitzungEvents
	tischSessions          map[string]kasse.TischSession
	tischNames             map[int]string // Tischnamen für ReadFavoritenTischStates
	tischSessionErr        error
	kassenbestand          int                                   // configurable return value for GetKassenbestand
	druckauftraege         []druckauftrag_repo.NeuerDruckauftrag // captured via WriteEventWithDruckauftraege
}

// versionConflict mirrors the UNIQUE(subject, version) constraint of the kassenjournal.
func (m *MockRepo) versionConflict(e event.Event) bool {
	for _, existing := range m.events {
		if existing.Subject == e.Subject && existing.Version == e.Version {
			return true
		}
	}
	return false
}

// EroeffneKassensitzung mirrors the atomic open: assigns the next z_nr, runs build,
// and stores the event. NextZNr configures the assigned number (default 1).
func (m *MockRepo) EroeffneKassensitzung(ctx context.Context, _ time.Time, _ string, build func(zNr int) (event.Event, error)) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	zNr := m.NextZNr
	if zNr == 0 {
		zNr = 1
	}
	evt, err := build(zNr)
	if err != nil {
		return 0, err
	}
	if _, err := m.WriteEvent(ctx, evt, kasse.StreamTypeKassensitzung, zNr); err != nil {
		return 0, err
	}
	return zNr, nil
}

func (m *MockRepo) WriteEvent(_ context.Context, e event.Event, _ kasse.StreamType, _ int) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	if m.err != nil {
		return 0, m.err
	}
	if m.versionConflict(e) {
		return 0, db.ErrAlreadyExists
	}
	newID := len(m.events) + 1
	e.ID = newID
	m.events[newID] = e
	return newID, nil
}

func (m *MockRepo) WriteEventWithDruckauftraege(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	id, err := m.WriteEvent(ctx, e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}
	e.ID = id
	m.druckauftraege = append(m.druckauftraege, buildAuftraege(e)...)
	return id, nil
}

func (m *MockRepo) WriteTischSessionEventsAtomic(_ context.Context, events []event.Event, _ int) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.err != nil {
		return m.err
	}

	for _, evt := range events {
		if m.versionConflict(evt) {
			return db.ErrAlreadyExists
		}
		newID := len(m.events) + 1
		evt.ID = newID
		m.events[newID] = evt
	}

	return nil
}

func (m *MockRepo) WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, kassensitzungNr int) error {
	return m.WriteTischSessionEventsAtomic(ctx, []event.Event{quellEvent, zielEvent}, kassensitzungNr)
}

// CapturedDruckauftraege returns the print jobs produced via WriteEventWithDruckauftraege.
func (m *MockRepo) CapturedDruckauftraege() []druckauftrag_repo.NeuerDruckauftrag {
	return m.druckauftraege
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

// SetTischName registers a tisch name so ReadFavoritenTischStates returns the tisch.
// A tisch id without a registered name is treated as missing (deleted/unknown).
func (m *MockRepo) SetTischName(tischID int, name string) {
	if m.tischNames == nil {
		m.tischNames = make(map[int]string)
	}
	m.tischNames[tischID] = name
}

// ReadFavoritenTischStates mirrors the batch join: a tisch id with a registered
// name maps to its name plus the projected session (zero-value when none exists);
// an id without a name is absent from the map (like a deleted/unknown tisch).
func (m *MockRepo) ReadFavoritenTischStates(_ context.Context, tischIDs []int, kassensitzungNr int) (map[int]TischNameUndSession, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]TischNameUndSession, len(tischIDs))
	for _, tischID := range tischIDs {
		name, ok := m.tischNames[tischID]
		if !ok {
			continue
		}
		var session kasse.TischSession
		if m.tischSessions != nil {
			if s, ok := m.tischSessions[kasse.TischSessionSubject(kassensitzungNr, tischID)]; ok {
				session = s
			}
		}
		result[tischID] = TischNameUndSession{Name: name, Session: session}
	}
	return result, nil
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

// SetReadKassensitzungEventsErr configures an error returned only by ReadKassensitzungEvents.
// Use to trigger a post-barrier failure without affecting other journal operations.
func (m *MockRepo) SetReadKassensitzungEventsErr(err error) {
	m.kassensitzungEventsErr = err
}

// EventExistsByTypeAndVorgangsID prüft, ob im Mock ein gespeichertes Event des
// gegebenen Typs existiert, bei dem data[jsonKey] == vorgangsID.
func (m *MockRepo) EventExistsByTypeAndVorgangsID(_ context.Context, eventType, vorgangsID, jsonKey string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for _, e := range m.events {
		if e.Type != eventType {
			continue
		}
		var data map[string]json.RawMessage
		if err := json.Unmarshal(e.Data, &data); err != nil {
			continue
		}
		val, ok := data[jsonKey]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(val, &s); err != nil {
			continue
		}
		if s == vorgangsID {
			return true, nil
		}
	}
	return false, nil
}

// ReadKassensitzungEvents returns all events whose subject belongs to the given
// Kassensitzung (exact match or prefix "kassensitzung-N/"), ordered by ID ascending.
func (m *MockRepo) ReadKassensitzungEvents(_ context.Context, kassensitzungNr int) ([]event.Event, error) {
	if m.kassensitzungEventsErr != nil {
		return nil, m.kassensitzungEventsErr
	}
	if m.err != nil {
		return nil, m.err
	}
	prefix := kasse.KassensitzungSubject(kassensitzungNr)
	var events []event.Event
	for _, e := range m.events {
		if e.Subject == prefix || strings.HasPrefix(e.Subject, prefix+"/") {
			events = append(events, e)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].ID < events[j].ID
	})
	return events, nil
}
