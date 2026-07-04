package kassenjournal_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/dsfinvk"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
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
	eingereiht := false
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, auftragEingereiht, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}
		id = stored.ID
		eingereiht = auftragEingereiht
		return nil
	})
	if err != nil {
		return 0, err
	}

	notifySignaturWorker(eingereiht)
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
	eingereiht := false
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		stored, auftragEingereiht, err := r.writeEventInTx(ctx, qtx, e, streamType, kassensitzungNr)
		if err != nil {
			return err
		}

		if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, buildAuftraege(stored)); err != nil {
			return err
		}

		id = stored.ID
		eingereiht = auftragEingereiht
		return nil
	})
	if err != nil {
		return 0, err
	}

	notifySignaturWorker(eingereiht)
	return id, nil
}

// WriteTischSessionEventsAtomic writes the given tisch-session events atomically
// (all-or-nothing); each event takes its Signaturauftrag from the fiskalische
// Projektion within the same transaction. Each event must already carry its final
// subject and version. Backs UI actions that map to multiple typed events — the
// Umbuchung (two linked tables) and the Storno (one geldneutrale Korrektur plus
// one Warenrücknahme per betroffener Zahlung).
func (r Repository) WriteTischSessionEventsAtomic(ctx context.Context, events []event.Event, kassensitzungNr int) error {
	eingereiht := false
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		for _, evt := range events {
			_, auftragEingereiht, err := r.writeEventInTx(ctx, qtx, evt, kasse.StreamTypeTischSession, kassensitzungNr)
			if err != nil {
				return err
			}
			eingereiht = eingereiht || auftragEingereiht
		}

		return nil
	})
	if err != nil {
		return err
	}

	notifySignaturWorker(eingereiht)
	return nil
}

// EroeffneKassensitzung legt die Kassensitzungs-Entität an und schreibt das
// Eröffnungs-Event in EINER Transaktion: Schlägt der Event-Write fehl, bleibt
// keine offene Sitzung ohne Eröffnungs-Event (und damit ohne Anfangsbestand)
// zurück. build erhält die vergebene z_nr und erzeugt das Eröffnungs-Event.
func (r Repository) EroeffneKassensitzung(ctx context.Context, datum time.Time, bezeichnung string, build func(zNr int) (event.Event, error)) (int, error) {
	var zNr int
	eingereiht := false
	err := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		n, err := qtx.InsertKassensitzung(ctx, dbgen.InsertKassensitzungParams{
			Datum:       datum,
			Bezeichnung: bezeichnung,
			Status:      string(kasse.KassensitzungOffen),
		})
		if err != nil {
			return db.Error(err)
		}

		evt, err := build(n)
		if err != nil {
			return err
		}

		_, auftragEingereiht, err := r.writeEventInTx(ctx, qtx, evt, kasse.StreamTypeKassensitzung, n)
		if err != nil {
			return err
		}

		zNr = n
		eingereiht = auftragEingereiht
		return nil
	})
	if err != nil {
		return 0, err
	}

	notifySignaturWorker(eingereiht)
	return zNr, nil
}

// WriteUmbuchung writes the linked source and target umbuchung events atomically;
// both sides take their Signaturauftrag from the fiskalische Projektion. Both
// events must already carry their final subject/version.
func (r Repository) WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, kassensitzungNr int) error {
	return r.WriteTischSessionEventsAtomic(ctx, []event.Event{quellEvent, zielEvent}, kassensitzungNr)
}

// notifySignaturWorker stößt den Signatur-Worker nach einem Commit mit neuem
// Signaturauftrag sofort an (non-blocking; der Polling-Tick bleibt Fallback).
func notifySignaturWorker(auftragEingereiht bool) {
	if auftragEingereiht {
		tse_repo.NotifySignaturWorker()
	}
}

