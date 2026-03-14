package event_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{DB: db, q: dbgen.New(db)}
}

// WriteEvent stores a new event in the database and synchronously updates the table_state projection
// within the same transaction.
func (r Repository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	// 1. Insert event
	id, err := qtx.WriteEvent(ctx, dbgen.WriteEventParams{
		UserID:    e.UserID,
		UserName:  e.UserName,
		Type:      e.Type,
		Subject:   e.Subject,
		Version:   e.Version,
		Data:      e.Data,
		Timestamp: e.Time,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	// 2. Extract tischID from subject (format: "tisch:<id>")
	tischID, err := parseTischID(e.Subject)
	if err != nil {
		return 0, fmt.Errorf("parse tisch ID from subject %q: %w", e.Subject, err)
	}

	// 3. Read current TischState within TX
	currentState, err := getTableStateInTx(ctx, qtx, tischID)
	if err != nil {
		return 0, err
	}

	// 4. Apply event to state
	e.ID = id
	newState, err := table.ApplyEvent(currentState, e)
	if err != nil {
		return 0, fmt.Errorf("apply event to table state: %w", err)
	}

	// 5. Marshal positions to JSON
	unbezahltJSON, err := json.Marshal(newState.UnbezahltePositionen)
	if err != nil {
		return 0, fmt.Errorf("marshal unbezahlte positionen: %w", err)
	}
	ungeliefertJSON, err := json.Marshal(newState.AusstehendePositionen)
	if err != nil {
		return 0, fmt.Errorf("marshal ausstehende positionen: %w", err)
	}

	// 6. Upsert table_state
	err = qtx.UpsertTableState(ctx, dbgen.UpsertTableStateParams{
		TischID:               tischID,
		SaldoCents:            newState.SaldoCents,
		UnbezahltePositionen:  unbezahltJSON,
		AusstehendePositionen: ungeliefertJSON,
		GesamtZahlungenCents:  newState.GesamtZahlungenCents,
		LastEventID:           newState.LastEventID,
		LastEventVersion:      newState.LastEventVersion,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

// ReadTableState reads the projected state of a table from the table_state table.
// Returns zero-value TischState if no entry exists (no events written yet).
func (r Repository) ReadTableState(ctx context.Context, tischID int) (table.TischState, error) {
	row, err := r.q.GetTableState(ctx, tischID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return table.TischState{}, nil
		}
		return table.TischState{}, db.Error(err)
	}

	return toTischState(row)
}

// parseTischID extracts the table ID from a subject string (format: "tisch:<id>").
func parseTischID(subject string) (int, error) {
	prefix := "tisch:"
	if !strings.HasPrefix(subject, prefix) {
		return 0, fmt.Errorf("subject %q does not start with %q", subject, prefix)
	}
	return strconv.Atoi(subject[len(prefix):])
}

// getTableStateInTx reads the current TischState within a transaction.
// Returns zero-value TischState if no entry exists.
func getTableStateInTx(ctx context.Context, qtx *dbgen.Queries, tischID int) (table.TischState, error) {
	row, err := qtx.GetTableState(ctx, tischID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return table.TischState{}, nil
		}
		return table.TischState{}, db.Error(err)
	}

	return toTischState(row)
}

func toTischState(row dbgen.TableState) (table.TischState, error) {
	var unbezahlt []table.Position
	if err := json.Unmarshal(row.UnbezahltePositionen, &unbezahlt); err != nil {
		return table.TischState{}, fmt.Errorf("unmarshal unbezahlte positionen: %w", err)
	}

	var ungeliefert []table.Position
	if err := json.Unmarshal(row.AusstehendePositionen, &ungeliefert); err != nil {
		return table.TischState{}, fmt.Errorf("unmarshal ausstehende positionen: %w", err)
	}

	return table.TischState{
		SaldoCents:            row.SaldoCents,
		UnbezahltePositionen:  unbezahlt,
		AusstehendePositionen: ungeliefert,
		GesamtZahlungenCents:  row.GesamtZahlungenCents,
		LastEventID:           row.LastEventID,
		LastEventVersion:      row.LastEventVersion,
	}, nil
}

func (r Repository) ReadEvent(ctx context.Context, eventID int) (event.Event, error) {
	row, err := r.q.ReadEvent(ctx, eventID)
	if err != nil {
		return event.Event{}, db.Error(err)
	}

	return event.Event{
		ID:       row.ID,
		UserID:   row.UserID,
		UserName: row.UserName,
		Version:  row.Version,
		Type:     row.Type,
		Subject:  row.Subject,
		Data:     row.Data,
		Time:     row.Timestamp,
	}, nil
}

// ReadEventsBySubject retrieves all events of the given subject.
// Events are ordered by ID ascending (first element in slice is first event).
func (r Repository) ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error) {
	rows, err := r.q.ReadEventsBySubject(ctx, subject)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, event.Event{
			ID:       row.ID,
			UserID:   row.UserID,
			UserName: row.UserName,
			Version:  row.Version,
			Type:     row.Type,
			Subject:  row.Subject,
			Data:     row.Data,
			Time:     row.Timestamp,
		})
	}

	return events, nil
}

