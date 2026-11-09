package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/fiskal/export/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// exportWriteTimeout replaces the server's write deadline (docs/handbuch.md §6.2): the DSFinV-K ZIP
// may take longer to send and must not be cut silently, as it holds records subject to retention.
const exportWriteTimeout = 5 * time.Minute

type service interface {
	Erstellen(ctx context.Context, kassensitzungNr int) (application.Archiv, error)
}

type Handler struct {
	Service service
}

// exportRequest selects the Kassensitzung; the number is required.
type exportRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

var exportSchema = z.Struct(z.Shape{
	"KassensitzungNr": kasse.KassensitzungNrSchema().Required(),
})

// ExportHandler streams the chosen Kassensitzung's DSFinV-K ZIP. The /admin/ mount
// already enforces the admin role.
func (h *Handler) ExportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context())

		// First deadline at handler entry: it covers the early error paths that answer
		// before Erstellen() (unreadable or invalid body).
		helper.ExtendWriteDeadline(w, r, exportWriteTimeout)

		body := exportRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, exportSchema) {
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
			log.Error("Failed to write dsfinvk archive", "error", err)
		}
	}
}
