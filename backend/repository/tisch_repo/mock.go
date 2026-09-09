//go:build unit

package tisch_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

// NewMock creates a new mock repository with the given tische and error.
func NewMock(tische []tisch.Tisch, err error) *mockRepo {
	tischMap := make(map[int]tisch.Tisch)
	for _, t := range tische {
		tischMap[t.ID] = t
	}

	return &mockRepo{
		tische:      tischMap,
		offeneSaldi: make(map[int]int),
		err:         err,
	}
}

type mockRepo struct {
	tische map[int]tisch.Tisch
	// offeneSaldi enthält die offenen Saldi (tischID → saldoCents) der offenen
	// Kassensitzung — für die saldoCents-Projektion und den Schutz-Guard.
	offeneSaldi map[int]int
	// favoritenCleanup ist der Anteil an DeleteTischMitFavoriten, den das echte
	// Repository in derselben Transaktion wie den Statuswechsel ausführt.
	favoritenCleanup func(ctx context.Context, tischID int) error
	err              error
}

// SetFavoritenCleanup hinterlegt den Favoriten-Cleanup, den
// DeleteTischMitFavoriten mit dem Statuswechsel zusammen ausführt. Scheitert er,
// unterbleibt der Statuswechsel — wie beim Rollback der echten Transaktion.
// Ohne hinterlegten Cleanup löscht der Mock nur den Tisch.
func (m *mockRepo) SetFavoritenCleanup(cleanup func(ctx context.Context, tischID int) error) {
	m.favoritenCleanup = cleanup
}

// SetOffenerSaldo markiert einen Tisch mit einem offenen Saldo in der offenen
// Kassensitzung (Testhilfe für den tisch_saldo_offen-Pfad).
func (m *mockRepo) SetOffenerSaldo(tischID, saldoCents int) {
	m.offeneSaldi[tischID] = saldoCents
}

func (m *mockRepo) GetTischSaldiOffeneSitzung(ctx context.Context) (map[int]int, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]int, len(m.offeneSaldi))
	for id, saldo := range m.offeneSaldi {
		result[id] = saldo
	}
	return result, nil
}

func (m *mockRepo) TischHatOffenenSaldo(ctx context.Context, tischID int) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.offeneSaldi[tischID] > 0, nil
}

func (m mockRepo) GetTisch(ctx context.Context, id int) (tisch.Tisch, error) {
	if m.err != nil {
		return tisch.Tisch{}, m.err
	}
	t, ok := m.tische[id]
	if !ok {
		return tisch.Tisch{}, db.ErrNotFound
	}
	return t, nil
}

func (m mockRepo) GetAlleTische(ctx context.Context) ([]tisch.Tisch, error) {
	var result []tisch.Tisch
	for _, t := range m.tische {
		result = append(result, t)
	}
	return result, m.err
}

func (m mockRepo) GetAktiveTische(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error) {
	var result []tisch.AktiverTisch
	for _, t := range m.tische {
		if t.Status == tisch.ActiveStatus {
			result = append(result, tisch.AktiverTisch{ID: t.ID, Name: t.Name, SaldoCents: 0})
		}
	}
	return result, m.err
}

func (m mockRepo) CreateTisch(ctx context.Context, t tisch.Tisch) (int, error) {
	newID := len(m.tische) + 1
	t.ID = newID
	m.tische[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateTisch(ctx context.Context, t tisch.Tisch) error {
	m.tische[t.ID] = t
	return m.err
}

func (m mockRepo) DeleteTischMitFavoriten(ctx context.Context, t tisch.Tisch) error {
	if m.err != nil {
		return m.err
	}
	if m.favoritenCleanup != nil {
		if err := m.favoritenCleanup(ctx, t.ID); err != nil {
			return err
		}
	}
	m.tische[t.ID] = t
	return nil
}

func (m mockRepo) GetAktiveTischeMitFavoriten(_ context.Context, _ int, _ int) ([]tisch.AktiverTischMitFavorit, error) {
	return nil, m.err
}
