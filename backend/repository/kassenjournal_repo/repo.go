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

// ErrKassensitzungNichtOffen: der Event-Write traf auf eine Kassensitzung, die
// nicht (mehr) offen ist — z. B. weil parallel der Tagesabschluss lief. Die
// Application-Schicht mappt das auf ihren Konflikt-Fehler (HTTP 409).
var ErrKassensitzungNichtOffen = errors.New("kassensitzung ist nicht offen")

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

// withTx runs fn within a single transaction: it begins the tx, rolls back on
// any error (a rollback after commit is a no-op), and commits otherwise. fn
// receives the transaction-bound queries and owns its own error wrapping; only
// begin/commit failures are normalized via db.Error.
func (r Repository) withTx(ctx context.Context, fn func(*dbgen.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if err := fn(r.q.WithTx(tx)); err != nil {
		return err
	}

	return db.Error(tx.Commit())
}

// WriteEvent stores a new event in the kassenjournal and synchronously updates
// the appropriate projection within the same transaction.
// Routing by streamType:
//   - "kassensitzung" → INSERT/UPDATE kassensitzungen (CRUD entity)
//   - "tisch-session" → UPSERT tisch_sessions (synchronous projection)
//   - "direktverkauf" → kassenjournal only (no projection)
func (r Repository) WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	var id int
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}
		id = stored.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
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
	var id int
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}

		if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, buildAuftraege(stored)); err != nil {
			return err
		}

		id = stored.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
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
	var id int
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}

		err = qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
			TxID:        txID,
			ProcessType: processType,
			ProcessData: processData,
		})
		if err != nil {
			return db.Error(err)
		}

		id = stored.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
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
	var id int
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}

		if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, buildAuftraege(stored)); err != nil {
			return err
		}

		err = qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
			TxID:        txID,
			ProcessType: processType,
			ProcessData: processData,
		})
		if err != nil {
			return db.Error(err)
		}

		id = stored.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// TSENachsignierung beschreibt einen TSE-Nachsignier-Auftrag, der atomar mit einem
// während eines TSE-Ausfalls unsigniert persistierten Vorgang geschrieben wird.
type TSENachsignierung struct {
	TxID        string
	ProcessType string
	ProcessData string
}

// WriteTischSessionEventsAtomic writes the given tisch-session events atomically
// (all-or-nothing), together with any TSE-Nachsignier-Aufträge for events that could
// not be signed at capture time. Each event must already carry its final subject and
// version. Backs UI actions that map to multiple typed events with one TSE-transaction
// each — the Umbuchung (two linked tables) and the Storno (one geldneutrale Korrektur
// plus one Warenrücknahme per betroffener Zahlung).
func (r Repository) WriteTischSessionEventsAtomic(ctx context.Context, events []event.Event, nachsignierungen []TSENachsignierung, kassensitzungNr int) error {
	return r.withTx(ctx, func(qtx *dbgen.Queries) error {
		for _, evt := range events {
			if _, err := r.writeEventInTx(ctx, qtx, evt, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
				return err
			}
		}

		for _, ns := range nachsignierungen {
			err := qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
				TxID:        ns.TxID,
				ProcessType: ns.ProcessType,
				ProcessData: ns.ProcessData,
			})
			if err != nil {
				return db.Error(err)
			}
		}

		return nil
	})
}

// EroeffneKassensitzung legt die Kassensitzungs-Entität an und schreibt das
// Eröffnungs-Event in EINER Transaktion: Schlägt der Event-Write fehl, bleibt
// keine offene Sitzung ohne Eröffnungs-Event (und damit ohne Anfangsbestand)
// zurück. build erhält die vergebene z_nr, erzeugt und signiert das Event und
// liefert optional den Nachsignier-Auftrag eines TSE-Ausfalls.
func (r Repository) EroeffneKassensitzung(ctx context.Context, datum time.Time, bezeichnung string, build func(zNr int) (event.Event, *TSENachsignierung, error)) (int, error) {
	var zNr int
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		n, err := qtx.InsertKassensitzung(ctx, dbgen.InsertKassensitzungParams{
			Datum:       datum,
			Bezeichnung: bezeichnung,
			Status:      string(kasse.KassensitzungOffen),
		})
		if err != nil {
			return db.Error(err)
		}

		evt, nachsignierung, err := build(n)
		if err != nil {
			return err
		}

		if _, err := r.writeEventInTx(ctx, qtx, evt, kasse.StreamTypeKassensitzung, n); err != nil {
			return err
		}

		if nachsignierung != nil {
			err := qtx.InsertTSENachsignierAuftrag(ctx, dbgen.InsertTSENachsignierAuftragParams{
				TxID:        nachsignierung.TxID,
				ProcessType: nachsignierung.ProcessType,
				ProcessData: nachsignierung.ProcessData,
			})
			if err != nil {
				return db.Error(err)
			}
		}

		zNr = n
		return nil
	})
	if err != nil {
		return 0, err
	}

	return zNr, nil
}

