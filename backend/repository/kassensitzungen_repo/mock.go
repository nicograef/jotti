//go:build unit

package kassensitzungen_repo

import (
	"context"
	"time"

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
}

func (m *MockRepo) GetOffeneKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.offeneKS, nil
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

func (m *MockRepo) InsertKassensitzung(_ context.Context, _ time.Time, _ string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
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
