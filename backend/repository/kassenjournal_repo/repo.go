package kassenjournal_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{DB: database, q: dbgen.New(database)}
}

// WriteEvent stores a new event in the kassenjournal and synchronously updates
// the appropriate projection within the same transaction.
// Routing by streamType:
//   - "kassensitzung" → INSERT/UPDATE kassensitzungen (CRUD entity)
//   - "tisch-session" → UPSERT tisch_sessions (synchronous projection)
//   - "direktverkauf" → kassenjournal only (no projection)
func (r Repository) WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stored, err := r.writeEventInTx(ctx, r.q.WithTx(tx), e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return stored.ID, nil
}

// WriteEventWithDruckauftraege writes an event and the print jobs derived from it
// within a single transaction (transactional outbox). buildAuftraege receives the
// stored event including its generated ID, so a print job's referenz can depend on
// it. If inserting a print job fails, the event is rolled back — there is never an
// order without its work tickets, nor work tickets without their order.
func (r Repository) WriteEventWithDruckauftraege(
	ctx context.Context,
	e event.Event,
	streamType kasse.StreamType,
	kassensitzungNr int,
	buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag,
) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}

	if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, buildAuftraege(stored)); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return stored.ID, nil
}

// WriteEventWithNachsignierAuftrag writes an event and enqueues a TSE retry
// job in one transaction. This is used for the DON'T BLOCK THE TILL fallback:
// the sale is persisted immediately and signing can be retried asynchronously.
func (r Repository) WriteEventWithNachsignierAuftrag(
	ctx context.Context,
	e event.Event,
	streamType kasse.StreamType,
	kassensitzungNr int,
	txID string,
	processType string,
	processData string,
) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}

	err = qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
		TxID:        txID,
		ProcessType: processType,
		ProcessData: processData,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return stored.ID, nil
}

// WriteEventWithDruckauftraegeUndNachsignierAuftrag writes an event, derived
// print jobs, and a TSE retry job in one transaction.
func (r Repository) WriteEventWithDruckauftraegeUndNachsignierAuftrag(
	ctx context.Context,
	e event.Event,
	streamType kasse.StreamType,
	kassensitzungNr int,
	buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag,
	txID string,
	processType string,
	processData string,
) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}

	if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, buildAuftraege(stored)); err != nil {
		return 0, err
	}

	err = qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
		TxID:        txID,
		ProcessType: processType,
		ProcessData: processData,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return stored.ID, nil
}

// WriteUmbuchung writes the source stornierung and target bestellung atomically.
// Both events must already carry their final subject/version.
func (r Repository) WriteUmbuchung(ctx context.Context, stornierungEvent event.Event, bestellungEvent event.Event, kassensitzungNr int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	if _, err := r.writeEventInTx(ctx, qtx, stornierungEvent, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		return err
	}

	if _, err := r.writeEventInTx(ctx, qtx, bestellungEvent, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return db.Error(err)
	}

	return nil
}

