package api

import (
	"net/http"

	relayHTTP "github.com/nicograef/jotti/backend/api/druck/relay/http"
)

func NewRelayApi(deps Deps, relayToken string) (http.Handler, []string) {
	r := newRouteMux()

	handler := relayHTTP.Handler{
		Repo:       deps.DruckauftragRepo,
		RelayToken: relayToken,
	}

	r.HandleFunc("/poll", handler.PollHandler())
	r.HandleFunc("/ergebnis", handler.ErgebnisHandler())

	return r.Handler(), r.Paths()
}
