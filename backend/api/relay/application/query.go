package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/rs/zerolog"
)

type Query struct {
	EventRepo        eventRepo
	DruckstationRepo druckstationRepo
}

type eventRepo interface {
	GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error)
}

type druckstationRepo interface {
	// Gibt nur Kategorien zurück, für die eine drucker_ip konfiguriert ist.
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]Druckstation, error)
}

// DruckAuftrag ist das Application-DTO, das an den HTTP-Handler weitergegeben wird.
type DruckAuftrag struct {
	EventID   int    // Für Cursor-Tracking
	DruckerIP string // Zur Lesezeit aufgelöst (immer aktuell)
	Payload   string // Base64-kodierter ESC/POS-Byte-String
}

// GetDruckAuftraege liest neue BestellungAufgenommenV1-Events seit dem Cursor
// und erzeugt daraus Druck-Aufträge (1 pro Position oder 1 pro Bestellung je nach Bonmodus).
// Gibt (nil, 0, nil) zurück wenn keine neuen Events vorhanden.
// MaxEventID ist die höchste verarbeitete Event-ID — für Cursor-Tracking auch ohne Drucker-Match.
func (q Query) GetDruckAuftraege(ctx context.Context, lastEventID int) ([]DruckAuftrag, int, error) {
	log := zerolog.Ctx(ctx)

	// 1. Neue BestellungAufgenommenV1-Events lesen
	events, err := q.EventRepo.GetBestellungEventsSinceCursor(ctx, lastEventID)
	if err != nil {
		return nil, 0, err
	}
	if len(events) == 0 {
		return nil, 0, nil
	}

	// 2. Druckstationen zur Lesezeit holen (immer aktuell)
	druckstationen, err := q.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 3. Pro Event Druck-Aufträge erzeugen
	var auftraege []DruckAuftrag
	for _, evt := range events {
		jobs := createDruckAuftraegeFromEvent(evt, druckstationen)
		auftraege = append(auftraege, jobs...)
	}

	maxEventID := events[len(events)-1].ID

	log.Debug().Int("cursor", lastEventID).Int("new_events", len(events)).
		Int("auftraege", len(auftraege)).Msg("Relay poll")

	return auftraege, maxEventID, nil
}