// writeEventInTx inserts the event into the kassenjournal and updates the matching
// projection within the given transaction, returning the event with its generated
// ID. The caller owns the transaction (commit/rollback).
func (r Repository) writeEventInTx(ctx context.Context, qtx *dbgen.Queries, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (event.Event, error) {
	// 1. Insert event into kassenjournal
	id, err := qtx.WriteEvent(ctx, dbgen.WriteEventParams{
		UserID:          e.UserID,
		UserName:        e.UserName,
		Type:            e.Type,
		Subject:         e.Subject,
		Version:         e.Version,
		Data:            e.Data,
		Timestamp:       e.Time,
		KassensitzungNr: kassensitzungNr,
	})
	if err != nil {
		return event.Event{}, db.Error(err)
	}

	e.ID = id

	// 2. Route to the appropriate projection/entity
	switch streamType {
	case kasse.StreamTypeKassensitzung:
		if err := r.handleKassensitzungEvent(ctx, qtx, e, kassensitzungNr); err != nil {
			return event.Event{}, err
		}

	case kasse.StreamTypeTischSession:
		if err := r.handleTischSessionEvent(ctx, qtx, e, kassensitzungNr); err != nil {
			return event.Event{}, err
		}

	case kasse.StreamTypeDirektverkauf:
		// Direktverkauf lives entirely in the kassenjournal — no projection to update.

	default:
		return event.Event{}, fmt.Errorf("unknown stream type: %s", streamType)
	}

	return e, nil
}

// handleKassensitzungEvent handles kassensitzung events by updating the kassensitzungen CRUD entity.
// Note: For kassensitzung-eroeffnet:v1, the kassensitzungen row is created by the application layer
// BEFORE calling WriteEvent (required because kassenjournal has a FK to kassensitzungen).
// The repo only handles tagesabschluss-erstellt:v1 (setting status to 'abgeschlossen').
func (r Repository) handleKassensitzungEvent(ctx context.Context, qtx *dbgen.Queries, e event.Event, kassensitzungNr int) error {
	switch e.Type {
	case string(kasse.EventTypeTagesabschlussErstelltV1):
		err := qtx.UpdateKassensitzung(ctx, dbgen.UpdateKassensitzungParams{
			ZNr:    kassensitzungNr,
			Status: kasse.KassensitzungAbgeschlossen,
		})
		if err != nil {
			return db.Error(err)
		}

	default:
		// Other kassensitzung events (eroeffnet, anfangsbestand, kassenbewegung, kassensturz, differenz)
		// don't change the CRUD entity — only the kassenjournal entry is written.
	}

	return nil
}

// handleTischSessionEvent handles tisch-session events by upserting the tisch_sessions projection.
func (r Repository) handleTischSessionEvent(ctx context.Context, qtx *dbgen.Queries, e event.Event, kassensitzungNr int) error {
	tischID, err := kasse.ParseTischIDFromSubject(e.Subject)
	if err != nil {
		return fmt.Errorf("parse tisch ID from subject %q: %w", e.Subject, err)
	}

	// Read current TischSession within TX
	currentState, err := getTischSessionInTx(ctx, qtx, e.Subject)
	if err != nil {
		return err
	}

	// Apply event to state
	newState, err := kasse.ApplyEvent(currentState, e)
	if err != nil {
		return fmt.Errorf("apply event to tisch session: %w", err)
	}

	// Marshal positions to JSON
	unbezahltJSON, err := json.Marshal(newState.UnbezahltePositionen)
	if err != nil {
		return fmt.Errorf("marshal unbezahlte positionen: %w", err)
	}
	ausstehendeJSON, err := json.Marshal(newState.AusstehendePositionen)
	if err != nil {
		return fmt.Errorf("marshal ausstehende positionen: %w", err)
	}

	// Upsert tisch_sessions
	err = qtx.UpsertTischSession(ctx, dbgen.UpsertTischSessionParams{
		Subject:               e.Subject,
		TischID:               tischID,
		KassensitzungNr:       kassensitzungNr,
		SaldoCents:            newState.SaldoCents,
		UnbezahltePositionen:  unbezahltJSON,
		AusstehendePositionen: ausstehendeJSON,
		GesamtZahlungenCents:  newState.GesamtZahlungenCents,
		LastEventID:           newState.LastEventID,
		LastEventVersion:      newState.LastEventVersion,
	})
	if err != nil {
		return db.Error(err)
	}

	return nil
}

// ReadTischSession reads the projected state of a tisch session by subject.
// Returns zero-value TischSession if no entry exists (no events written yet).
func (r Repository) ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error) {
	row, err := r.q.GetTischSession(ctx, subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return kasse.TischSession{}, nil
		}
		return kasse.TischSession{}, db.Error(err)
	}

	return toTischSession(row)
}

