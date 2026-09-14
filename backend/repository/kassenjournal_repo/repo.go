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
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// ErrKassensitzungNichtOffen: der Write traf auf eine nicht (mehr) offene
// Kassensitzung; die Application-Schicht mappt das auf HTTP 409.
var ErrKassensitzungNichtOffen = errors.New("kassensitzung ist nicht offen")

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

// WriteEvent stores the event and, in the same transaction, updates the entity or
// projection streamType routes to (Direktverkauf: none).
func (r Repository) WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	var id int
	eingereiht := false
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
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

// WriteEventWithDruckauftraege writes the event and its derived print jobs in one
// transaction (transactional outbox); buildAuftraege receives the stored event
// including its generated ID.
func (r Repository) WriteEventWithDruckauftraege(
	ctx context.Context,
	e event.Event,
	streamType kasse.StreamType,
	kassensitzungNr int,
	buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag,
) (int, error) {
	var id int
	eingereiht := false
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
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

// WriteTischSessionEventsAtomic writes the given tisch-session events all-or-nothing;
// each event must already carry its final subject and version.
func (r Repository) WriteTischSessionEventsAtomic(ctx context.Context, events []event.Event, kassensitzungNr int) error {
	eingereiht := false
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
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

// EroeffneKassensitzung legt Entität und Eröffnungs-Event in EINER Transaktion an —
// sonst bliebe eine offene Sitzung ohne Anfangsbestand zurück. build erhält die
// vergebene z_nr.
func (r Repository) EroeffneKassensitzung(ctx context.Context, datum time.Time, bezeichnung string, build func(zNr int) (event.Event, error)) (int, error) {
	var zNr int
	eingereiht := false
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
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

// WriteUmbuchung writes the linked source and target events atomically; both must
// already carry their final subject and version.
func (r Repository) WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, kassensitzungNr int) error {
	return r.WriteTischSessionEventsAtomic(ctx, []event.Event{quellEvent, zielEvent}, kassensitzungNr)
}

// notifySignaturWorker stößt den Worker non-blocking an; der Polling-Tick bleibt Fallback.
func notifySignaturWorker(auftragEingereiht bool) {
	if auftragEingereiht {
		tse_repo.NotifySignaturWorker()
	}
}

// writeEventInTx writes event, Signaturauftrag and projection into the given
// transaction and returns the stored event plus whether a Signaturauftrag was
// enqueued. The caller owns the transaction and triggers the Signatur-Worker after
// the commit.
func (r Repository) writeEventInTx(ctx context.Context, qtx *dbgen.Queries, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (event.Event, bool, error) {
	// Status-Guard mit Zeilensperre: FOR SHARE serialisiert gegen den Statuswechsel auf
	// 'wird_abgeschlossen', den KasseAbschliessen als erste Handlung committet (UPDATE =
	// FOR UPDATE): Entweder committet dieser Write vor der Barriere (und wird von der
	// Saldo-Sperre erfasst), oder er sieht den neuen Status und scheitert. Im Status
	// 'wird_abgeschlossen' passieren nur die Abschluss-Events; der Tagesabschluss schreibt
	// in diesem Status und setzt in derselben Transaktion 'abgeschlossen'.
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

	// Transaktionale Outbox: jeder signaturpflichtige Vorgang erhält im selben Commit
	// genau einen offenen Signaturauftrag — auch ohne TSE-Konfiguration. Die fiskalische
	// Projektion entscheidet als einzige Stelle über die Signaturpflicht.
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

// handleKassensitzungEvent updates the kassensitzungen CRUD entity. Only
// tagesabschluss-erstellt:v1 changes it; the row for kassensitzung-eroeffnet:v1 already
// exists because kassenjournal has an FK to kassensitzungen and EroeffneKassensitzung
// inserts it first.
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
		// Other kassensitzung events don't change the CRUD entity.
	}

	return nil
}

func (r Repository) handleTischSessionEvent(ctx context.Context, qtx *dbgen.Queries, e event.Event, kassensitzungNr int) error {
	tischID, err := kasse.ParseTischIDFromSubject(e.Subject)
	if err != nil {
		return fmt.Errorf("parse tisch ID from subject %q: %w", e.Subject, err)
	}

	currentState, err := getTischSessionInTx(ctx, qtx, e.Subject)
	if err != nil {
		return err
	}

	newState, err := kasse.ApplyEvent(currentState, e)
	if err != nil {
		return fmt.Errorf("apply event to tisch session: %w", err)
	}

	return upsertTischSessionState(ctx, qtx, e.Subject, tischID, kassensitzungNr, newState)
}

func upsertTischSessionState(ctx context.Context, qtx *dbgen.Queries, subject string, tischID, kassensitzungNr int, state kasse.TischSession) error {
	unbezahltJSON, err := json.Marshal(state.UnbezahltePositionen)
	if err != nil {
		return fmt.Errorf("marshal unbezahlte positionen: %w", err)
	}

	err = qtx.UpsertTischSession(ctx, dbgen.UpsertTischSessionParams{
		Subject:                subject,
		TischID:                tischID,
		KassensitzungNr:        kassensitzungNr,
		SaldoCents:             state.SaldoCents,
		UnbezahltePositionen:   unbezahltJSON,
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

// ReadTischSession returns a zero-value TischSession when no entry exists.
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

type TischNameUndSession struct {
	Name    string
	Session kasse.TischSession
}

// ReadFavoritenTischStates liefert Name und projizierte Session je Tisch-ID in einer
// Query. Ein Favorit ohne Session erhält eine Null-TischSession; eine gelöschte oder
// unbekannte Tisch-ID fehlt in der Map.
func (r Repository) ReadFavoritenTischStates(ctx context.Context, tischIDs []int, kassensitzungNr int) (map[int]TischNameUndSession, error) {
	if len(tischIDs) == 0 {
		return make(map[int]TischNameUndSession), nil
	}

	ids32 := make([]int32, len(tischIDs))
	for i, id := range tischIDs {
		ids32[i] = int32(id) //nolint:gosec // Tisch-IDs sind positive Entity-IDs
	}

	const query = `SELECT t.id, t.name,
			ts.subject, ts.kassensitzung_nr, ts.saldo_cents,
			ts.unbezahlte_positionen, ts.gesamt_zahlungen_cents,
			ts.erste_bestellung_logtime, ts.last_event_id, ts.last_event_version
		FROM tische t
		LEFT JOIN tisch_sessions ts ON ts.tisch_id = t.id AND ts.kassensitzung_nr = $2
		WHERE t.id = ANY($1) AND t.status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, ids32, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close() //nolint:errcheck // expliziter Close mit Fehlerprüfung unten

	result := make(map[int]TischNameUndSession, len(tischIDs))
	for rows.Next() {
		var (
			id               int
			name             string
			subject          sql.NullString
			sessionKsNr      sql.NullInt64
			saldoCents       sql.NullInt64
			unbezahlt        []byte
			gesamtZahlungen  sql.NullInt64
			ersteBestellung  sql.NullTime
			lastEventID      sql.NullInt64
			lastEventVersion sql.NullInt64
		)
		if err := rows.Scan(&id, &name, &subject, &sessionKsNr, &saldoCents, &unbezahlt, &gesamtZahlungen, &ersteBestellung, &lastEventID, &lastEventVersion); err != nil {
			return nil, db.Error(err)
		}

		var session kasse.TischSession
		if subject.Valid {
			session, err = toTischSession(dbgen.TischSession{
				Subject:                subject.String,
				TischID:                id,
				KassensitzungNr:        int(sessionKsNr.Int64),
				SaldoCents:             int(saldoCents.Int64),
				UnbezahltePositionen:   unbezahlt,
				GesamtZahlungenCents:   int(gesamtZahlungen.Int64),
				ErsteBestellungLogtime: ersteBestellung,
				LastEventID:            int(lastEventID.Int64),
				LastEventVersion:       int(lastEventVersion.Int64),
			})
			if err != nil {
				return nil, err
			}
		}

		result[id] = TischNameUndSession{Name: name, Session: session}
	}
	if err := rows.Close(); err != nil {
		return nil, db.Error(err)
	}
	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return result, nil
}

// getTischSessionInTx returns a zero-value TischSession when no entry exists.
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

// eventFromReadRow baut ein Event aus einer Kassenjournal-Zeile.
// ReadEventsBySubjectRow, ReadDirektverkaufEventsRow und
// ReadKassensitzungEventsRow sind feldgleich (dieselben sqlc-Query-Spalten),
// deshalb konvertiert jeder Aufrufer seine Zeile per Typkonvertierung auf
// ReadEventsBySubjectRow.
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

// eventSignaturFromKassensitzungRow: der processType steht immer, die Signaturspalten
// erst nach der Quittierung.
func eventSignaturFromKassensitzungRow(row dbgen.ReadEventsByKassensitzungRow) tse.EventSignatur {
	signatur := tse.EventSignatur{ProcessType: row.ProcessType.String}
	if row.Signatur.Valid {
		signatur.Signatur = &tse.Signatur{
			TransaktionNummer: int(row.TransaktionNummer.Int64),
			SignaturZaehler:   int(row.SignaturZaehler.Int64),
			TSESeriennummer:   row.TseSeriennummer.String,
			LogTimeStart:      row.LogTimeStart.Time,
			LogTimeEnd:        row.LogTimeEnd.Time,
			Signatur:          row.Signatur.String,
			QRCodeData:        row.QrCodeData.String,
		}
	}
	return signatur
}

// ReadEventsBySubject returns the subject's events ordered by ID ascending.
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

// ReadDirektverkaufEvents returns the Kassensitzung's Direktverkauf events ordered by ID ascending.
func (r Repository) ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error) {
	rows, err := r.q.ReadDirektverkaufEvents(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for i := range rows {
		events = append(events, eventFromReadRow(dbgen.ReadEventsBySubjectRow(rows[i])))
	}

	return events, nil
}

// ReadEventsByKassensitzung returns all events of the Kassensitzung ordered by ID
// ascending, each with its Signaturauftrag-Stand (kein Auftrag = nicht
// signaturpflichtig). Read side of the DSFinV-K export.
func (r Repository) ReadEventsByKassensitzung(ctx context.Context, kassensitzungNr int) ([]event.Event, map[int]tse.EventSignatur, error) {
	rows, err := r.q.ReadEventsByKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		return nil, nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	signaturen := make(map[int]tse.EventSignatur)
	for i := range rows {
		events = append(events, eventFromKassensitzungRow(rows[i]))
		if rows[i].ProcessType.Valid {
			signaturen[rows[i].ID] = eventSignaturFromKassensitzungRow(rows[i])
		}
	}

	return events, signaturen, nil
}

// ReadKassensitzungEvents returns all events of the Kassensitzung ordered by ID
// ascending, without the Signaturauftrag JOIN — only the DSFinV-K export
// (ReadEventsByKassensitzung) needs it.
func (r Repository) ReadKassensitzungEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error) {
	rows, err := r.q.ReadKassensitzungEvents(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	events := make([]event.Event, 0, len(rows))
	for i := range rows {
		events = append(events, eventFromReadRow(dbgen.ReadEventsBySubjectRow(rows[i])))
	}

	return events, nil
}

// GetMaxVersion returns 0 if the subject has no events.
func (r Repository) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	version, err := r.q.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, db.Error(err)
	}

	return version, nil
}

// EventExistsByTypeAndVorgangsID wird nur auf dem Fehler-Pfad (nach UniqueViolation)
// aufgerufen, um idempotente Einreichung von echtem OCC-Konflikt zu unterscheiden.
func (r Repository) EventExistsByTypeAndVorgangsID(ctx context.Context, eventType, vorgangsID, jsonKey string) (bool, error) {
	var dummy int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM kassenjournal WHERE type = $1 AND data->>$2 = $3 LIMIT 1`,
		eventType, jsonKey, vorgangsID,
	).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, db.Error(err)
	}
	return true, nil
}

// RebuildAllProjections rebuilds the tisch_sessions projection from scratch in one
// transaction and returns the number of subjects rebuilt. kassensitzungen is a CRUD
// entity and is NOT replayed.
func (r Repository) RebuildAllProjections(ctx context.Context) (int, error) {
	rebuiltCount := 0
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		if err := qtx.DeleteAllTischSession(ctx); err != nil {
			return fmt.Errorf("delete all tisch sessions: %w", err)
		}

		subjects, err := qtx.GetDistinctTischSessionSubjects(ctx)
		if err != nil {
			return fmt.Errorf("get distinct tisch-session subjects: %w", err)
		}

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

func (r Repository) GetKassenbestand(ctx context.Context, kassensitzungNr int) (kasse.Kassenbestand, error) {
	row, err := r.q.GetKassenbestand(ctx, kassensitzungNr)
	if err != nil {
		return kasse.Kassenbestand{}, db.Error(err)
	}
	return kasse.Kassenbestand{
		SollBestandCents:    row.SollBestandCents,
		AnfangsbestandCents: row.AnfangsbestandCents,
		BareinnahmenCents:   row.BareinnahmenCents,
		EinlagenCents:       row.EinlagenCents,
		EntnahmenCents:      row.EntnahmenCents,
	}, nil
}

// GetGeldtransitListe returns the Kassensitzung's Geldbewegungen, newest first.
func (r Repository) GetGeldtransitListe(ctx context.Context, kassensitzungNr int) ([]kasse.Geldtransit, error) {
	rows, err := r.q.GetGeldtransitListe(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	buchungen := make([]kasse.Geldtransit, 0, len(rows))
	for i := range rows {
		buchungen = append(buchungen, kasse.Geldtransit{
			Zeitpunkt:   rows[i].Zeitpunkt,
			Richtung:    rows[i].Richtung,
			BetragCents: rows[i].BetragCents,
			Kommentar:   rows[i].Kommentar,
			GebuchtVon:  rows[i].GebuchtVon,
		})
	}

	return buchungen, nil
}

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
