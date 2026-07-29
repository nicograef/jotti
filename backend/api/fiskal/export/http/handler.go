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

// exportWriteTimeout ersetzt fuer diesen Handler die globale 10-Sekunden-
// Schreibfrist des Servers (backend/app/app.go): Das DSFinV-K-ZIP kann laenger
// zum Uebertragen brauchen als jede andere Antwort und darf dabei nicht
// stillschweigend abgeschnitten werden (aufbewahrungspflichtige Daten).
const exportWriteTimeout = 5 * time.Minute

type service interface {
	Erstellen(ctx context.Context, kassensitzungNr int) (application.Archiv, error)
}

type Handler struct {
	Service service
}

// exportRequest wählt die zu exportierende Kassensitzung. 0 (bzw. fehlend) steht
// für die Standard-Sitzung (offen, sonst jüngste abgeschlossene).
type exportRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

// ExportHandler streamt das DSFinV-K-ZIP der gewählten Kassensitzung. Die
// Admin-Rolle wird bereits durch den /admin/-Mount erzwungen.
func (h *Handler) ExportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		// Erste Setzung, am Handler-Eingang: Sie gilt den fruehen Fehlerpfaden,
		// die vor Erstellen() antworten (unlesbarer Body, invalid_kassensitzung).
		// Die Antwort nach einem langen Archivbau deckt sie nicht — dafuer steht
		// die zweite Setzung unten.
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

		// Zweites Setzen der Schreibfrist, jetzt fuer den Schreibvorgang selbst:
		// Die Frist oben ist eine absolute Zeit ab Request-Start und nach einem
		// langen Archivbau abgelaufen. Erst dieser Aufruf gibt der Uebertragung
		// des ZIP ihr eigenes Budget; er deckt zugleich den Fehlerzweig ab.
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
