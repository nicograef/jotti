package tse_repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// MaxNachsignierVersuche ist die Anzahl Fehlversuche, nach der ein
// Nachsignier-Auftrag als fehlgeschlagen markiert und nicht mehr automatisch
// versucht wird. Mit dem exponentiellen Backoff (1, 2, 4, ... Minuten,
// gedeckelt auf 30) ueberbrueckt der Worker damit Ausfaelle von rund drei
// Stunden, bevor ein Admin eingreifen muss.
const MaxNachsignierVersuche = 10

type OffenerNachsignierAuftrag struct {
	ID          int
	TxID        string
	ProcessType string
	ProcessData string
}

// NachsignierAuftrag ist die Admin-Sicht eines Nachsignier-Auftrags. Sie dient
// zugleich als TSE-Ausfalldokumentation (AEAO zu § 146a, 1.14.1):
// ErstelltAm = Beginn, ErledigtAm = Ende, LetzterFehler = Grund.
type NachsignierAuftrag struct {
	ID            int
	TxID          string
	ProcessType   string
	Status        string
	Versuche      int
	LetzterFehler string
	ErstelltAm    time.Time
	ErledigtAm    *time.Time
}

type Signatur struct {
	TxID              string
	TransaktionNummer int
	SignaturZaehler   int
	TSESeriennummer   string
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	Signatur          string
	QRCodeData        string
}

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

func (r Repository) GetOffeneTSENachsignierAuftraege(ctx context.Context, limit int) ([]OffenerNachsignierAuftrag, error) {
	rows, err := r.q.GetOffeneTSENachsignierAuftraege(ctx, int32(limit))
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]OffenerNachsignierAuftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, OffenerNachsignierAuftrag{
			ID:          row.ID,
			TxID:        row.TxID,
			ProcessType: row.ProcessType,
			ProcessData: row.ProcessData,
		})
	}

	return result, nil
}

func (r Repository) QuittiereTSENachsignierAuftrag(ctx context.Context, auftragID int, signatur Signatur) error {
	return r.withTx(ctx, func(qtx *dbgen.Queries) error {
		err := qtx.UpsertTSESignatur(ctx, dbgen.UpsertTSESignaturParams{
			TxID:              strings.TrimSpace(signatur.TxID),
			TransaktionNummer: signatur.TransaktionNummer,
			SignaturZaehler:   signatur.SignaturZaehler,
			TseSeriennummer:   strings.TrimSpace(signatur.TSESeriennummer),
			LogTimeStart:      signatur.LogTimeStart.UTC(),
			LogTimeEnd:        signatur.LogTimeEnd.UTC(),
			Signatur:          strings.TrimSpace(signatur.Signatur),
			QrCodeData:        strings.TrimSpace(signatur.QRCodeData),
		})
		if err != nil {
			return db.Error(err)
		}

		err = qtx.MarkTSENachsignierAuftragErledigt(ctx, auftragID)
		if err != nil {
			return db.Error(err)
		}

		return nil
	})
}

// TSENachsignierAuftragFehlversuch verbucht einen Fehlversuch: Zaehler hoch,
// Fehlertext speichern, naechster Versuch mit exponentiellem Backoff. Beim
// MaxNachsignierVersuche-ten Fehlversuch wechselt der Auftrag auf
// fehlgeschlagen (Backoff-Logik liegt in der SQL-Query).
func (r Repository) TSENachsignierAuftragFehlversuch(ctx context.Context, auftragID int, fehler string) error {
	return db.Error(r.q.TSENachsignierAuftragFehlversuch(ctx, dbgen.TSENachsignierAuftragFehlversuchParams{
		ID:            auftragID,
		LetzterFehler: sql.NullString{String: fehler, Valid: true},
		MaxVersuche:   MaxNachsignierVersuche,
	}))
}

// GetTSENachsignierAuftraege liefert die Nachsignier-Auftraege fuer die
// Admin-Verwaltung und die Ausfalldokumentation, neueste zuerst.
func (r Repository) GetTSENachsignierAuftraege(ctx context.Context) ([]NachsignierAuftrag, error) {
	rows, err := r.q.GetTSENachsignierAuftraege(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]NachsignierAuftrag, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		auftrag := NachsignierAuftrag{
			ID:            row.ID,
			TxID:          row.TxID,
			ProcessType:   row.ProcessType,
			Status:        row.Status,
			Versuche:      row.Versuche,
			LetzterFehler: row.LetzterFehler.String,
			ErstelltAm:    row.ErstelltAm,
		}
		if row.ErledigtAm.Valid {
			erledigtAm := row.ErledigtAm.Time
			auftrag.ErledigtAm = &erledigtAm
		}
		result = append(result, auftrag)
	}

	return result, nil
}

// TSENachsignierAuftragZuruecksetzen reiht einen fehlgeschlagenen Auftrag
// wieder ein (fehlgeschlagen -> offen, Zaehler und Fehler zurueckgesetzt).
// Der Status-Guard wirkt nur auf fehlgeschlagene Auftraege.
func (r Repository) TSENachsignierAuftragZuruecksetzen(ctx context.Context, auftragID int) error {
	return db.Error(r.q.TSENachsignierAuftragZuruecksetzen(ctx, auftragID))
}

// TSENachsignierAuftragVerwerfen markiert einen fehlgeschlagenen Auftrag als
// verworfen. Der Eintrag bleibt fuer die Ausfalldokumentation erhalten; der
// Status-Guard wirkt nur auf fehlgeschlagene Auftraege.
func (r Repository) TSENachsignierAuftragVerwerfen(ctx context.Context, auftragID int) error {
	return db.Error(r.q.TSENachsignierAuftragVerwerfen(ctx, auftragID))
}

func (r Repository) CountOffeneTSENachsignierAuftraege(ctx context.Context) (int, error) {
	count, err := r.q.CountOffeneTSENachsignierAuftraege(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return count, nil
}

func (r Repository) GetTSESignaturByTxID(ctx context.Context, txID string) (Signatur, error) {
	row, err := r.q.GetTSESignaturByTxID(ctx, strings.TrimSpace(txID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Signatur{}, db.ErrNotFound
		}
		return Signatur{}, db.Error(err)
	}

	return Signatur{
		TxID:              row.TxID,
		TransaktionNummer: row.TransaktionNummer,
		SignaturZaehler:   row.SignaturZaehler,
		TSESeriennummer:   row.TseSeriennummer,
		LogTimeStart:      row.LogTimeStart,
		LogTimeEnd:        row.LogTimeEnd,
		Signatur:          row.Signatur,
		QRCodeData:        row.QrCodeData,
	}, nil
}
