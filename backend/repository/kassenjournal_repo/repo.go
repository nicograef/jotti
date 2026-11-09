package kassenjournal_repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// ErrKassensitzungNichtOffen: der Write traf auf eine nicht (mehr) offene
// Kassensitzung; die Application-Schicht mappt das auf HTTP 409.
var ErrKassensitzungNichtOffen = errors.New("kassensitzung ist nicht offen")

// ErrKassensitzungWirdAbgeschlossen: der Write traf auf die Abschluss-Barriere
// ('wird_abgeschlossen'); nur Abschluss-Events passieren sie.
var ErrKassensitzungWirdAbgeschlossen = errors.New("kassensitzung wird abgeschlossen")

// ErrTischNichtAktiv refuses an event that raises the saldo of a Tisch that is no longer active.
var ErrTischNichtAktiv = errors.New("tisch ist nicht aktiv")

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

// eventFromReadRow also serves ReadDirektverkaufEventsRow and ReadKassensitzungEventsRow:
// they share the sqlc columns, so callers convert their row to ReadEventsBySubjectRow.
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

// eventSignaturFromKassensitzungRow: processType und Status stehen immer, die Signaturspalten
// erst nach der Quittierung.
func eventSignaturFromKassensitzungRow(row dbgen.ReadEventsByKassensitzungRow) tse.EventSignatur {
	signatur := tse.EventSignatur{ProcessType: row.ProcessType.String, Status: row.Status.String}
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

// GetDirektverkauf returns the sale event of verkaufID from any Kassensitzung; false means none is
// stored.
func (r Repository) GetDirektverkauf(ctx context.Context, verkaufID string) (event.Event, bool, error) {
	row, err := r.q.GetDirektverkauf(ctx, verkaufID)
	if errors.Is(err, sql.ErrNoRows) {
		return event.Event{}, false, nil
	}
	if err != nil {
		return event.Event{}, false, db.Error(err)
	}
	return eventFromReadRow(dbgen.ReadEventsBySubjectRow(row)), true, nil
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
