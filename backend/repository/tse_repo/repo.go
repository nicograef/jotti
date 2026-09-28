package tse_repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// MaxSignaturVersuche is the number of order-specific failures (tse.AuftragsFehler) that make a Signaturauftrag fehlgeschlagen.
// They are near-deterministic, so the 5 s/15 s backoff ends a poison order before tse.RueckstandSchwelle opens a Rückstand;
// TSE-wide failures never count here (docs/handbuch.md §3.13).
const MaxSignaturVersuche = 3

// OffenerSignaturauftrag is the worker's view of a due order.
type OffenerSignaturauftrag struct {
	ID          int
	TxID        string
	ProcessType string
	ProcessData string
}

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

func (r Repository) GetOffeneTSESignaturauftraege(ctx context.Context, limit int) ([]OffenerSignaturauftrag, error) {
	rows, err := r.q.GetOffeneTSESignaturauftraege(ctx, int32(limit))
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]OffenerSignaturauftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, OffenerSignaturauftrag{
			ID:          row.ID,
			TxID:        row.TxID,
			ProcessType: row.ProcessType,
			ProcessData: row.ProcessData,
		})
	}

	return result, nil
}

// QuittiereTSESignaturauftrag writes the signature onto the order in a single update and sets it erledigt.
// The status guard (offen) makes it idempotent: the signature columns are written exactly once.
func (r Repository) QuittiereTSESignaturauftrag(ctx context.Context, auftragID int, signatur tse.Signatur) error {
	return db.Error(r.q.QuittiereTSESignaturauftrag(ctx, dbgen.QuittiereTSESignaturauftragParams{
		ID:                auftragID,
		TransaktionNummer: sql.NullInt64{Int64: int64(signatur.TransaktionNummer), Valid: true},
		SignaturZaehler:   sql.NullInt64{Int64: int64(signatur.SignaturZaehler), Valid: true},
		TseSeriennummer:   sql.NullString{String: strings.TrimSpace(signatur.TSESeriennummer), Valid: true},
		LogTimeStart:      sql.NullTime{Time: signatur.LogTimeStart.UTC(), Valid: true},
		LogTimeEnd:        sql.NullTime{Time: signatur.LogTimeEnd.UTC(), Valid: true},
		Signatur:          sql.NullString{String: strings.TrimSpace(signatur.Signatur), Valid: true},
		QrCodeData:        sql.NullString{String: strings.TrimSpace(signatur.QRCodeData), Valid: true},
	}))
}

// TSESignaturauftragFehlversuch records an order-specific failure and schedules the next attempt with backoff.
// The MaxSignaturVersuche-th failure makes the order fehlgeschlagen; the backoff logic lives in the SQL query.
func (r Repository) TSESignaturauftragFehlversuch(ctx context.Context, auftragID int, fehler string) error {
	return db.Error(r.q.TSESignaturauftragFehlversuch(ctx, dbgen.TSESignaturauftragFehlversuchParams{
		ID:            auftragID,
		LetzterFehler: sql.NullString{String: fehler, Valid: true},
		MaxVersuche:   MaxSignaturVersuche,
	}))
}

// MarkOffeneAlsNichtKonfiguriert finally marks all open orders tse_nicht_konfiguriert and returns their count.
// Without a TSE configuration there is no signature, so neither retries nor later re-signing apply.
func (r Repository) MarkOffeneAlsNichtKonfiguriert(ctx context.Context) (int64, error) {
	n, err := r.q.MarkOffeneTSESignaturauftraegeNichtKonfiguriert(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return n, nil
}

// GetTSESignaturQueueZustand computes the signature queue state for admin monitoring on demand.
func (r Repository) GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error) {
	row, err := r.q.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		return tse.SignaturQueueZustand{}, db.Error(err)
	}
	return tse.SignaturQueueZustand{
		OffeneAuftraege:          row.OffeneAuftraege,
		FehlgeschlageneAuftraege: row.FehlgeschlageneAuftraege,
		LetzterFehler:            row.LetzterFehler,
		RueckstandSekunden:       row.RueckstandSekunden,
		SignaturenProMinute:      row.SignaturenProMinute,
		SignierdauerP95Sekunden:  row.SignierdauerP95Sekunden,
	}, nil
}

