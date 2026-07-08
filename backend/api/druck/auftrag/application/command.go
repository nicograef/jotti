package application

import (
	"context"

	"github.com/rs/zerolog"
)

type druckauftragCommandRepo interface {
	RetryDruckauftrag(ctx context.Context, id int) error
	DiscardDruckauftrag(ctx context.Context, id int) error
	DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error)
}

type Command struct {
	DruckauftragRepo druckauftragCommandRepo
}

// RetryDruckauftrag reiht einen fehlgeschlagenen Auftrag wieder ein.
// Der Status-Guard liegt im Repository: nur fehlgeschlagene Aufträge wechseln
// zurück auf offen (versuche = 0), andere Status bleiben unberührt.
func (c Command) RetryDruckauftrag(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.DruckauftragRepo.RetryDruckauftrag(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to retry druckauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("Druckauftrag erneut eingereiht")
	return nil
}

// DiscardDruckauftrag markiert einen fehlgeschlagenen Auftrag als verworfen.
// Der Status-Guard liegt im Repository; der Eintrag bleibt erhalten.
func (c Command) DiscardDruckauftrag(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.DruckauftragRepo.DiscardDruckauftrag(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to discard druckauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("Druckauftrag verworfen")
	return nil
}

// DiscardAlleFehlgeschlagenen verwirft alle fehlgeschlagenen Auftraege in einem
// Schritt und liefert die Anzahl.
func (c Command) DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error) {
	log := zerolog.Ctx(ctx)
	n, err := c.DruckauftragRepo.DiscardAlleFehlgeschlagenen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to discard all fehlgeschlagene druckauftraege")
		return 0, ErrDatabase
	}
	log.Info().Int64("verworfen", n).Msg("Alle fehlgeschlagenen Druckauftraege verworfen")
	return n, nil
}
