package druckstation_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{q: dbgen.New(database)}
}

// GetAlleDruckstationen gibt die Konfiguration aller drei Kategorien zurück
// (auch unkonfigurierte, mit leerem DruckerIP).
func (r Repository) GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error) {
	rows, err := r.q.GetDruckstationen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make([]druckstation.Druckstation, 0, len(rows))
	for _, row := range rows {
		result = append(result, druckstation.Druckstation{
			Kategorie: string(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  row.Bonmodus,
		})
	}
	return result, nil
}

// GetKonfigurierteDruckstationen gibt nur Kategorien mit konfiguriertem Drucker zurück.
// Wird von der Arbeitsbon-Policy (Table-Command) genutzt, um Druckaufträge je Kategorie zu erzeugen.
func (r Repository) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	rows, err := r.q.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make(map[string]druckstation.Druckstation, len(rows))
	for _, row := range rows {
		result[string(row.Kategorie)] = druckstation.Druckstation{
			Kategorie: string(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  row.Bonmodus,
		}
	}
	return result, nil
}

// UpsertDruckstation speichert die Drucker-IP und den Bonmodus für eine Kategorie.
func (r Repository) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	return db.Error(r.q.UpsertDruckstation(ctx, dbgen.UpsertDruckstationParams{
		Kategorie: dbgen.Produktkategorie(kategorie),
		DruckerIp: druckerIP,
		Bonmodus:  bonmodus,
	}))
}
