package user_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/user"
)

func NewMock(users []user.User, err error) *mockRepo {
	userMap := make(map[int]user.User)
	for _, t := range users {
		userMap[t.ID] = t
	}

	return &mockRepo{
		user: userMap,
		err:  err,
	}
}

type mockRepo struct {
	user map[int]user.User
	err  error
}

func (m mockRepo) GetUser(ctx context.Context, id int) (user.User, error) {
	t, ok := m.user[id]
	if !ok {
		return user.User{}, m.err
	}
	return t, m.err
}

func (m mockRepo) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	for _, u := range m.user {
		if u.Username == username {
			return u, m.err
		}
	}
	return user.User{}, m.err
}

func (m mockRepo) GetAllUsers(ctx context.Context) ([]user.User, error) {
	users := []user.User{}
	for _, u := range m.user {
		users = append(users, u)
	}
	return users, m.err
}

func (m mockRepo) CreateUser(ctx context.Context, t user.User) (int, error) {
	newID := len(m.user) + 1
	t.ID = newID
	m.user[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateUser(ctx context.Context, t user.User) error {
	m.user[t.ID] = t
	return m.err
}
