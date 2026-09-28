package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/nicograef/jotti/backend/api/fiskal/export/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/rs/zerolog"
)

// exportWriteTimeout replaces the server's global 10-second write deadline
// (backend/app/app.go) here: the DSFinV-K ZIP may take longer to send and must
// not be cut silently, as it holds records subject to retention.
const exportWriteTimeout = 5 * time.Minute

type service interface {
	Erstellen(ctx context.Context, kassensitzungNr int) (application.Archiv, error)
}

type Handler struct {
	Service service
}

// exportRequest selects the Kassensitzung; 0 or missing means the default session
// (the open one, else the latest closed).
type exportRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

// ExportHandler streams the chosen Kassensitzung's DSFinV-K ZIP. The /admin/ mount
// already enforces the admin role.
func (h *Handler) ExportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		// First deadline at handler entry: it covers the early error paths that answer
		// before Erstellen() (unreadable body, invalid_kassensitzung).
		helper.ExtendWriteDeadline(w, r, exportWriteTimeout)

		body := exportRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}
		if body.KassensitzungNr < 0 {
			helper.SendClientError(w, "invalid_kassensitzung", nil)
			return
		}

		archiv, err := h.Service.Erstellen(r.Context(), body.KassensitzungNr)

		// Second deadline, for the write itself: the one above is absolute from request
		// start and has expired after a long archive build. It also covers the error branch.
		helper.ExtendWriteDeadline(w, r, exportWriteTimeout)

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
