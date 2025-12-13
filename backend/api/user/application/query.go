package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type Query struct {
	UserRepo queryPersistence
}

func (q Query) GetAllUsers(ctx context.Context) ([]user.User, error) {
	log := zerolog.Ctx(ctx)

	users, err := q.UserRepo.GetAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all users")
		return nil, ErrDatabase
	}

	return users, nil
}
