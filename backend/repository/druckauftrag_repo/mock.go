//go:build unit

package druckauftrag_repo

import "context"

func NewMock(offene []OffenerDruckauftrag, err error) *mockRepo {
	return &mockRepo{offene: offene, err: err}
}

type mockRepo struct {
	offene    []OffenerDruckauftrag
	enqueued  []NeuerDruckauftrag
	quittiert []int
	err       error
}

func (m *mockRepo) EnqueueDruckauftraege(_ context.Context, auftraege []NeuerDruckauftrag) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, auftraege...)
	return nil
}

func (m *mockRepo) GetOffeneDruckauftraege(_ context.Context) ([]OffenerDruckauftrag, error) {
	return m.offene, m.err
}

func (m *mockRepo) QuittiereGedruckteAuftraege(_ context.Context, ids []int) error {
	if m.err != nil {
		return m.err
	}
	m.quittiert = append(m.quittiert, ids...)
	return nil
}

func (m *mockRepo) Enqueued() []NeuerDruckauftrag {
	return m.enqueued
}
