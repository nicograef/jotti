package application

import (
	"context"

	"github.com/nicograef/jotti/backend/internal/logging"
)

type druckauftragCommandRepo interface {
	RetryDruckauftrag(ctx context.Context, id int) error
	DiscardDruckauftrag(ctx context.Context, id int) error
	DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error)
}

type Command struct {
	DruckauftragRepo druckauftragCommandRepo
}

// RetryDruckauftrag: Der Status-Guard liegt im Repository — nur fehlgeschlagene
// Aufträge wechseln zurück auf offen (versuche = 0).
func (c Command) RetryDruckauftrag(ctx context.Context, id int) error {
	log := logging.FromContext(ctx)

	if err := c.DruckauftragRepo.RetryDruckauftrag(ctx, id); err != nil {
		log.Error("Failed to retry druckauftrag", "error", err, "id", id)
		return ErrDatabase
	}

	log.Info("Druckauftrag erneut eingereiht", "id", id)
	return nil
}

// DiscardDruckauftrag: Der Status-Guard liegt im Repository; der Eintrag bleibt erhalten.
func (c Command) DiscardDruckauftrag(ctx context.Context, id int) error {
	log := logging.FromContext(ctx)

	if err := c.DruckauftragRepo.DiscardDruckauftrag(ctx, id); err != nil {
		log.Error("Failed to discard druckauftrag", "error", err, "id", id)
		return ErrDatabase
	}

	log.Info("Druckauftrag verworfen", "id", id)
	return nil
}

func (c Command) DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error) {
	log := logging.FromContext(ctx)
	n, err := c.DruckauftragRepo.DiscardAlleFehlgeschlagenen(ctx)
	if err != nil {
		log.Error("Failed to discard all fehlgeschlagene druckauftraege", "error", err)
		return 0, ErrDatabase
	}
	log.Info("Alle fehlgeschlagenen Druckauftraege verworfen", "verworfen", n)
	return n, nil
}
