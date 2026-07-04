package api

import (
	"context"
	"database/sql"
	"net/http"

	relayHTTP "github.com/nicograef/jotti/backend/api/druck/relay/http"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// druckauftragRepoRelayAdapter mappt die Repository-Typen auf die Typen der
// Relay-HTTP-Schicht, damit diese das Repository nicht direkt importiert.
type druckauftragRepoRelayAdapter struct {
	repo druckauftrag_repo.Repository
}

func (a druckauftragRepoRelayAdapter) GetOffeneDruckauftraege(ctx context.Context) ([]relayHTTP.OffenerDruckauftrag, error) {
	rows, err := a.repo.GetOffeneDruckauftraege(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]relayHTTP.OffenerDruckauftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, relayHTTP.OffenerDruckauftrag{
			ID:      row.ID,
			ZielIP:  row.ZielIP,
			Payload: row.Payload,
		})
	}

	return result, nil
}

func (a druckauftragRepoRelayAdapter) MeldeDruckergebnis(ctx context.Context, gedruckteIDs []int, fehlversuche []relayHTTP.Fehlversuch) error {
	repoFehlversuche := make([]druckauftrag_repo.Fehlversuch, 0, len(fehlversuche))
	for _, f := range fehlversuche {
		repoFehlversuche = append(repoFehlversuche, druckauftrag_repo.Fehlversuch{ID: f.ID, Fehler: f.Fehler})
	}

	return a.repo.MeldeDruckergebnis(ctx, gedruckteIDs, repoFehlversuche)
}

func NewRelayApi(db *sql.DB, relayToken string) http.Handler {
	r := http.NewServeMux()

	handler := relayHTTP.Handler{
		Repo:       druckauftragRepoRelayAdapter{repo: druckauftrag_repo.NewRepository(db)},
		RelayToken: relayToken,
	}

	r.HandleFunc("/poll", handler.PollHandler())
	r.HandleFunc("/ergebnis", handler.ErgebnisHandler())

	return r
}
