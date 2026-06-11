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

// GetAlleDruckstationen gibt die Konfiguration aller fünf Druckstationen zurück
// (auch unkonfigurierte, mit leerem DruckerIP).
func (r Repository) GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error) {
	rows, err := r.q.GetDruckstationen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make([]druckstation.Druckstation, 0, len(rows))
	for _, row := range rows {
		result = append(result, druckstation.Druckstation{
			Kategorie: druckstation.Kategorie(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  druckstation.Bonmodus(row.Bonmodus.String),
		})
	}
	return result, nil
}

// GetKonfigurierteDruckstationen gibt nur Stationen mit konfiguriertem Drucker zurück.
// Wird von der Arbeitsbon-Policy (Table-Command) genutzt, um Druckaufträge je Kategorie zu erzeugen.
func (r Repository) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	rows, err := r.q.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make(map[string]druckstation.Druckstation, len(rows))
	for _, row := range rows {
		result[string(row.Kategorie)] = druckstation.Druckstation{
			Kategorie: druckstation.Kategorie(row.Kategorie),
			DruckerIP: row.DruckerIp,
			Bonmodus:  druckstation.Bonmodus(row.Bonmodus.String),
		}
	}
	return result, nil
}

// UpsertDruckstation speichert die Drucker-IP und (für Produktkategorien) den Bonmodus
// einer Station. Ein leerer Bonmodus wird als NULL persistiert (kassenbeleg/abholbon).
func (r Repository) UpsertDruckstation(ctx context.Context, station druckstation.Druckstation) error {
	return db.Error(r.q.UpsertDruckstation(ctx, dbgen.UpsertDruckstationParams{
		Kategorie: dbgen.Druckstationkategorie(station.Kategorie),
		DruckerIp: station.DruckerIP,
		Bonmodus:  bonmodusToNull(station.Bonmodus),
	}))
}

func bonmodusToNull(bonmodus druckstation.Bonmodus) sql.NullString {
	if bonmodus == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: string(bonmodus), Valid: true}
}
