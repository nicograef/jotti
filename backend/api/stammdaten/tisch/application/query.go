package application

import (
	"context"

	t "github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type tischQueryRepo interface {
	GetAlleTische(ctx context.Context) ([]t.Tisch, error)
	GetTischSaldiOffeneSitzung(ctx context.Context) (map[int]int, error)
}

type Query struct {
	TischRepo tischQueryRepo
}

// TischMitSaldo ergänzt den offenen Saldo der aktuell offenen Kassensitzung
// (0 ohne offenen Saldo oder ohne offene Sitzung). Der Saldo ist keine
// Domäneneigenschaft, sondern eine tisch_sessions-Projektion.
type TischMitSaldo struct {
	Tisch      t.Tisch
	SaldoCents int
}

func (q Query) GetAllTische(ctx context.Context) ([]TischMitSaldo, error) {
	log := logging.FromContext(ctx)

	tische, err := q.TischRepo.GetAlleTische(ctx)
	if err != nil {
		log.Error("Failed to retrieve all tische", "error", err)
		return nil, ErrDatabase
	}

	saldi, err := q.TischRepo.GetTischSaldiOffeneSitzung(ctx)
	if err != nil {
		log.Error("Failed to retrieve open tisch saldi", "error", err)
		return nil, ErrDatabase
	}

	result := make([]TischMitSaldo, 0, len(tische))
	for i := range tische {
		result = append(result, TischMitSaldo{
			Tisch:      tische[i],
			SaldoCents: saldi[tische[i].ID],
		})
	}

	log.Debug("Retrieved all tische", "count", len(result))
	return result, nil
}
