package user

import (
	"context"

	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetUserID(ctx context.Context, username string) (int, error)
	GetUser(ctx context.Context, id int) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
}

// Query provides user-related operations.
type Query struct {
	Persistence queryPersistence
}

// GetAllUsers retrieves all users from the database.
func (s *Query) GetAllUsers(ctx context.Context) ([]*User, error) {
	log := zerolog.Ctx(ctx)

	users, err := s.Persistence.GetAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all users")
		return nil, ErrDatabase
	}

	return users, nil
}