// writeEventInTx inserts the event into the kassenjournal, enqueues its
// Signaturauftrag (transactional outbox) and updates the matching projection
// within the given transaction, returning the event with its generated ID and
// whether a Signaturauftrag was enqueued. The caller owns the transaction
// (commit/rollback) and triggers the Signatur-Worker after a successful commit.
func (r Repository) writeEventInTx(ctx context.Context, qtx *dbgen.Queries, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (event.Event, bool, error) {
	// 0. Status-Guard mit Zeilensperre: Buchungs-Events dürfen nur in eine offene Kassensitzung.
	// FOR SHARE serialisiert gegen den Statuswechsel auf 'wird_abgeschlossen', den KasseAbschliessen
	// als erste Handlung committet (UPDATE = FOR UPDATE): Entweder committet dieser Write vor der
	// Barriere (und wird von der Saldo-Sperre erfasst), oder er sieht den neuen Status und scheitert.
	// Im Zwischenstatus 'wird_abgeschlossen' passieren nur die drei Abschluss-Events selbst; der
	// Tagesabschluss schreibt in diesem Status und setzt in derselben Transaktion 'abgeschlossen'.
	status, err := qtx.GetKassensitzungStatusForShare(ctx, kassensitzungNr)
	if err != nil {
		return event.Event{}, false, db.Error(err)
	}
	switch kasse.KassensitzungStatus(status) {
	case kasse.KassensitzungOffen:
		// Alle Events erlaubt.
	case kasse.KassensitzungWirdAbgeschlossen:
		if !kasse.IsAbschlussEventType(e.Type) {
			return event.Event{}, false, ErrKassensitzungNichtOffen
		}
	default: // abgeschlossen
		return event.Event{}, false, ErrKassensitzungNichtOffen
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
		return event.Event{}, false, db.Error(err)
	}

	e.ID = id

	// 2. Transaktionale Outbox: Jeder signaturpflichtige Vorgang erhält im selben
	// Commit genau einen offenen Signaturauftrag (auch ohne TSE-Konfiguration).
	// Die fiskalische Projektion ist die einzige Stelle, die über Signaturpflicht
	// entscheidet; die tx-ID ist eine zufällige UUID.
	vorgang, signaturpflichtig, err := kasse.FiskalischeProjektion(e)
	if err != nil {
		return event.Event{}, false, err
	}
	if signaturpflichtig {
		err = qtx.InsertTSESignaturauftrag(ctx, dbgen.InsertTSESignaturauftragParams{
			EventID:     id,
			TxID:        uuid.New().String(),
			ProcessType: vorgang.ProcessType,
			ProcessData: vorgang.ProcessData,
		})
		if err != nil {
			return event.Event{}, false, db.Error(err)
		}
	}

	// 3. Route to the appropriate projection/entity
	switch streamType {
	case kasse.StreamTypeKassensitzung:
		if err := r.handleKassensitzungEvent(ctx, qtx, e, kassensitzungNr); err != nil {
			return event.Event{}, false, err
		}

	case kasse.StreamTypeTischSession:
		if err := r.handleTischSessionEvent(ctx, qtx, e, kassensitzungNr); err != nil {
			return event.Event{}, false, err
		}

	case kasse.StreamTypeDirektverkauf:
		// Direktverkauf lives entirely in the kassenjournal — no projection to update.

	default:
		return event.Event{}, false, fmt.Errorf("unknown stream type: %s", streamType)
	}

	return e, signaturpflichtig, nil
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

// eventSignaturFromKassensitzungRow maps the LEFT-JOIN columns of a read row to
// the export view of the event's Signaturauftrag: processType always, the
// Signaturspalten only once quittiert.
func eventSignaturFromKassensitzungRow(row dbgen.ReadEventsByKassensitzungRow) dsfinvk.EventSignatur {
	signatur := dsfinvk.EventSignatur{ProcessType: row.ProcessType.String}
	if row.Signatur.Valid {
		signatur.Signatur = &tse.Signatur{
			TransaktionNummer: int(row.TransaktionNummer.Int32),
			SignaturZaehler:   int(row.SignaturZaehler.Int32),
			TSESeriennummer:   row.TseSeriennummer.String,
			LogTimeStart:      row.LogTimeStart.Time,
			LogTimeEnd:        row.LogTimeEnd.Time,
			Signatur:          row.Signatur.String,
			QRCodeData:        row.QrCodeData.String,
		}
	}
	return signatur
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
// ascending, together with each event's Signaturauftrag-Stand (LEFT JOIN:
// kein Auftrag = nicht signaturpflichtig). It is the read side of the
// DSFinV-K export.
func (r Repository) ReadEventsByKassensitzung(ctx context.Context, kassensitzungNr int) ([]event.Event, map[int]dsfinvk.EventSignatur, error) {
	rows, err := r.q.ReadEventsByKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		return nil, nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	signaturen := make(map[int]dsfinvk.EventSignatur)
	for i := range rows {
		events = append(events, eventFromKassensitzungRow(rows[i]))
		if rows[i].ProcessType.Valid {
			signaturen[rows[i].ID] = eventSignaturFromKassensitzungRow(rows[i])
		}
	}

	return events, signaturen, nil
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
