package repotest

import (
	"context"
	"slices"
)

// NewFavoritRepo creates an in-memory Favoriten repository that maps a user ID to the marked Tisch IDs.
// A set err is returned by every method, which then changes nothing.
func NewFavoritRepo(favoriten map[int][]int, err error) *FavoritRepo {
	kopie := make(map[int][]int, len(favoriten))
	for userID, tischIDs := range favoriten {
		kopie[userID] = slices.Clone(tischIDs)
	}

	return &FavoritRepo{favoriten: kopie, err: err}
}

type FavoritRepo struct {
	favoriten map[int][]int
	err       error
}

func (m *FavoritRepo) Add(_ context.Context, userID, tischID int) error {
	if m.err != nil {
		return m.err
	}
	if !slices.Contains(m.favoriten[userID], tischID) {
		m.favoriten[userID] = append(m.favoriten[userID], tischID)
	}
	return nil
}

func (m *FavoritRepo) Remove(_ context.Context, userID, tischID int) error {
	if m.err != nil {
		return m.err
	}
	m.favoriten[userID] = slices.DeleteFunc(m.favoriten[userID], func(id int) bool { return id == tischID })
	return nil
}

func (m *FavoritRepo) RemoveByTisch(_ context.Context, tischID int) error {
	if m.err != nil {
		return m.err
	}
	for userID := range m.favoriten {
		m.favoriten[userID] = slices.DeleteFunc(m.favoriten[userID], func(id int) bool { return id == tischID })
	}
	return nil
}

func (m *FavoritRepo) GetByUser(_ context.Context, userID int) ([]int, error) {
	if m.err != nil {
		return nil, m.err
	}
	return slices.Clone(m.favoriten[userID]), nil
}
