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
		return saveEinrichtung(ctx, qtx, c)
	})
}

// SaveEinrichtungMitStammdaten stores the configuration and the master data of its TSS in one
// transaction, so the current Stammdaten never name a TSS the configuration does not.
func (r Repository) SaveEinrichtungMitStammdaten(ctx context.Context, c tse.Konfiguration, s tse.Stammdaten) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		if err := upsertStammdaten(ctx, qtx, s); err != nil {
			return err
		}
		return saveEinrichtung(ctx, qtx, c)
	})
}

func saveEinrichtung(ctx context.Context, qtx *dbgen.Queries, c tse.Konfiguration) error {
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
}

func upsertTSEKonfigurationParams(c tse.Konfiguration) dbgen.UpsertTSEKonfigurationParams {
	return dbgen.UpsertTSEKonfigurationParams{
		ApiKey:    c.ApiKey,
		ApiSecret: c.ApiSecret,
		TssID:     c.TssID,
		ClientID:  c.ClientID,
	}
}

// GetTSEStammdaten reads the master data of the TSS configured now; the fields are empty before TSE setup.
func (r Repository) GetTSEStammdaten(ctx context.Context) (tse.Stammdaten, error) {
	row, err := r.q.GetTSEStammdaten(ctx)
	if err != nil {
		return tse.Stammdaten{}, db.Error(err)
	}
	return tse.Stammdaten(row), nil
}

// GetTSEStammdatenJeTSS maps each serial to the master data of every TSS ever set up on this Kasse.
func (r Repository) GetTSEStammdatenJeTSS(ctx context.Context) (map[string]tse.Stammdaten, error) {
	rows, err := r.q.GetTSEStammdatenJeTSS(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	stammdaten := make(map[string]tse.Stammdaten, len(rows))
	for _, row := range rows {
		stammdaten[row.Seriennummer] = tse.Stammdaten(row)
	}
	return stammdaten, nil
}

// UpsertTSEStammdaten stores the master data as the current TSS's and keeps them under the TSS serial,
// so a later TSS change cannot overwrite the data that verify this TSS's signatures.
func (r Repository) UpsertTSEStammdaten(ctx context.Context, s tse.Stammdaten) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		return upsertStammdaten(ctx, qtx, s)
	})
}

func upsertStammdaten(ctx context.Context, qtx *dbgen.Queries, s tse.Stammdaten) error {
	params := dbgen.UpsertTSEStammdatenJeTSSParams{
		Seriennummer:        s.Seriennummer,
		SignaturAlgorithmus: s.SignaturAlgorithmus,
		PublicKey:           s.PublicKey,
		Zertifikat:          s.Zertifikat,
		LogTimeFormat:       s.LogTimeFormat,
	}
	if err := qtx.UpsertTSEStammdaten(ctx, dbgen.UpsertTSEStammdatenParams(params)); err != nil {
		return db.Error(err)
	}
	if err := qtx.UpsertTSEStammdatenJeTSS(ctx, params); err != nil {
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
