package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type betreiberRepo interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

type Query struct {
	BetreiberRepo betreiberRepo
}

func (q Query) GetBetreiber(ctx context.Context) (betreiber.Betreiber, error) {
	log := logging.FromContext(ctx)

	b, err := q.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return betreiber.Betreiber{}, ErrNotFound
		}
		log.Error("Failed to retrieve betreiber", "error", err)
		return betreiber.Betreiber{}, ErrDatabase
	}
	return b, nil
}
