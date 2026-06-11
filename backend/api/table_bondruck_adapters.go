package api

import (
	"context"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

type druckstationRepoTableAdapter struct {
	repo druckstation_repo.Repository
}

func (a druckstationRepoTableAdapter) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]bondruckApp.Druckstation, error) {
	rows, err := a.repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bondruckApp.Druckstation, len(rows))
	for kategorie, row := range rows {
		result[kategorie] = bondruckApp.Druckstation{
			IP:       row.DruckerIP,
			Bonmodus: string(row.Bonmodus),
		}
	}
	return result, nil
}
