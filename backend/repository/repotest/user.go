package repotest

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

func NewUserRepo(users []user.User, err error) *UserRepo {
	userMap := make(map[int]user.User)
	for i := range users {
		userMap[users[i].ID] = users[i]
	}

	return &UserRepo{
		user: userMap,
		err:  err,
	}
}

type UserRepo struct {
	user map[int]user.User
	err  error
}

func (m UserRepo) GetUser(ctx context.Context, id int) (user.User, error) {
	if m.err != nil {
		return user.User{}, m.err
	}
	t, ok := m.user[id]
	if !ok {
		return user.User{}, db.ErrNotFound
	}
	return t, nil
}

func (m UserRepo) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	if m.err != nil {
		return user.User{}, m.err
	}
	for _, u := range m.user { //nolint:gocritic // iterating small map for lookup
		if u.Username == username {
			return u, nil
		}
	}
	return user.User{}, db.ErrNotFound
}

func (m UserRepo) GetAllUsers(ctx context.Context) ([]user.User, error) {
	users := []user.User{}
	for _, u := range m.user { //nolint:gocritic // collecting all values
		users = append(users, u)
	}
	return users, m.err
}

func (m UserRepo) CountUsers(ctx context.Context) (int, error) {
	return len(m.user), m.err
}

func (m UserRepo) CreateUser(ctx context.Context, t user.User) (int, error) {
	newID := len(m.user) + 1
	t.ID = newID
	m.user[newID] = t
	return newID, m.err
}

func (m UserRepo) UpdateUser(ctx context.Context, t user.User) error {
	m.user[t.ID] = t
	return m.err
}

// SetPasswordTx spiegelt den transaktionalen Repo-Pfad in-memory: Benutzer laden,
// apply ausführen, Ergebnis persistieren. Der Fachfehler aus apply wird nach der
// (simulierten) Persistenz zurückgegeben.
func (m UserRepo) SetPasswordTx(ctx context.Context, username string, apply func(*user.User) error) error {
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
