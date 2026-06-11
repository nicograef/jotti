//go:build unit

package druckauftrag_repo

import "context"

func NewMock(offene []OffenerDruckauftrag, err error) *mockRepo {
	return &mockRepo{offene: offene, err: err}
}

type mockRepo struct {
	offene       []OffenerDruckauftrag
	enqueued     []NeuerDruckauftrag
	gedruckt     []int
	fehlversuche []Fehlversuch
	err          error
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

func (m *mockRepo) MeldeDruckergebnis(_ context.Context, gedruckteIDs []int, fehlversuche []Fehlversuch) error {
	if m.err != nil {
		return m.err
	}
	m.gedruckt = append(m.gedruckt, gedruckteIDs...)
	m.fehlversuche = append(m.fehlversuche, fehlversuche...)
	return nil
}

func (m *mockRepo) Enqueued() []NeuerDruckauftrag {
	return m.enqueued
}
