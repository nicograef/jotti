package http

import (
	"context"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// --- Query Handler ---

type tseSignaturauftragQuery interface {
	GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error)
	GetTSESignaturQueueZustand(ctx context.Context) (tse_repo.SignaturQueueZustand, error)
	GetTSEStoerungen(ctx context.Context) ([]tse_repo.Stoerungszeitraum, error)
}

type QueryHandler struct {
	Query tseSignaturauftragQuery
}

type signaturauftragDTO struct {
	ID             int        `json:"id"`
	TxID           string     `json:"txId"`
	ProcessType    string     `json:"processType"`
	Status         string     `json:"status"`
	Versuche       int        `json:"versuche"`
	LetzterFehler  string     `json:"letzterFehler"`
	ErstelltAm     time.Time  `json:"erstelltAm"`
	ErledigtAm     *time.Time `json:"erledigtAm"`
	VerworfenGrund string     `json:"verworfenGrund"`
	VerworfenVon   string     `json:"verworfenVon"`
	VerworfenAm    *time.Time `json:"verworfenAm"`
}

type getSignaturauftraegeResponse struct {
	Auftraege []signaturauftragDTO `json:"auftraege"`
}

// POST /admin/get-tse-signaturauftraege
func (h *QueryHandler) GetTSESignaturauftraegeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auftraege, err := h.Query.GetTSESignaturauftraege(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]signaturauftragDTO, 0, len(auftraege))
		for _, a := range auftraege {
			dtos = append(dtos, signaturauftragDTO{
				ID:             a.ID,
				TxID:           a.TxID,
				ProcessType:    a.ProcessType,
				Status:         a.Status,
				Versuche:       a.Versuche,
				LetzterFehler:  a.LetzterFehler,
				ErstelltAm:     a.ErstelltAm,
				ErledigtAm:     a.ErledigtAm,
				VerworfenGrund: a.VerworfenGrund,
				VerworfenVon:   a.VerworfenVon,
				VerworfenAm:    a.VerworfenAm,
			})
		}

		helper.SendResponse(w, getSignaturauftraegeResponse{Auftraege: dtos})
	}
}

type signaturQueueResponse struct {
	OffeneAuftraege          int     `json:"offeneAuftraege"`
	FehlgeschlageneAuftraege int     `json:"fehlgeschlageneAuftraege"`
	RueckstandSekunden       int     `json:"rueckstandSekunden"`
	SignaturenProMinute      float64 `json:"signaturenProMinute"`
	SignierdauerP95Sekunden  float64 `json:"signierdauerP95Sekunden"`
}

// POST /admin/get-tse-signatur-queue
func (h *QueryHandler) GetTSESignaturQueueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zustand, err := h.Query.GetTSESignaturQueueZustand(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, signaturQueueResponse{
			OffeneAuftraege:          zustand.OffeneAuftraege,
			FehlgeschlageneAuftraege: zustand.FehlgeschlageneAuftraege,
			RueckstandSekunden:       zustand.RueckstandSekunden,
			SignaturenProMinute:      zustand.SignaturenProMinute,
			SignierdauerP95Sekunden:  zustand.SignierdauerP95Sekunden,
		})
	}
}

type stoerungDTO struct {
	ID         int        `json:"id"`
	Beginn     time.Time  `json:"beginn"`
	Ende       *time.Time `json:"ende"`
	GrundArt   string     `json:"grundArt"`
	Fehlertext string     `json:"fehlertext"`
}

type getStoerungenResponse struct {
	Stoerungen []stoerungDTO `json:"stoerungen"`
}

// POST /admin/get-tse-stoerungen (Stoerungsprotokoll / Ausfalldokumentation)
func (h *QueryHandler) GetTSEStoerungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stoerungen, err := h.Query.GetTSEStoerungen(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]stoerungDTO, 0, len(stoerungen))
		for _, s := range stoerungen {
			dtos = append(dtos, stoerungDTO{
				ID:         s.ID,
				Beginn:     s.Beginn,
				Ende:       s.Ende,
				GrundArt:   s.GrundArt,
				Fehlertext: s.Fehlertext,
			})
		}

		helper.SendResponse(w, getStoerungenResponse{Stoerungen: dtos})
	}
}

// --- Command Handler ---

type tseSignaturauftragCommand interface {
	TSESignaturauftragZuruecksetzen(ctx context.Context, id int) error
	TSESignaturauftraegeZuruecksetzenGesamt(ctx context.Context) (int, error)
	TSESignaturauftragVerwerfen(ctx context.Context, id int, grund string, benutzer string) error
}

type CommandHandler struct {
	Command tseSignaturauftragCommand
}

type signaturauftragIDRequest struct {
	ID int `json:"id"`
}

var signaturauftragIDSchema = z.Struct(z.Shape{
	"ID": z.Int().GTE(1, z.Message("Ungültige Auftrag-ID")).Required(),
})

// POST /admin/tse-signaturauftrag-zuruecksetzen
func (h *CommandHandler) TSESignaturauftragZuruecksetzenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body signaturauftragIDRequest
		if !helper.ReadAndValidateBody(w, r, &body, signaturauftragIDSchema) {
			return
		}

		if err := h.Command.TSESignaturauftragZuruecksetzen(r.Context(), body.ID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type signaturauftraegeZuruecksetzenResponse struct {
	Anzahl int `json:"anzahl"`
}

// POST /admin/tse-signaturauftraege-zuruecksetzen (gesamt)
func (h *CommandHandler) TSESignaturauftraegeZuruecksetzenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		anzahl, err := h.Command.TSESignaturauftraegeZuruecksetzenGesamt(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, signaturauftraegeZuruecksetzenResponse{Anzahl: anzahl})
	}
}

type signaturauftragVerwerfenRequest struct {
	ID    int    `json:"id"`
	Grund string `json:"grund"`
}

var signaturauftragVerwerfenSchema = z.Struct(z.Shape{
	"ID":    z.Int().GTE(1, z.Message("Ungültige Auftrag-ID")).Required(),
	"Grund": z.String().Trim().Min(1, z.Message("Eine Begründung ist erforderlich")).Max(500, z.Message("Die Begründung darf höchstens 500 Zeichen lang sein")).Required(),
})

// POST /admin/tse-signaturauftrag-verwerfen
func (h *CommandHandler) TSESignaturauftragVerwerfenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body signaturauftragVerwerfenRequest
		if !helper.ReadAndValidateBody(w, r, &body, signaturauftragVerwerfenSchema) {
			return
		}

		_, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		if err := h.Command.TSESignaturauftragVerwerfen(r.Context(), body.ID, body.Grund, userName); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
