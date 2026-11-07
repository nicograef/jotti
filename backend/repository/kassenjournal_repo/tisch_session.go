package kassenjournal_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

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
	return getTischSession(ctx, r.q, subject)
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

	const query = `SELECT t.id, t.name,
			ts.subject, ts.kassensitzung_nr, ts.saldo_cents,
			ts.unbezahlte_positionen, ts.gesamt_zahlungen_cents,
			ts.erste_bestellung_logtime, ts.last_event_id, ts.last_event_version
		FROM tische t
		LEFT JOIN tisch_sessions ts ON ts.tisch_id = t.id AND ts.kassensitzung_nr = $2
		WHERE t.id = ANY($1) AND t.status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, db.Int4Array(tischIDs), kassensitzungNr)
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

// getTischSession returns a zero-value TischSession when no entry exists; q is the
// repository's queries or those of the write transaction.
func getTischSession(ctx context.Context, q *dbgen.Queries, subject string) (kasse.TischSession, error) {
	row, err := q.GetTischSession(ctx, subject)
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
