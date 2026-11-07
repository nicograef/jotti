package kassenjournal_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// WriteEvent stores the event and, in the same transaction, updates the entity or
// projection streamType routes to (Direktverkauf: none).
func (r Repository) WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	ids, err := r.writeAtomic(ctx, kassensitzungNr, []NeuesEvent{{Event: e, StreamType: streamType}})
	if err != nil {
		return 0, err
	}
	return ids[0], nil
}

// NeuesEvent is one event of an atomic multi-event write. Event carries its final subject and version;
// a non-nil BuildAuftraege derives print jobs from the stored event, generated ID included, in the
// same transaction (transactional outbox).
type NeuesEvent struct {
	Event          event.Event
	StreamType     kasse.StreamType
	BuildAuftraege func(event.Event) []druckstation.NeuerDruckauftrag
}

// WriteEventsAtomic writes the given events across stream types all-or-nothing and triggers the
// Signatur-Worker once after the commit.
func (r Repository) WriteEventsAtomic(ctx context.Context, kassensitzungNr int, events []NeuesEvent) error {
	_, err := r.writeAtomic(ctx, kassensitzungNr, events)
	return err
}

// writeAtomic returns the stored event IDs in the order of events.
func (r Repository) writeAtomic(ctx context.Context, kassensitzungNr int, events []NeuesEvent) ([]int, error) {
	ids := make([]int, 0, len(events))
	eingereiht := false
	err := db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		for _, neu := range events {
			stored, auftragEingereiht, err := r.writeEventInTx(ctx, qtx, neu.Event, neu.StreamType, kassensitzungNr)
			if err != nil {
				return err
			}
			eingereiht = eingereiht || auftragEingereiht

			if neu.BuildAuftraege != nil {
				if err := druckauftrag_repo.InsertDruckauftraege(ctx, qtx, neu.BuildAuftraege(stored)); err != nil {
					return err
				}
			}
			ids = append(ids, stored.ID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	notifySignaturWorker(eingereiht)
	return ids, nil
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

// notifySignaturWorker stößt den Worker non-blocking an; der Polling-Tick bleibt Fallback.
func notifySignaturWorker(auftragEingereiht bool) {
	if auftragEingereiht {
		tse_repo.NotifySignaturWorker()
	}
}

// writeEventInTx reports whether a Signaturauftrag was enqueued.
// The caller owns the transaction and triggers the Signatur-Worker after the commit.
func (r Repository) writeEventInTx(ctx context.Context, qtx *dbgen.Queries, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (event.Event, bool, error) {
	// FOR SHARE serialises against the 'wird_abgeschlossen' barrier: a write commits before it or fails.
	// See docs/handbuch.md §3.7 (Kassensitzung-Invarianten).
	status, err := qtx.GetKassensitzungStatusForShare(ctx, kassensitzungNr)
	if err != nil {
		return event.Event{}, false, db.Error(err)
	}
	switch kasse.KassensitzungStatus(status) {
	case kasse.KassensitzungOffen:
	case kasse.KassensitzungWirdAbgeschlossen:
		if !kasse.IsAbschlussEventType(e.Type) {
			return event.Event{}, false, ErrKassensitzungNichtOffen
		}
	case kasse.KassensitzungAbgeschlossen:
		fallthrough
	default:
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

	case kasse.StreamTypeDirektverkauf, kasse.StreamTypePfandrueckgabe:
		// Direktverkauf and Pfandrückgabe live entirely in the kassenjournal — no projection to update.

	default:
		return event.Event{}, false, fmt.Errorf("unknown stream type: %s", streamType)
	}

	return e, signaturpflichtig, nil
}

// handleKassensitzungEvent handles only tagesabschluss-erstellt:v1: EroeffneKassensitzung inserts the row
// before kassensitzung-eroeffnet:v1 because kassenjournal has an FK to kassensitzungen.
func (r Repository) handleKassensitzungEvent(ctx context.Context, qtx *dbgen.Queries, e event.Event, kassensitzungNr int) error {
	switch kasse.EventType(e.Type) {
	case kasse.EventTypeTagesabschlussErstelltV1:
		err := qtx.UpdateKassensitzung(ctx, dbgen.UpdateKassensitzungParams{
			ZNr:    kassensitzungNr,
			Status: string(kasse.KassensitzungAbgeschlossen),
		})
		if err != nil {
			return db.Error(err)
		}

	case kasse.EventTypeKassensitzungEroeffnetV1, kasse.EventTypeGeldtransitGebuchtV1,
		kasse.EventTypeKassensturzDurchgefuehrtV1, kasse.EventTypeDifferenzSollIstGebuchtV1:
		// These kassensitzung events don't change the CRUD entity.

	case kasse.EventTypeBestellungAufgenommenV1, kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2,
		kasse.EventTypeStornierungErteiltV1, kasse.EventTypeBestellungKorrigiertV1, kasse.EventTypeBestellungUmgebuchtV1,
		kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2, kasse.EventTypeDirektverkaufStorniertV1,
		kasse.EventTypePfandZurueckgegebenV1, kasse.EventTypePfandrueckgabeStorniertV1:
		return fmt.Errorf("event type %s does not belong to the kassensitzung stream", e.Type)

	default:
		return fmt.Errorf("unknown event type: %s", e.Type)
	}

	return nil
}

func (r Repository) handleTischSessionEvent(ctx context.Context, qtx *dbgen.Queries, e event.Event, kassensitzungNr int) error {
	tischID, err := kasse.ParseTischIDFromSubject(e.Subject)
	if err != nil {
		return fmt.Errorf("parse tisch ID from subject %q: %w", e.Subject, err)
	}

	currentState, err := getTischSession(ctx, qtx, e.Subject)
	if err != nil {
		return err
	}

	newState, err := kasse.ApplyEvent(currentState, e)
	if err != nil {
		return fmt.Errorf("apply event to tisch session: %w", err)
	}

	return upsertTischSessionState(ctx, qtx, e.Subject, tischID, kassensitzungNr, newState)
}
