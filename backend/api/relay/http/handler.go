package http

import (
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
)

type Handler struct {
	Query      relayApp.Query
	RelayToken string
}

// POST /relay/poll
// Request:  {"token": "...", "lastEventId": 42}
// Response: {"auftraege": [...], "cursor": 55}
type pollRequest struct {
	Token       string `json:"token"`
	LastEventID int    `json:"lastEventId"`
}

type pollResponse struct {
	Auftraege []druckAuftragDTO `json:"auftraege"`
	Cursor    int               `json:"cursor"`
}

type druckAuftragDTO struct {
	EventID   int    `json:"eventId"`
	DruckerIP string `json:"druckerIp"`
	Payload   string `json:"payload"` // Base64 ESC/POS
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

		auftraege, err := h.Query.GetDruckAuftraege(r.Context(), body.LastEventID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		// Cursor = höchste verarbeitete Event-ID (oder unverändert wenn keine neuen Events)
		cursor := body.LastEventID
		if len(auftraege) > 0 {
			cursor = auftraege[len(auftraege)-1].EventID
		}

		dtos := make([]druckAuftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, druckAuftragDTO{
				EventID:   a.EventID,
				DruckerIP: a.DruckerIP,
				Payload:   a.Payload,
			})
		}

		helper.SendResponse(w, pollResponse{
			Auftraege: dtos,
			Cursor:    cursor,
		})
	}
}
