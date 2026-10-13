package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type userQueryRepo interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type Query struct {
	UserRepo userQueryRepo
}

func (q Query) GetAllUsers(ctx context.Context) ([]user.User, error) {
	log := logging.FromContext(ctx)

	users, err := q.UserRepo.GetAllUsers(ctx)
	if err != nil {
		log.Error("Failed to retrieve all users", "error", err)
		return nil, ErrDatabase
	}

	return users, nil
}