// GetMaxVersion returns the highest event version for the given subject.
// Returns 0 if no events exist for the subject.
func (r Repository) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	version, err := r.q.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, db.Error(err)
	}

	return version, nil
}

// RebuildAllProjections replays all events and rebuilds the table_state projection from scratch.
// Runs in a single transaction: deletes all existing table_state rows, then replays all events
// per subject and upserts the resulting state. Returns the number of subjects rebuilt.
func (r Repository) RebuildAllProjections(ctx context.Context) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	// 1. Delete all existing projections
	if err := qtx.DeleteAllTableState(ctx); err != nil {
		return 0, fmt.Errorf("delete all table state: %w", err)
	}

	// 2. Get all distinct subjects
	subjects, err := qtx.GetDistinctSubjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("get distinct subjects: %w", err)
	}

	// 3. Replay events for each subject
	for _, subject := range subjects {
		tischID, err := parseTischID(subject)
		if err != nil {
			return 0, fmt.Errorf("parse tisch ID from subject %q: %w", subject, err)
		}

		rows, err := qtx.ReadEventsBySubject(ctx, subject)
		if err != nil {
			return 0, fmt.Errorf("read events for subject %q: %w", subject, err)
		}

		state := table.TischState{}
		for _, row := range rows {
			evt := event.Event{
				ID:       row.ID,
				UserID:   row.UserID,
				UserName: row.UserName,
				Version:  row.Version,
				Type:     row.Type,
				Subject:  row.Subject,
				Data:     row.Data,
				Time:     row.Timestamp,
			}
			state, err = table.ApplyEvent(state, evt)
			if err != nil {
				return 0, fmt.Errorf("apply event %d to subject %q: %w", row.ID, subject, err)
			}
		}

		unbezahltJSON, err := json.Marshal(state.UnbezahltePositionen)
		if err != nil {
			return 0, fmt.Errorf("marshal unbezahlte positionen for subject %q: %w", subject, err)
		}
		ungeliefertJSON, err := json.Marshal(state.AusstehendePositionen)
		if err != nil {
			return 0, fmt.Errorf("marshal ausstehende positionen for subject %q: %w", subject, err)
		}

		err = qtx.UpsertTableState(ctx, dbgen.UpsertTableStateParams{
			TischID:               tischID,
			SaldoCents:            state.SaldoCents,
			UnbezahltePositionen:  unbezahltJSON,
			AusstehendePositionen: ungeliefertJSON,
			GesamtZahlungenCents:  state.GesamtZahlungenCents,
			LastEventID:           state.LastEventID,
			LastEventVersion:      state.LastEventVersion,
		})
		if err != nil {
			return 0, fmt.Errorf("upsert table state for subject %q: %w", subject, err)
		}
	}

	// 4. Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return len(subjects), nil
}
