package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/nicograef/jotti/backend/api/fiskal/export/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/rs/zerolog"
)

type service interface {
	Erstellen(ctx context.Context, kassensitzungNr int) (application.Archiv, error)
}

type Handler struct {
	Service service
}

// exportRequest wählt die zu exportierende Kassensitzung. 0 (bzw. fehlend) steht
// für die Standard-Sitzung (offen, sonst jüngste abgeschlossene).
type exportRequest struct {
	Kassensitzung int `json:"kassensitzung"`
}

// ExportHandler streamt das DSFinV-K-ZIP der gewählten Kassensitzung. Die
// Admin-Rolle wird bereits durch den /admin/-Mount erzwungen.
func (h *Handler) ExportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		body := exportRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}
		if body.Kassensitzung < 0 {
			helper.SendClientError(w, "invalid_kassensitzung", nil)
			return
		}

		archiv, err := h.Service.Erstellen(r.Context(), body.Kassensitzung)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrKassensitzungNichtGefunden):
				helper.SendNotFound(w, "kassensitzung_nicht_gefunden")
			case errors.Is(err, application.ErrLeereKassensitzung):
				helper.SendClientError(w, "leere_kassensitzung", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="`+archiv.Dateiname+`"`)
		w.Header().Set("Content-Length", strconv.Itoa(len(archiv.Inhalt)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(archiv.Inhalt); err != nil {
			log.Error().Err(err).Msg("Failed to write dsfinvk archive")
		}
	}
}
