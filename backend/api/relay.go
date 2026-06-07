package api

import (
	"context"
	"database/sql"
	"net/http"

	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
	relayHTTP "github.com/nicograef/jotti/backend/api/relay/http"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

type druckauftragRepoRelayAdapter struct {
	repo druckauftrag_repo.Repository
}

func (a druckauftragRepoRelayAdapter) GetOffeneDruckauftraege(ctx context.Context) ([]relayApp.DruckAuftrag, error) {
	rows, err := a.repo.GetOffeneDruckauftraege(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]relayApp.DruckAuftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, relayApp.DruckAuftrag{
			ID:      row.ID,
			ZielIP:  row.ZielIP,
			Payload: row.Payload,
		})
	}

	return result, nil
}

func (a druckauftragRepoRelayAdapter) QuittiereGedruckteAuftraege(ctx context.Context, ids []int) error {
	return a.repo.QuittiereGedruckteAuftraege(ctx, ids)
}

func NewRelayApi(db *sql.DB, relayToken string) http.Handler {
	r := http.NewServeMux()

	druckauftragRepo := druckauftragRepoRelayAdapter{repo: druckauftrag_repo.NewRepository(db)}

	handler := relayHTTP.Handler{
		Query: relayApp.Query{
			DruckauftragRepo: druckauftragRepo,
		},
		Command:    relayApp.Command{DruckauftragRepo: druckauftragRepo},
		RelayToken: relayToken,
	}

	r.HandleFunc("/poll", handler.PollHandler())
	r.HandleFunc("/quittieren", handler.QuittierenHandler())

	return r
}
