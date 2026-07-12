package application

import (
	"context"

	t "github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

type tischQueryRepo interface {
	GetAllTables(ctx context.Context) ([]t.Tisch, error)
	GetTischSaldiOffeneSitzung(ctx context.Context) (map[int]int, error)
}

type Query struct {
	TischRepo tischQueryRepo
}

// TischMitSaldo ergänzt einen Tisch um den offenen Saldo der aktuell offenen
// Kassensitzung (0 ohne offenen Saldo oder ohne offene Sitzung). Der Saldo ist
// keine Domäneneigenschaft, sondern eine tisch_sessions-Projektion, und lebt
// deshalb hier statt am Domain-Modell.
type TischMitSaldo struct {
	Tisch      t.Tisch
	SaldoCents int
}

func (q Query) GetAllTische(ctx context.Context) ([]TischMitSaldo, error) {
	log := zerolog.Ctx(ctx)

	tische, err := q.TischRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all tische")
		return nil, ErrDatabase
	}

	saldi, err := q.TischRepo.GetTischSaldiOffeneSitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve open tisch saldi")
		return nil, ErrDatabase
	}

	result := make([]TischMitSaldo, 0, len(tische))
	for i := range tische {
		result = append(result, TischMitSaldo{
			Tisch:      tische[i],
			SaldoCents: saldi[tische[i].ID],
		})
	}

	log.Debug().Int("count", len(result)).Msg("Retrieved all tische")
	return result, nil
}
