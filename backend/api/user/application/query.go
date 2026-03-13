package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/rs/zerolog"
)

type userQueryRepo interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

func GetAllUsers(ctx context.Context, repo userQueryRepo) ([]user.User, error) {
	log := zerolog.Ctx(ctx)

	users, err := repo.GetAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all users")
		return nil, ErrDatabase
	}

	return users, nil
}
