package repotest

import (
	"context"
	"slices"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

// NewKassensitzungenRepo creates a fake holding a copy of ks (nil: no Kassensitzung) that
// returns err from every read and write.
func NewKassensitzungenRepo(ks *kasse.Kassensitzung, err error) *KassensitzungenRepo {
	m := &KassensitzungenRepo{err: err}
	m.SetOffeneKassensitzung(ks)
	return m
}

// KassensitzungenRepo keeps its own copy of the Kassensitzung, so the status transitions
// never touch a fixture shared between tests.
type KassensitzungenRepo struct {
	offeneKS *kasse.Kassensitzung
	err      error
}

// GetAktiveKassensitzung returns a copy of the Kassensitzung when it is 'offen' or
// 'wird_abgeschlossen' (both count as active) and nil when it is closed.
func (m *KassensitzungenRepo) GetAktiveKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.offeneKS == nil || m.offeneKS.Status == kasse.KassensitzungAbgeschlossen {
		return nil, nil
	}
	ks := *m.offeneKS
	return &ks, nil
}

// SetKassensitzungWirdAbgeschlossen sets the barrier like its query: only an active
// Kassensitzung with this zNr changes, and the affected row count reports it.
func (m *KassensitzungenRepo) SetKassensitzungWirdAbgeschlossen(_ context.Context, zNr int) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.setStatus(zNr, kasse.KassensitzungWirdAbgeschlossen, kasse.KassensitzungOffen, kasse.KassensitzungWirdAbgeschlossen), nil
}

// SetKassensitzungOffen resets the barrier like its query: only from 'wird_abgeschlossen'.
func (m *KassensitzungenRepo) SetKassensitzungOffen(_ context.Context, zNr int) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.setStatus(zNr, kasse.KassensitzungOffen, kasse.KassensitzungWirdAbgeschlossen), nil
}

func (m *KassensitzungenRepo) setStatus(zNr int, neu kasse.KassensitzungStatus, von ...kasse.KassensitzungStatus) int64 {
	if m.offeneKS == nil || m.offeneKS.ZNr != zNr || !slices.Contains(von, m.offeneKS.Status) {
		return 0
	}
	m.offeneKS.Status = neu
	return 1
}

func (m *KassensitzungenRepo) GetOffeneKassensitzungNr(_ context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.offeneKS == nil {
		return 0, nil
	}
	return m.offeneKS.ZNr, nil
}

func (m *KassensitzungenRepo) GetAllKassensitzungen(_ context.Context) ([]kasse.Kassensitzung, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.offeneKS != nil {
		return []kasse.Kassensitzung{*m.offeneKS}, nil
	}
	return []kasse.Kassensitzung{}, nil
}

// SetOffeneKassensitzung replaces the fake's Kassensitzung with a copy of ks (nil: none).
func (m *KassensitzungenRepo) SetOffeneKassensitzung(ks *kasse.Kassensitzung) {
	m.offeneKS = nil
	if ks != nil {
		kopie := *ks
		m.offeneKS = &kopie
	}
}
