package http

import (
	"context"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// --- Query Handler ---

type druckauftragQuery interface {
	GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckauftrag_repo.FehlgeschlagenerDruckauftrag, error)
}

type QueryHandler struct {
	Query druckauftragQuery
}

type fehlgeschlagenerDruckauftragDTO struct {
	ID            int       `json:"id"`
	BonArt        string    `json:"bonArt"`
	ZielIP        string    `json:"zielIp"`
	Referenz      string    `json:"referenz"`
	Versuche      int       `json:"versuche"`
	LetzterFehler string    `json:"letzterFehler"`
	ErstelltAm    time.Time `json:"erstelltAm"`
}

type getFehlgeschlageneResponse struct {
	Druckauftraege []fehlgeschlagenerDruckauftragDTO `json:"druckauftraege"`
}

// POST /admin/get-fehlgeschlagene-druckauftraege
func (h *QueryHandler) GetFehlgeschlageneDruckauftraegeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auftraege, err := h.Query.GetFehlgeschlageneDruckauftraege(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]fehlgeschlagenerDruckauftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, fehlgeschlagenerDruckauftragDTO{
				ID:            a.ID,
				BonArt:        a.BonArt,
				ZielIP:        a.ZielIP,
				Referenz:      a.Referenz,
				Versuche:      a.Versuche,
				LetzterFehler: a.LetzterFehler,
				ErstelltAm:    a.ErstelltAm,
			})
		}

		helper.SendResponse(w, getFehlgeschlageneResponse{Druckauftraege: dtos})
	}
}

// --- Command Handler ---

type druckauftragCommand interface {
	RetryDruckauftrag(ctx context.Context, id int) error
	DiscardDruckauftrag(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command druckauftragCommand
}

type druckauftragRequest struct {
	ID int `json:"id"`
}

var druckauftragSchema = z.Struct(z.Shape{
	"ID": z.Int().GTE(1, z.Message("Ungültige Druckauftrag-ID")).Required(),
})

// POST /admin/druckauftrag-erneut-versuchen
func (h *CommandHandler) RetryDruckauftragHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body druckauftragRequest
		if !helper.ReadAndValidateBody(w, r, &body, druckauftragSchema) {
			return
		}

		if err := h.Command.RetryDruckauftrag(r.Context(), body.ID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// POST /admin/druckauftrag-verwerfen
func (h *CommandHandler) DiscardDruckauftragHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body druckauftragRequest
		if !helper.ReadAndValidateBody(w, r, &body, druckauftragSchema) {
			return
		}

		if err := h.Command.DiscardDruckauftrag(r.Context(), body.ID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
