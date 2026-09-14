package http

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
)

// druckauftragRepo: zwischen Handler und Repository liegt bewusst keine reine
// Durchreich-Schicht; verdrahtet wird im Composition Root (api/relay.go).
type druckauftragRepo interface {
	GetOffeneDruckauftraege(ctx context.Context) ([]OffenerDruckauftrag, error)
	ReportDruckergebnis(ctx context.Context, gedruckteIDs []int, fehlversuche []Fehlversuch) error
}

type OffenerDruckauftrag struct {
	ID      int
	ZielIP  string
	Payload string
}

type Fehlversuch struct {
	ID     int
	Fehler string
}

type Handler struct {
	Repo       druckauftragRepo
	RelayToken string
}

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

type ergebnisRequest struct {
	Token        string           `json:"token"`
	GedruckteIDs []int            `json:"gedruckteIds"`
	Fehlversuche []fehlversuchDTO `json:"fehlversuche"`
}

type fehlversuchDTO struct {
	ID     int    `json:"id"`
	Fehler string `json:"fehler"`
}

func (h *Handler) isRelayTokenValid(token string) bool {
	if h.RelayToken == "" {
		return false
	}

	if token == "" {
		return false
	}

	// Konstant-zeitlicher Vergleich (wie beim Passwortpfad): kein Timing-Seitenkanal
	// auf den statischen Relay-Token.
	return subtle.ConstantTimeCompare([]byte(token), []byte(h.RelayToken)) == 1
}

func (h *Handler) PollHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body pollRequest
		if !helper.ReadBody(w, r, &body) {
			return
		}

		// Statischer Token-Vergleich — das Relay ist kein Benutzer, kein JWT
		if !h.isRelayTokenValid(body.Token) {
			helper.SendClientError(w, "unauthorized", nil)
			return
		}

		auftraege, err := h.Repo.GetOffeneDruckauftraege(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]druckAuftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, druckAuftragDTO(a))
		}

		helper.SendResponse(w, pollResponse{
			Auftraege: dtos,
		})
	}
}

func (h *Handler) ErgebnisHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body ergebnisRequest
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if !h.isRelayTokenValid(body.Token) {
			helper.SendClientError(w, "unauthorized", nil)
			return
		}

		fehlversuche := make([]Fehlversuch, 0, len(body.Fehlversuche))
		for _, f := range body.Fehlversuche {
			fehlversuche = append(fehlversuche, Fehlversuch(f))
		}

		if err := h.Repo.ReportDruckergebnis(r.Context(), body.GedruckteIDs, fehlversuche); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
