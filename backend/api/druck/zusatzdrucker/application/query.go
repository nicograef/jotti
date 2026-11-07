package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type zusatzdruckerQueryRepo interface {
	GetAlleZusatzdrucker(ctx context.Context) ([]druckstation.Zusatzdrucker, error)
}

type Query struct {
	ZusatzdruckerRepo zusatzdruckerQueryRepo
}

func (q Query) GetAlle(ctx context.Context) ([]druckstation.Zusatzdrucker, error) {
	log := logging.FromContext(ctx)

	alle, err := q.ZusatzdruckerRepo.GetAlleZusatzdrucker(ctx)
	if err != nil {
		log.Error("Failed to retrieve zusatzdrucker", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved zusatzdrucker", "count", len(alle))
	return alle, nil
}
