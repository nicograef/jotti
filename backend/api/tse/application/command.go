package application

import (
	"context"

	"github.com/rs/zerolog"
)

type tseSignaturauftragCommandRepo interface {
	TSESignaturauftragZuruecksetzen(ctx context.Context, auftragID int) error
	TSESignaturauftraegeZuruecksetzenGesamt(ctx context.Context) (int64, error)
	TSESignaturauftragVerwerfen(ctx context.Context, auftragID int, grund string, benutzer string) error
}

type Command struct {
	TSERepo tseSignaturauftragCommandRepo
}

// TSESignaturauftragZuruecksetzen reiht einen endgueltig markierten Auftrag
// wieder ein. Der Status-Guard liegt im Repository: Nur fehlgeschlagene und
// tse_nicht_konfiguriert Auftraege wechseln zurueck auf offen, andere Status
// bleiben unberuehrt.
func (c Command) TSESignaturauftragZuruecksetzen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSESignaturauftragZuruecksetzen(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to reset tse signaturauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Msg("TSE-Signaturauftrag zurueckgesetzt")
	return nil
}

// TSESignaturauftraegeZuruecksetzenGesamt reiht alle endgueltig markierten
// Auftraege wieder ein und liefert die Anzahl der zurueckgesetzten Auftraege.
func (c Command) TSESignaturauftraegeZuruecksetzenGesamt(ctx context.Context) (int, error) {
	log := zerolog.Ctx(ctx)

	n, err := c.TSERepo.TSESignaturauftraegeZuruecksetzenGesamt(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset all tse signaturauftraege")
		return 0, ErrDatabase
	}

	log.Info().Int64("count", n).Msg("TSE-Signaturauftraege gesamt zurueckgesetzt")
	return int(n), nil
}

// TSESignaturauftragVerwerfen markiert einen offenen oder fehlgeschlagenen
// Auftrag als verworfen und protokolliert Grund, Benutzer und Zeitpunkt. Der
// Status-Guard liegt im Repository; der Eintrag bleibt fuer die
// Ausfalldokumentation erhalten.
func (c Command) TSESignaturauftragVerwerfen(ctx context.Context, id int, grund string, benutzer string) error {
	log := zerolog.Ctx(ctx)

	if err := c.TSERepo.TSESignaturauftragVerwerfen(ctx, id, grund, benutzer); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to discard tse signaturauftrag")
		return ErrDatabase
	}

	log.Info().Int("id", id).Str("benutzer", benutzer).Msg("TSE-Signaturauftrag verworfen")
	return nil
}
