package application

import (
	"context"

	"github.com/rs/zerolog"
)

type druckauftragCommandRepo interface {
	DruckauftragErneutVersuchen(ctx context.Context, id int) error
	DruckauftragVerwerfen(ctx context.Context, id int) error
}

type Command struct {
	DruckauftragRepo druckauftragCommandRepo
}

// DruckauftragErneutVersuchen reiht einen fehlgeschlagenen Auftrag wieder ein.
// Der Status-Guard liegt im Repository: nur fehlgeschlagene Aufträge wechseln
// zurück auf offen (versuche = 0), andere Status bleiben unberührt.
func (c Command) DruckauftragErneutVersuchen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.DruckauftragRepo.DruckauftragErneutVersuchen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to retry druckauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("Druckauftrag erneut eingereiht")
	return nil
}

// DruckauftragVerwerfen markiert einen fehlgeschlagenen Auftrag als verworfen.
// Der Status-Guard liegt im Repository; der Eintrag bleibt erhalten.
func (c Command) DruckauftragVerwerfen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.DruckauftragRepo.DruckauftragVerwerfen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to discard druckauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("Druckauftrag verworfen")
	return nil
}
