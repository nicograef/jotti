package http

import (
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
)

type Handler struct {
	Query      relayApp.Query
	Command    relayApp.Command
	RelayToken string
}

// POST /relay/poll
// Request:  {"token": "..."}
// Response: {"auftraege": [...]}
type pollRequest struct {
	Token string `json:"token"`
}

type pollResponse struct {
	Auftraege []druckAuftragDTO `json:"auftraege"`
}

type druckAuftragDTO struct {
	ID      int    `json:"id"`
	ZielIP  string `json:"zielIp"`
	Payload string `json:"payload"` // Base64 ESC/POS
}

type quittierenRequest struct {
	Token        string `json:"token"`
	GedruckteIDs []int  `json:"gedruckteIds"`
}

func (h *Handler) PollHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body pollRequest
		if !helper.ReadBody(w, r, &body) {
			return
		}

		// Statischer Token-Vergleich — das Relay ist kein Benutzer, kein JWT
		if body.Token != h.RelayToken {
			helper.SendClientError(w, "unauthorized", nil)
			return
		}

		auftraege, err := h.Query.GetOffeneDruckauftraege(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]druckAuftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, druckAuftragDTO{
				ID:      a.ID,
				ZielIP:  a.ZielIP,
				Payload: a.Payload,
			})
		}

		helper.SendResponse(w, pollResponse{
			Auftraege: dtos,
		})
	}
}

// POST /relay/quittieren
// Request:  {"token": "...", "gedruckteIds": [1,2,3]}
// Response: {"status": "ok"}
func (h *Handler) QuittierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body quittierenRequest
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Token != h.RelayToken {
			helper.SendClientError(w, "unauthorized", nil)
			return
		}

		if err := h.Command.QuittiereGedruckteAuftraege(r.Context(), body.GedruckteIDs); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, map[string]string{"status": "ok"})
	}
}
