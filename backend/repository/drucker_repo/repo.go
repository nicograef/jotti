package drucker_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type DruckerKonfig struct {
	Kategorie string
	DruckerIP string
	Bonmodus  string // "pro_position" | "pro_bestellung"
}

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{q: dbgen.New(database)}
}

// GetAlleKategorieDrucker gibt die Konfiguration aller drei Kategorien zurück
// (auch unkonfigurierte, mit leerem DruckerIP).
func (r Repository) GetAlleKategorieDrucker(ctx context.Context) ([]DruckerKonfig, error) {
	rows, err := r.q.GetKategorieDrucker(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make([]DruckerKonfig, 0, len(rows))
	for _, row := range rows {
		result = append(result, DruckerKonfig{
			Kategorie: string(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  row.Bonmodus,
		})
	}
	return result, nil
}

// GetKonfigurierteKategorieDrucker gibt nur Kategorien mit konfiguriertem Drucker zurück.
// Wird vom Relay-Service verwendet.
func (r Repository) GetKonfigurierteKategorieDrucker(ctx context.Context) (map[string]DruckerKonfig, error) {
	rows, err := r.q.GetKonfigurierteKategorieDrucker(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make(map[string]DruckerKonfig, len(rows))
	for _, row := range rows {
		result[string(row.Kategorie)] = DruckerKonfig{
			Kategorie: string(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  row.Bonmodus,
		}
	}
	return result, nil
}

// UpsertKategorieDrucker speichert die Drucker-IP und den Bonmodus für eine Kategorie.
func (r Repository) UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	return db.Error(r.q.UpsertKategorieDrucker(ctx, dbgen.UpsertKategorieDruckerParams{
		Kategorie: dbgen.Produktkategorie(kategorie),
		DruckerIp: druckerIP,
		Bonmodus:  bonmodus,
	}))
}
