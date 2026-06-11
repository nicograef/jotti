package application

import (
	"context"

	"github.com/rs/zerolog"
)

type tseNachsignierCommandRepo interface {
	TSENachsignierAuftragZuruecksetzen(ctx context.Context, auftragID int) error
	TSENachsignierAuftragVerwerfen(ctx context.Context, auftragID int) error
}

type Command struct {
	TSERepo tseNachsignierCommandRepo
}

// TSENachsignierAuftragZuruecksetzen reiht einen fehlgeschlagenen Auftrag
// wieder ein. Der Status-Guard liegt im Repository: Nur fehlgeschlagene
// Auftraege wechseln zurueck auf offen, andere Status bleiben unberuehrt.
func (c Command) TSENachsignierAuftragZuruecksetzen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSENachsignierAuftragZuruecksetzen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to reset tse nachsignier auftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("TSE-Nachsignier-Auftrag zurueckgesetzt")
	return nil
}

// TSENachsignierAuftragVerwerfen markiert einen fehlgeschlagenen Auftrag als
// verworfen. Der Status-Guard liegt im Repository; der Eintrag bleibt fuer
// die Ausfalldokumentation erhalten.
func (c Command) TSENachsignierAuftragVerwerfen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSENachsignierAuftragVerwerfen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to discard tse nachsignier auftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("TSE-Nachsignier-Auftrag verworfen")
	return nil
}
