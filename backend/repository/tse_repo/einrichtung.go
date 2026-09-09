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

// SaveEinrichtung speichert die TSE-Konfiguration (alle Schreibpfade:
// Einrichtung, Übernahme, Zugangsdaten-Wechsel, Leeren) und führt beim
// Übergang von nicht konfiguriert zu konfiguriert in derselben Transaktion den
// Einrichtungs-Sweep aus: alle noch offenen Aufträge aus der
// konfigurationslosen Zeit werden endgültig als tse_nicht_konfiguriert
// markiert und der keine_konfiguration-Störungszeitraum wird geschlossen. War
// die TSE schon vorher konfiguriert (reiner Zugangsdaten-Wechsel), bleibt es
// beim reinen Speichern — laufende Aufträge werden nie versehentlich als nicht
// konfiguriert markiert. Auch das Speichern einer unvollständigen
// Konfiguration (Leeren) sweept nichts: Der Dauerzustand ohne Konfiguration
// gehört dem Signatur-Worker, der Störungszeitraum bleibt offen.
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

// GetTSEStammdaten liest die fiskalischen TSS-Stammdaten für den
// DSFinV-K-Export (Singleton). Vor der TSE-Einrichtung sind die Felder leer.
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

// UpsertTSEStammdaten speichert die fiskalischen TSS-Stammdaten für den
// DSFinV-K-Export (Singleton).
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
