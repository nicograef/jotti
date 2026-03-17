package api

import (
	"context"
	"database/sql"
	"net/http"

	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
	relayHTTP "github.com/nicograef/jotti/backend/api/relay/http"
	"github.com/nicograef/jotti/backend/repository/drucker_repo"
	"github.com/nicograef/jotti/backend/repository/event_repo"
)

// druckerRepoRelayAdapter adapts drucker_repo.Repository to the relay application's druckerRepo interface.
// The repo returns drucker_repo.DruckerKonfig; the relay application expects application.DruckerKonfig.
type druckerRepoRelayAdapter struct {
	repo drucker_repo.Repository
}

func (a druckerRepoRelayAdapter) GetKonfigurierteKategorieDrucker(ctx context.Context) (map[string]relayApp.DruckerKonfig, error) {
	rows, err := a.repo.GetKonfigurierteKategorieDrucker(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]relayApp.DruckerKonfig, len(rows))
	for k, v := range rows {
		result[k] = relayApp.DruckerKonfig{
			IP:       v.DruckerIP,
			Bonmodus: v.Bonmodus,
		}
	}
	return result, nil
}

func NewRelayApi(db *sql.DB, relayToken string) http.Handler {
	r := http.NewServeMux()

	eventRepo := event_repo.NewRepository(db)
	druckerRepo := drucker_repo.NewRepository(db)

	handler := relayHTTP.Handler{
		Query: relayApp.Query{
			EventRepo:   eventRepo,
			DruckerRepo: druckerRepoRelayAdapter{repo: druckerRepo},
		},
		RelayToken: relayToken,
	}

	r.HandleFunc("/poll", handler.PollHandler())

	return r
}
