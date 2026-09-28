package tse_repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// GetKassenidentitaet returns the kasse identity, or db.ErrNotFound if not yet initialized.
func (r Repository) GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error) {
	row, err := r.q.GetKassenidentitaet(ctx)
	if err != nil {
		return tse.Kassenidentitaet{}, db.Error(err)
	}
	return tse.Kassenidentitaet{
		Seriennummer: row.Seriennummer,
		AngelegtAm:   row.AngelegtAm,
	}, nil
}

func (r Repository) GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error) {
	row, err := r.q.GetTSEKonfiguration(ctx)
	if err != nil {
		return tse.Konfiguration{}, db.Error(err)
	}
	return toTSEKonfiguration(row), nil
}

// SaveEinrichtung stores the TSE configuration for every write path (setup, takeover, credentials change, clearing).
// Only the transition from unconfigured to configured also marks the open orders from the unconfigured period
// tse_nicht_konfiguriert and closes the keine_konfiguration Störungszeitraum, in the same transaction (docs/handbuch.md §3.13).
func (r Repository) SaveEinrichtung(ctx context.Context, c tse.Konfiguration) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		warKonfiguriert := false
		if vorher, err := qtx.GetTSEKonfiguration(ctx); err == nil {
			warKonfiguriert = toTSEKonfiguration(vorher).IstKonfiguriert()
		} else if !errors.Is(err, sql.ErrNoRows) {
			return db.Error(err)
		}

		if err := qtx.UpsertTSEKonfiguration(ctx, upsertTSEKonfigurationParams(c)); err != nil {
			return db.Error(err)
		}

		if warKonfiguriert || !c.IstKonfiguriert() {
			return nil
		}
		if _, err := qtx.MarkOffeneTSESignaturauftraegeNichtKonfiguriert(ctx); err != nil {
			return db.Error(err)
		}
		if err := qtx.CloseTSEStoerung(ctx, tse.StoerungGrundKeineKonfiguration); err != nil {
			return db.Error(err)
		}
		return nil
	})
}

func upsertTSEKonfigurationParams(c tse.Konfiguration) dbgen.UpsertTSEKonfigurationParams {
	return dbgen.UpsertTSEKonfigurationParams{
		ApiKey:    c.ApiKey,
		ApiSecret: c.ApiSecret,
		TssID:     c.TssID,
		ClientID:  c.ClientID,
	}
}

// GetTSEStammdaten reads the singleton TSS master data for the DSFinV-K export; the fields are empty before TSE setup.
func (r Repository) GetTSEStammdaten(ctx context.Context) (tse.Stammdaten, error) {
	row, err := r.q.GetTSEStammdaten(ctx)
	if err != nil {
		return tse.Stammdaten{}, db.Error(err)
	}
	return tse.Stammdaten{
		Seriennummer:        row.Seriennummer,
		SignaturAlgorithmus: row.SignaturAlgorithmus,
		PublicKey:           row.PublicKey,
		Zertifikat:          row.Zertifikat,
		LogTimeFormat:       row.LogTimeFormat,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// UpsertTSEStammdaten stores the singleton TSS master data for the DSFinV-K export.
func (r Repository) UpsertTSEStammdaten(ctx context.Context, s tse.Stammdaten) error {
	err := r.q.UpsertTSEStammdaten(ctx, dbgen.UpsertTSEStammdatenParams{
		Seriennummer:        s.Seriennummer,
		SignaturAlgorithmus: s.SignaturAlgorithmus,
		PublicKey:           s.PublicKey,
		Zertifikat:          s.Zertifikat,
		LogTimeFormat:       s.LogTimeFormat,
	})
	if err != nil {
		return db.Error(err)
	}
	return nil
}

func toTSEKonfiguration(row dbgen.GetTSEKonfigurationRow) tse.Konfiguration {
	return tse.Konfiguration{
		ApiKey:    row.ApiKey,
		ApiSecret: row.ApiSecret,
		TssID:     row.TssID,
		ClientID:  row.ClientID,
		UpdatedAt: row.UpdatedAt,
	}
}
