package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type tseSignaturauftragQueryRepo interface {
	GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error)
}

type Query struct {
	TSERepo tseSignaturauftragQueryRepo
}

// GetTSESignaturauftraege liefert die Signaturauftraege fuer die
// Admin-Verwaltung; die Liste dient zugleich als TSE-Ausfalldokumentation
// (AEAO zu § 146a, 1.14.1).
func (q Query) GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error) {
	log := zerolog.Ctx(ctx)

	auftraege, err := q.TSERepo.GetTSESignaturauftraege(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse signaturauftraege")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(auftraege)).Msg("Retrieved tse signaturauftraege")
	return auftraege, nil
}
