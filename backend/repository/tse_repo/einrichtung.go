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

// SpeichereEinrichtung speichert die TSE-Konfiguration (alle Schreibpfade:
// Einrichtung, Uebernahme, Zugangsdaten-Wechsel, Leeren) und fuehrt beim
// Uebergang von nicht konfiguriert zu konfiguriert in derselben Transaktion den
// Einrichtungs-Sweep aus: alle noch offenen Auftraege aus der
// konfigurationslosen Zeit werden endgueltig als tse_nicht_konfiguriert
// markiert und der keine_konfiguration-Stoerungszeitraum wird geschlossen. War
// die TSE schon vorher konfiguriert (reiner Zugangsdaten-Wechsel), bleibt es
// beim reinen Speichern — laufende Auftraege werden nie versehentlich als nicht
// konfiguriert markiert. Auch das Speichern einer unvollstaendigen
// Konfiguration (Leeren) sweept nichts: Der Dauerzustand ohne Konfiguration
// gehoert dem Signatur-Worker, der Stoerungszeitraum bleibt offen.
func (r Repository) SpeichereEinrichtung(ctx context.Context, c tse.Konfiguration) error {
	return r.withTx(ctx, func(qtx *dbgen.Queries) error {
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
		if _, err := qtx.MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert(ctx); err != nil {
			return db.Error(err)
		}
		if err := qtx.SchliesseTSEStoerung(ctx, tse.StoerungGrundKeineKonfiguration); err != nil {
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

// withTx runs fn within a single transaction: begin, rollback on any error
// (a rollback after commit is a no-op), commit otherwise. fn owns its own error
// wrapping; only begin/commit are normalized via db.Error.
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

// GetTSEStammdaten liest die fiskalischen TSS-Stammdaten fuer den
// DSFinV-K-Export (Singleton). Vor der TSE-Einrichtung sind die Felder leer.
func (r Repository) GetTSEStammdaten(ctx context.Context) (tse.Stammdaten, error) {
	row, err := r.q.GetTSEStammdaten(ctx)
	if err != nil {
		return tse.Stammdaten{}, db.Error(err)
	}
	return tse.Stammdaten{
		SignaturAlgorithmus: row.SignaturAlgorithmus,
		PublicKey:           row.PublicKey,
		Zertifikat:          row.Zertifikat,
		LogTimeFormat:       row.LogTimeFormat,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// UpsertTSEStammdaten speichert die fiskalischen TSS-Stammdaten fuer den
// DSFinV-K-Export (Singleton).
func (r Repository) UpsertTSEStammdaten(ctx context.Context, s tse.Stammdaten) error {
	err := r.q.UpsertTSEStammdaten(ctx, dbgen.UpsertTSEStammdatenParams{
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
