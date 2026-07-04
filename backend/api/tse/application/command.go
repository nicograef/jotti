package application

import (
	"context"

	"github.com/rs/zerolog"
)

type tseSignaturauftragCommandRepo interface {
	TSESignaturauftragZuruecksetzen(ctx context.Context, auftragID int) error
	TSESignaturauftragVerwerfen(ctx context.Context, auftragID int) error
}

type Command struct {
	TSERepo tseSignaturauftragCommandRepo
}

// TSESignaturauftragZuruecksetzen reiht einen fehlgeschlagenen Auftrag
// wieder ein. Der Status-Guard liegt im Repository: Nur fehlgeschlagene
// Auftraege wechseln zurueck auf offen, andere Status bleiben unberuehrt.
func (c Command) TSESignaturauftragZuruecksetzen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSESignaturauftragZuruecksetzen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to reset tse signaturauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("TSE-Signaturauftrag zurueckgesetzt")
	return nil
}

// TSESignaturauftragVerwerfen markiert einen fehlgeschlagenen Auftrag als
// verworfen. Der Status-Guard liegt im Repository; der Eintrag bleibt fuer
// die Ausfalldokumentation erhalten.
func (c Command) TSESignaturauftragVerwerfen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSESignaturauftragVerwerfen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to discard tse signaturauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("TSE-Signaturauftrag verworfen")
	return nil
}
