package http

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
)

// druckauftragRepo ist die Repository-Schnittstelle, die das Relay-API direkt
// nutzt. Zwischen Handler und Repository liegt bewusst keine reine
// Durchreich-Schicht; die Verdrahtung erfolgt im Composition Root (api/relay.go).
type druckauftragRepo interface {
	GetOffeneDruckauftraege(ctx context.Context) ([]OffenerDruckauftrag, error)
	MeldeDruckergebnis(ctx context.Context, gedruckteIDs []int, fehlversuche []Fehlversuch) error
}

// OffenerDruckauftrag ist ein offener Druckauftrag, wie ihn das Relay pollt.
type OffenerDruckauftrag struct {
	ID      int
	ZielIP  string
	Payload string
}

// Fehlversuch meldet einen fehlgeschlagenen Zustellversuch eines Druckauftrags.
type Fehlversuch struct {
	ID     int
	Fehler string
}

type Handler struct {
	Repo       druckauftragRepo
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

// POST /relay/ergebnis
// Request:  {"token": "...", "gedruckteIds": [1,2,3], "fehlversuche": [{"id": 4, "fehler": "..."}]}
// Response: {"status": "ok"}
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

		if err := h.Repo.MeldeDruckergebnis(r.Context(), body.GedruckteIDs, fehlversuche); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, map[string]string{"status": "ok"})
	}
}
