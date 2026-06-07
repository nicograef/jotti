package api

import (
	"context"
	"database/sql"
	"net/http"

	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
	relayHTTP "github.com/nicograef/jotti/backend/api/relay/http"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

// druckstationRepoRelayAdapter adapts druckstation_repo.Repository to the relay application's druckstationRepo interface.
// The repo returns druckstation_repo.Druckstation; the relay application expects application.Druckstation.
// This adapter is retained to avoid an import cycle: relay/application cannot directly import druckstation_repo
// without creating a circular dependency between the application and repository layers.
type druckstationRepoRelayAdapter struct {
	repo druckstation_repo.Repository
}

func (a druckstationRepoRelayAdapter) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]relayApp.Druckstation, error) {
	rows, err := a.repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]relayApp.Druckstation, len(rows))
	for k, v := range rows {
		result[k] = relayApp.Druckstation{
			IP:       v.DruckerIP,
			Bonmodus: v.Bonmodus,
		}
	}
	return result, nil
}

func NewRelayApi(db *sql.DB, relayToken string) http.Handler {
	r := http.NewServeMux()

	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	druckstationRepo := druckstation_repo.NewRepository(db)

	handler := relayHTTP.Handler{
		Query: relayApp.Query{
			EventRepo:        kassenjournalRepo,
			DruckstationRepo: druckstationRepoRelayAdapter{repo: druckstationRepo},
		},
		RelayToken: relayToken,
	}

	r.HandleFunc("/poll", handler.PollHandler())

	return r
}
