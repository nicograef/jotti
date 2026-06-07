package api

import (
	"context"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
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
			Bonmodus: row.Bonmodus,
		}
	}
	return result, nil
}

type druckauftragRepoTableAdapter struct {
	repo druckauftrag_repo.Repository
}

func (a druckauftragRepoTableAdapter) EnqueueDruckauftraege(ctx context.Context, auftraege []bondruckApp.Druckauftrag) error {
	toInsert := make([]druckauftrag_repo.NeuerDruckauftrag, 0, len(auftraege))
	for _, auftrag := range auftraege {
		toInsert = append(toInsert, druckauftrag_repo.NeuerDruckauftrag{
			ZielIP:   auftrag.ZielIP,
			Payload:  auftrag.Payload,
			BonArt:   auftrag.BonArt,
			Referenz: auftrag.Referenz,
		})
	}
	return a.repo.EnqueueDruckauftraege(ctx, toInsert)
}