// GetAlleTSEStoerungen returns the Störungsprotokoll: the latest 200 Störungszeiträume, newest first.
func (r Repository) GetAlleTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error) {
	rows, err := r.q.GetAlleTSEStoerungen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]tse.Stoerungszeitraum, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		zeitraum := tse.Stoerungszeitraum{
			ID:         row.ID,
			Beginn:     row.Beginn,
			GrundArt:   row.GrundArt,
			Fehlertext: row.Fehlertext,
		}
		if row.Ende.Valid {
			ende := row.Ende.Time
			zeitraum.Ende = &ende
		}
		result = append(result, zeitraum)
	}

	return result, nil
}

// GetSignaturauftragZuEvent returns an event's signature state for the Beleg.
// db.ErrNotFound means no order: the event is not signaturpflichtig.
func (r Repository) GetSignaturauftragZuEvent(ctx context.Context, eventID int) (tse.SignaturauftragStand, error) {
	row, err := r.q.GetTSESignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tse.SignaturauftragStand{}, db.ErrNotFound
		}
		return tse.SignaturauftragStand{}, db.Error(err)
	}

	stand := tse.SignaturauftragStand{Status: row.Status, ErstelltAm: row.ErstelltAm}
	if row.Status == tse.StatusErledigt {
		stand.Signatur = &tse.Signatur{
			TransaktionNummer: int(row.TransaktionNummer.Int64),
			SignaturZaehler:   int(row.SignaturZaehler.Int64),
			TSESeriennummer:   row.TseSeriennummer.String,
			LogTimeStart:      row.LogTimeStart.Time,
			LogTimeEnd:        row.LogTimeEnd.Time,
			Signatur:          row.Signatur.String,
			QRCodeData:        row.QrCodeData.String,
		}
	}
	return stand, nil
}

// GetOffeneSignaturauftragStaendeFuerKassensitzung returns the states of the session's orders not yet erledigt for the Kassenabschluss gate.
// The gate classifies them via DetermineSignaturstatus as ausstehend or Ausfall.
func (r Repository) GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx context.Context, kassensitzungNr int) ([]tse.SignaturauftragStand, error) {
	rows, err := r.q.GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]tse.SignaturauftragStand, 0, len(rows))
	for _, row := range rows {
		result = append(result, tse.SignaturauftragStand{Status: row.Status, ErstelltAm: row.ErstelltAm})
	}
	return result, nil
}

// GetAeltesterOffenerTSESignaturauftrag returns the creation time of the oldest open order, or nil if none is open.
// The Rückstand watchdog measures the backlog from it.
func (r Repository) GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error) {
	erstelltAm, err := r.q.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}
	return &erstelltAm, nil
}

// OpenTSEStoerung opens a Störungszeitraum; it is a no-op while any Störungszeitraum is active.
// A partial unique index enforces at most one active Störungszeitraum.
func (r Repository) OpenTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error {
	return db.Error(r.q.OpenTSEStoerung(ctx, dbgen.OpenTSEStoerungParams{
		GrundArt:   grundArt,
		Fehlertext: fehlertext,
	}))
}

// CloseTSEStoerung ends the active Störungszeitraum of grundArt and is a no-op if none is active.
// Each writer closes only its own Grund-Art.
func (r Repository) CloseTSEStoerung(ctx context.Context, grundArt string) error {
	return db.Error(r.q.CloseTSEStoerung(ctx, grundArt))
}

// GetAktiveTSEStoerung returns the active Störungszeitraum, or nil if none is active.
func (r Repository) GetAktiveTSEStoerung(ctx context.Context) (*tse.Stoerung, error) {
	row, err := r.q.GetAktiveTSEStoerung(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}
	return &tse.Stoerung{Beginn: row.Beginn, GrundArt: row.GrundArt, Fehlertext: row.Fehlertext}, nil
}