// getTischSessionInTx reads the current TischSession within a transaction.
// Returns zero-value TischSession if no entry exists.
func getTischSessionInTx(ctx context.Context, qtx *dbgen.Queries, subject string) (kasse.TischSession, error) {
	row, err := qtx.GetTischSession(ctx, subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return kasse.TischSession{}, nil
		}
		return kasse.TischSession{}, db.Error(err)
	}

	return toTischSession(row)
}

func toTischSession(row dbgen.TischSession) (kasse.TischSession, error) {
	var unbezahlt []kasse.Position
	if err := json.Unmarshal(row.UnbezahltePositionen, &unbezahlt); err != nil {
		return kasse.TischSession{}, fmt.Errorf("unmarshal unbezahlte positionen: %w", err)
	}

	var ausstehende []kasse.Position
	if err := json.Unmarshal(row.AusstehendePositionen, &ausstehende); err != nil {
		return kasse.TischSession{}, fmt.Errorf("unmarshal ausstehende positionen: %w", err)
	}

	return kasse.TischSession{
		Subject:               row.Subject,
		TischID:               row.TischID,
		KassensitzungNr:       row.KassensitzungNr,
		SaldoCents:            row.SaldoCents,
		UnbezahltePositionen:  unbezahlt,
		AusstehendePositionen: ausstehende,
		GesamtZahlungenCents:  row.GesamtZahlungenCents,
		LastEventID:           row.LastEventID,
		LastEventVersion:      row.LastEventVersion,
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
	for i := range rows {
		events = append(events, event.Event{
			ID:       rows[i].ID,
			UserID:   rows[i].UserID,
			UserName: rows[i].UserName,
			Version:  rows[i].Version,
			Type:     rows[i].Type,
			Subject:  rows[i].Subject,
			Data:     rows[i].Data,
			Time:     rows[i].Timestamp,
		})
	}

	return events, nil
}

// ReadDirektverkaufEvents retrieves all Direktverkauf events (getaetigt + storniert) of the given
// Kassensitzung across all verkauf streams, ordered by ID ascending.
func (r Repository) ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error) {
	rows, err := r.q.ReadDirektverkaufEvents(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for i := range rows {
		events = append(events, event.Event{
			ID:       rows[i].ID,
			UserID:   rows[i].UserID,
			UserName: rows[i].UserName,
			Version:  rows[i].Version,
			Type:     rows[i].Type,
			Subject:  rows[i].Subject,
			Data:     rows[i].Data,
			Time:     rows[i].Timestamp,
		})
	}

	return events, nil
}

// GetTSESignaturByTxID returns the backfilled signature data for a TSE tx_id.
// It is used when the event was persisted unsigned during a temporary TSE
// outage and later signed by the retry worker.
func (r Repository) GetTSESignaturByTxID(ctx context.Context, txID string) (kasse.TSEData, error) {
	row, err := r.q.GetTSESignaturByTxID(ctx, strings.TrimSpace(txID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return kasse.TSEData{}, db.ErrNotFound
		}
		return kasse.TSEData{}, db.Error(err)
	}

	return kasse.TSEData{
		TransactionNumber: row.TransaktionNummer,
		SignatureCounter:  row.SignaturZaehler,
		SerialNumberTSE:   row.TseSeriennummer,
		LogTimeStart:      row.LogTimeStart.UTC().Format(time.RFC3339),
		LogTimeEnd:        row.LogTimeEnd.UTC().Format(time.RFC3339),
		Signature:         row.Signatur,
		QRCodeData:        row.QrCodeData,
	}, nil
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

// RebuildAllProjections replays all events and rebuilds the tisch_sessions projection from scratch.
// Runs in a single transaction: deletes all existing tisch_sessions rows, then replays all events
// per subject and upserts the resulting state.
// Note: kassensitzungen is a CRUD entity and is NOT replayed.
// Returns the number of subjects rebuilt.
func (r Repository) RebuildAllProjections(ctx context.Context) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)

	// 1. Delete all existing tisch_sessions projections
	if err := qtx.DeleteAllTischSession(ctx); err != nil {
		return 0, fmt.Errorf("delete all tisch sessions: %w", err)
	}

	// 2. Get all distinct subjects
	subjects, err := qtx.GetDistinctSubjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("get distinct subjects: %w", err)
	}

	// 3. Replay events for each tisch-session subject only
	rebuiltCount := 0
	for _, subject := range subjects {
		// Only rebuild tisch-session subjects (contain "/tisch-")
		if !isTischSessionSubject(subject) {
			continue
		}

		tischID, err := kasse.ParseTischIDFromSubject(subject)
		if err != nil {
			return 0, fmt.Errorf("parse tisch ID from subject %q: %w", subject, err)
		}

		kassensitzungNr, err := kasse.ParseZNrFromSubject(subject)
		if err != nil {
			return 0, fmt.Errorf("parse z_nr from subject %q: %w", subject, err)
		}

		rows, err := qtx.ReadEventsBySubject(ctx, subject)
		if err != nil {
			return 0, fmt.Errorf("read events for subject %q: %w", subject, err)
		}

		state := kasse.TischSession{}
		for i := range rows {
			evt := event.Event{
				ID:       rows[i].ID,
				UserID:   rows[i].UserID,
				UserName: rows[i].UserName,
				Version:  rows[i].Version,
				Type:     rows[i].Type,
				Subject:  rows[i].Subject,
				Data:     rows[i].Data,
				Time:     rows[i].Timestamp,
			}
			state, err = kasse.ApplyEvent(state, evt)
			if err != nil {
				return 0, fmt.Errorf("apply event %d to subject %q: %w", rows[i].ID, subject, err)
			}
		}

		unbezahltJSON, err := json.Marshal(state.UnbezahltePositionen)
		if err != nil {
			return 0, fmt.Errorf("marshal unbezahlte positionen for subject %q: %w", subject, err)
		}
		ausstehendeJSON, err := json.Marshal(state.AusstehendePositionen)
		if err != nil {
			return 0, fmt.Errorf("marshal ausstehende positionen for subject %q: %w", subject, err)
		}

		err = qtx.UpsertTischSession(ctx, dbgen.UpsertTischSessionParams{
			Subject:               subject,
			TischID:               tischID,
			KassensitzungNr:       kassensitzungNr,
			SaldoCents:            state.SaldoCents,
			UnbezahltePositionen:  unbezahltJSON,
			AusstehendePositionen: ausstehendeJSON,
			GesamtZahlungenCents:  state.GesamtZahlungenCents,
			LastEventID:           state.LastEventID,
			LastEventVersion:      state.LastEventVersion,
		})
		if err != nil {
			return 0, fmt.Errorf("upsert tisch session for subject %q: %w", subject, err)
		}

		rebuiltCount++
	}

	// 4. Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, db.Error(err)
	}

	return rebuiltCount, nil
}

// GetKassenbestand returns the Soll-Kassenbestand for the given Kassensitzung in cents.
func (r Repository) GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error) {
	bestand, err := r.q.GetKassenbestand(ctx, kassensitzungNr)
	if err != nil {
		return 0, db.Error(err)
	}
	return bestand, nil
}

// GetTischSessionsByKassensitzungNr returns all tisch sessions for a given Kassensitzung.
// Used by TagesabschlussErstellen to check the Tisch-Saldo-Sperre (all saldi must be 0).
func (r Repository) GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error) {
	rows, err := r.q.GetTischSessionsByKassensitzungNr(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	sessions := make([]kasse.TischSession, 0, len(rows))
	for i := range rows {
		session, err := toTischSession(rows[i])
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// isTischSessionSubject checks if a subject is a tisch-session subject (contains "/tisch-").
func isTischSessionSubject(subject string) bool {
	return strings.Contains(subject, "/tisch-")
}
