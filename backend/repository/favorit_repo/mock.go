//go:build unit

package favorit_repo

import (
	"context"
	"slices"
)

// NewMock erzeugt ein In-Memory-Favoriten-Repository für Unit-Tests. favoriten
// bildet Benutzer-ID auf die markierten Tisch-IDs ab; err wird von jeder Methode
// zurückgegeben, die Zustandsänderung unterbleibt dann.
func NewMock(favoriten map[int][]int, err error) *mockRepo {
	kopie := make(map[int][]int, len(favoriten))
	for userID, tischIDs := range favoriten {
		kopie[userID] = slices.Clone(tischIDs)
	}

	return &mockRepo{favoriten: kopie, err: err}
}

type mockRepo struct {
	favoriten map[int][]int
	err       error
}

func (m *mockRepo) Add(_ context.Context, userID, tischID int) error {
	if m.err != nil {
		return m.err
	}
	if !slices.Contains(m.favoriten[userID], tischID) {
		m.favoriten[userID] = append(m.favoriten[userID], tischID)
	}
	return nil
}

func (m *mockRepo) Remove(_ context.Context, userID, tischID int) error {
	if m.err != nil {
		return m.err
	}
	m.favoriten[userID] = slices.DeleteFunc(m.favoriten[userID], func(id int) bool { return id == tischID })
	return nil
}

func (m *mockRepo) RemoveByTisch(_ context.Context, tischID int) error {
	if m.err != nil {
		return m.err
	}
	for userID := range m.favoriten {
		m.favoriten[userID] = slices.DeleteFunc(m.favoriten[userID], func(id int) bool { return id == tischID })
	}
	return nil
}

func (m *mockRepo) GetByUser(_ context.Context, userID int) ([]int, error) {
	if m.err != nil {
		return nil, m.err
	}
	return slices.Clone(m.favoriten[userID]), nil
}
