//go:build unit

package kassensitzungen_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

// NewMock creates a new mock repository with an optional open Kassensitzung and error.
func NewMock(offeneKS *kasse.Kassensitzung, err error) *MockRepo {
	return &MockRepo{
		offeneKS: offeneKS,
		err:      err,
	}
}

type MockRepo struct {
	offeneKS *kasse.Kassensitzung
	err      error

	// Abschluss-Barriere: Aufrufzähler, um Reset/Resume zu testen. Die Setter mutieren den
	// übergebenen Kassensitzungs-Zeiger bewusst nicht (er ist in Tests oft ein geteiltes Fixture).
	WirdAbgeschlossenCalls int
	OffenCalls             int
}

func (m *MockRepo) GetOffeneKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.offeneKS != nil && m.offeneKS.Status != kasse.KassensitzungOffen {
		return nil, nil
	}
	return m.offeneKS, nil
}

// GetAktiveKassensitzung returns the mock Kassensitzung when it is 'offen' or 'wird_abgeschlossen'
// (both count as active) and nil when it is closed.
func (m *MockRepo) GetAktiveKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.offeneKS != nil && m.offeneKS.Status == kasse.KassensitzungAbgeschlossen {
		return nil, nil
	}
	return m.offeneKS, nil
}

// SetKassensitzungWirdAbgeschlossen records the barrier call and reports one affected row.
func (m *MockRepo) SetKassensitzungWirdAbgeschlossen(_ context.Context, _ int) (int64, error) {
	m.WirdAbgeschlossenCalls++
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

// SetKassensitzungOffen records the reset call and reports one affected row.
func (m *MockRepo) SetKassensitzungOffen(_ context.Context, _ int) (int64, error) {
	m.OffenCalls++
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *MockRepo) GetOffeneKassensitzungNr(_ context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.offeneKS == nil {
		return 0, nil
	}
	return m.offeneKS.ZNr, nil
}

func (m *MockRepo) GetAllKassensitzungen(_ context.Context) ([]kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.offeneKS != nil {
		return []kasse.Kassensitzung{*m.offeneKS}, nil
	}
	return []kasse.Kassensitzung{}, nil
}

// SetOffeneKassensitzung sets the open Kassensitzung for the mock.
func (m *MockRepo) SetOffeneKassensitzung(ks *kasse.Kassensitzung) {
	m.offeneKS = ks
}
