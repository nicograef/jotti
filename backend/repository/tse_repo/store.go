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

type OffenerNachsignierAuftrag struct {
	ID          int
	TxID        string
	ProcessType string
	ProcessData string
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

type Store struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewStore(database *sql.DB) Store {
	return Store{DB: database, q: dbgen.New(database)}
}

func (r Store) GetOffeneTSENachsignierAuftraege(ctx context.Context, limit int) ([]OffenerNachsignierAuftrag, error) {
	rows, err := r.q.GetOffeneTSENachsignierAuftraege(ctx, int32(limit))
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]OffenerNachsignierAuftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, OffenerNachsignierAuftrag{
			ID:          int(row.ID),
			TxID:        row.TxID,
			ProcessType: row.ProcessType,
			ProcessData: row.ProcessData,
		})
	}

	return result, nil
}

func (r Store) QuittiereTSENachsignierAuftrag(ctx context.Context, auftragID int, signatur Signatur) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := r.q.WithTx(tx)
	err = qtx.UpsertTSESignatur(ctx, dbgen.UpsertTSESignaturParams{
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

	err = qtx.MarkTSENachsignierAuftragErledigt(ctx, int32(auftragID))
	if err != nil {
		return db.Error(err)
	}

	if err := tx.Commit(); err != nil {
		return db.Error(err)
	}

	return nil
}

func (r Store) CountOffeneTSENachsignierAuftraege(ctx context.Context) (int, error) {
	count, err := r.q.CountOffeneTSENachsignierAuftraege(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return count, nil
}

func (r Store) GetTSESignaturByTxID(ctx context.Context, txID string) (Signatur, error) {
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