// WriteUmbuchung writes the linked source and target umbuchung events atomically,
// together with any TSE-Nachsignier-Aufträge for sides that could not be signed at
// capture time. Both events must already carry their final subject/version.
func (r Repository) WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, nachsignierungen []TSENachsignierung, kassensitzungNr int) error {
	return r.WriteTischSessionEventsAtomic(ctx, []event.Event{quellEvent, zielEvent}, nachsignierungen, kassensitzungNr)
}

// writeEventInTx inserts the event into the kassenjournal and updates the matching
// projection within the given transaction, returning the event with its generated
// ID. The caller owns the transaction (commit/rollback).
func (r Repository) writeEventInTx(ctx context.Context, qtx *dbgen.Queries, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (event.Event, error) {
	// 0. Status-Guard mit Zeilensperre: Buchungs-Events dürfen nur in eine offene Kassensitzung.
	// FOR SHARE serialisiert gegen den Statuswechsel auf 'wird_abgeschlossen', den KasseAbschliessen
	// als erste Handlung committet (UPDATE = FOR UPDATE): Entweder committet dieser Write vor der
	// Barriere (und wird von der Saldo-Sperre erfasst), oder er sieht den neuen Status und scheitert.
	// Im Zwischenstatus 'wird_abgeschlossen' passieren nur die drei Abschluss-Events selbst; der
	// Tagesabschluss schreibt in diesem Status und setzt in derselben Transaktion 'abgeschlossen'.
	status, err := qtx.GetKassensitzungStatusForShare(ctx, kassensitzungNr)
	if err != nil {
		return event.Event{}, db.Error(err)
	}
	switch kasse.KassensitzungStatus(status) {
	case kasse.KassensitzungOffen:
		// Alle Events erlaubt.
	case kasse.KassensitzungWirdAbgeschlossen:
		if !kasse.IsAbschlussEventType(e.Type) {
			return event.Event{}, ErrKassensitzungNichtOffen
		}
	default: // abgeschlossen
		return event.Event{}, ErrKassensitzungNichtOffen
	}

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
			Status: string(kasse.KassensitzungAbgeschlossen),
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

	return upsertTischSessionState(ctx, qtx, e.Subject, tischID, kassensitzungNr, newState)
}

// upsertTischSessionState marshals the projected positions and upserts the
// tisch_sessions row for the given subject. Shared by handleTischSessionEvent
// (event-write path) and RebuildAllProjections (replay path).
func upsertTischSessionState(ctx context.Context, qtx *dbgen.Queries, subject string, tischID, kassensitzungNr int, state kasse.TischSession) error {
	unbezahltJSON, err := json.Marshal(state.UnbezahltePositionen)
	if err != nil {
		return fmt.Errorf("marshal unbezahlte positionen: %w", err)
	}
	ausstehendeJSON, err := json.Marshal(state.AusstehendePositionen)
	if err != nil {
		return fmt.Errorf("marshal ausstehende positionen: %w", err)
	}

	err = qtx.UpsertTischSession(ctx, dbgen.UpsertTischSessionParams{
		Subject:                subject,
		TischID:                tischID,
		KassensitzungNr:        kassensitzungNr,
		SaldoCents:             state.SaldoCents,
		UnbezahltePositionen:   unbezahltJSON,
		AusstehendePositionen:  ausstehendeJSON,
		GesamtZahlungenCents:   state.GesamtZahlungenCents,
		ErsteBestellungLogtime: toNullTime(state.ErsteBestellungLogTime),
		LastEventID:            state.LastEventID,
		LastEventVersion:       state.LastEventVersion,
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

	var ersteBestellungLogTime *time.Time
	if row.ErsteBestellungLogtime.Valid {
		v := row.ErsteBestellungLogtime.Time.UTC()
		ersteBestellungLogTime = &v
	}

	return kasse.TischSession{
		Subject:                row.Subject,
		TischID:                row.TischID,
		KassensitzungNr:        row.KassensitzungNr,
		SaldoCents:             row.SaldoCents,
		UnbezahltePositionen:   unbezahlt,
		AusstehendePositionen:  ausstehende,
		GesamtZahlungenCents:   row.GesamtZahlungenCents,
		ErsteBestellungLogTime: ersteBestellungLogTime,
		LastEventID:            row.LastEventID,
		LastEventVersion:       row.LastEventVersion,
	}, nil
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}

// eventFromReadRow maps a kassenjournal read row to a domain event.
func eventFromReadRow(row dbgen.ReadEventsBySubjectRow) event.Event {
	return event.Event{
		ID:       row.ID,
		UserID:   row.UserID,
		UserName: row.UserName,
		Version:  row.Version,
		Type:     row.Type,
		Subject:  row.Subject,
		Data:     row.Data,
		Time:     row.Timestamp,
	}
}

// eventFromDirektverkaufRow maps a Direktverkauf read row to a domain event.
// ReadDirektverkaufEvents selects the same columns as ReadEventsBySubject but
// sqlc emits a distinct row type, so it needs its own mapping.
func eventFromDirektverkaufRow(row dbgen.ReadDirektverkaufEventsRow) event.Event {
	return event.Event{
		ID:       row.ID,
		UserID:   row.UserID,
		UserName: row.UserName,
		Version:  row.Version,
		Type:     row.Type,
		Subject:  row.Subject,
		Data:     row.Data,
		Time:     row.Timestamp,
	}
}

// eventFromKassensitzungRow maps a Kassensitzung-wide read row to a domain event.
// ReadEventsByKassensitzung selects the same columns as ReadEventsBySubject but
// sqlc emits a distinct row type, so it needs its own mapping.
func eventFromKassensitzungRow(row dbgen.ReadEventsByKassensitzungRow) event.Event {
	return event.Event{
		ID:       row.ID,
		UserID:   row.UserID,
		UserName: row.UserName,
		Version:  row.Version,
		Type:     row.Type,
		Subject:  row.Subject,
		Data:     row.Data,
		Time:     row.Timestamp,
	}
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
		events = append(events, eventFromReadRow(rows[i]))
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
		events = append(events, eventFromDirektverkaufRow(rows[i]))
	}

	return events, nil
}

// ReadEventsByKassensitzung retrieves all events of the given Kassensitzung
// (Kassensitzungs-, Tisch-Session- and Direktverkauf-Streams) ordered by ID
// ascending. It is the read side of the DSFinV-K export.
func (r Repository) ReadEventsByKassensitzung(ctx context.Context, kassensitzungNr int) ([]event.Event, error) {
	rows, err := r.q.ReadEventsByKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for i := range rows {
		events = append(events, eventFromKassensitzungRow(rows[i]))
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
	rebuiltCount := 0
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		// 1. Delete all existing tisch_sessions projections
		if err := qtx.DeleteAllTischSession(ctx); err != nil {
			return fmt.Errorf("delete all tisch sessions: %w", err)
		}

		// 2. Get all distinct tisch-session subjects (filtered in SQL)
		subjects, err := qtx.GetDistinctTischSessionSubjects(ctx)
		if err != nil {
			return fmt.Errorf("get distinct tisch-session subjects: %w", err)
		}

		// 3. Replay events for each tisch-session subject
		for _, subject := range subjects {
			tischID, err := kasse.ParseTischIDFromSubject(subject)
			if err != nil {
				return fmt.Errorf("parse tisch ID from subject %q: %w", subject, err)
			}

			kassensitzungNr, err := kasse.ParseZNrFromSubject(subject)
			if err != nil {
				return fmt.Errorf("parse z_nr from subject %q: %w", subject, err)
			}

			rows, err := qtx.ReadEventsBySubject(ctx, subject)
			if err != nil {
				return fmt.Errorf("read events for subject %q: %w", subject, err)
			}

			state := kasse.TischSession{}
			for i := range rows {
				state, err = kasse.ApplyEvent(state, eventFromReadRow(rows[i]))
				if err != nil {
					return fmt.Errorf("apply event %d to subject %q: %w", rows[i].ID, subject, err)
				}
			}

			if err := upsertTischSessionState(ctx, qtx, subject, tischID, kassensitzungNr, state); err != nil {
				return fmt.Errorf("rebuild tisch session for subject %q: %w", subject, err)
			}

			rebuiltCount++
		}

		return nil
	})
	if err != nil {
		return 0, err
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
// Used by KasseAbschliessen to check the Tisch-Saldo-Sperre (all saldi must be 0).
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
