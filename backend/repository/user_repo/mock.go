//go:build unit

package user_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

func NewMock(users []user.User, err error) *mockRepo {
	userMap := make(map[int]user.User)
	for i := range users {
		userMap[users[i].ID] = users[i]
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
	for _, u := range m.user { //nolint:gocritic // iterating small map for lookup
		if u.Username == username {
			return u, m.err
		}
	}
	return user.User{}, m.err
}

func (m mockRepo) GetAllUsers(ctx context.Context) ([]user.User, error) {
	users := []user.User{}
	for _, u := range m.user { //nolint:gocritic // collecting all values
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

// SetPasswordTx spiegelt den transaktionalen Repo-Pfad in-memory: Benutzer laden,
// apply ausführen, Ergebnis persistieren. Der Fachfehler aus apply wird nach der
// (simulierten) Persistenz zurückgegeben.
func (m mockRepo) SetPasswordTx(ctx context.Context, username string, apply func(*user.User) error) error {
	if m.err != nil {
		return m.err
	}
	for id, u := range m.user { //nolint:gocritic // iterating small map for lookup
		if u.Username == username {
			applyErr := apply(&u)
			m.user[id] = u
			return applyErr
		}
	}
	return db.ErrNotFound
}
