package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type tseNachsignierQueryRepo interface {
	GetTSENachsignierAuftraege(ctx context.Context) ([]tse_repo.NachsignierAuftrag, error)
}

type Query struct {
	TSERepo tseNachsignierQueryRepo
}

// GetTSENachsignierAuftraege liefert die Nachsignier-Auftraege fuer die
// Admin-Verwaltung; die Liste dient zugleich als TSE-Ausfalldokumentation
// (AEAO zu § 146a, 1.14.1).
func (q Query) GetTSENachsignierAuftraege(ctx context.Context) ([]tse_repo.NachsignierAuftrag, error) {
	log := zerolog.Ctx(ctx)

	auftraege, err := q.TSERepo.GetTSENachsignierAuftraege(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse nachsignier auftraege")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(auftraege)).Msg("Retrieved tse nachsignier auftraege")
	return auftraege, nil
}
