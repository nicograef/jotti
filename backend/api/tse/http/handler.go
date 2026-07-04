package http

import (
	"context"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// --- Query Handler ---

type tseSignaturauftragQuery interface {
	GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error)
}

type QueryHandler struct {
	Query tseSignaturauftragQuery
}

type nachsignierAuftragDTO struct {
	ID            int        `json:"id"`
	TxID          string     `json:"txId"`
	ProcessType   string     `json:"processType"`
	Status        string     `json:"status"`
	Versuche      int        `json:"versuche"`
	LetzterFehler string     `json:"letzterFehler"`
	ErstelltAm    time.Time  `json:"erstelltAm"`
	ErledigtAm    *time.Time `json:"erledigtAm"`
}

type getNachsignierAuftraegeResponse struct {
	Auftraege []nachsignierAuftragDTO `json:"auftraege"`
}

// POST /admin/get-tse-nachsignier-auftraege (Signaturauftraege; Endpunkt-Name
// bleibt bis zur Umbenennung der Endpunktfamilie in Phase 6 bestehen)
func (h *QueryHandler) GetTSENachsignierAuftraegeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auftraege, err := h.Query.GetTSESignaturauftraege(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]nachsignierAuftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, nachsignierAuftragDTO{
				ID:            a.ID,
				TxID:          a.TxID,
				ProcessType:   a.ProcessType,
				Status:        a.Status,
				Versuche:      a.Versuche,
				LetzterFehler: a.LetzterFehler,
				ErstelltAm:    a.ErstelltAm,
				ErledigtAm:    a.ErledigtAm,
			})
		}

		helper.SendResponse(w, getNachsignierAuftraegeResponse{Auftraege: dtos})
	}
}

// --- Command Handler ---

type tseSignaturauftragCommand interface {
	TSESignaturauftragZuruecksetzen(ctx context.Context, id int) error
	TSESignaturauftragVerwerfen(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command tseSignaturauftragCommand
}

type nachsignierAuftragRequest struct {
	ID int `json:"id"`
}

var nachsignierAuftragSchema = z.Struct(z.Shape{
	"ID": z.Int().GTE(1, z.Message("Ungültige Auftrag-ID")).Required(),
})

// POST /admin/tse-nachsignier-auftrag-zuruecksetzen
func (h *CommandHandler) TSENachsignierAuftragZuruecksetzenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body nachsignierAuftragRequest
		if !helper.ReadAndValidateBody(w, r, &body, nachsignierAuftragSchema) {
			return
		}

		if err := h.Command.TSESignaturauftragZuruecksetzen(r.Context(), body.ID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// POST /admin/tse-nachsignier-auftrag-verwerfen
func (h *CommandHandler) TSENachsignierAuftragVerwerfenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body nachsignierAuftragRequest
		if !helper.ReadAndValidateBody(w, r, &body, nachsignierAuftragSchema) {
			return
		}

		if err := h.Command.TSESignaturauftragVerwerfen(r.Context(), body.ID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
