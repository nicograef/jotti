package repotest

import (
	"context"
	"maps"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

// NewTischRepo creates a new mock repository with the given tische and error.
func NewTischRepo(tische []tisch.Tisch, err error) *TischRepo {
	tischMap := make(map[int]tisch.Tisch)
	for _, t := range tische {
		tischMap[t.ID] = t
	}

	return &TischRepo{
		tische:      tischMap,
		offeneSaldi: make(map[int]int),
		err:         err,
	}
}

type TischRepo struct {
	tische map[int]tisch.Tisch
	// offeneSaldi enthält die offenen Saldi (tischID → saldoCents) der offenen
	// Kassensitzung — für die saldoCents-Projektion und den Schutz-Guard.
	offeneSaldi map[int]int
	// favoritenCleanup ist der Anteil an AendereStatus, den das echte Repository
	// beim Löschen in derselben Transaktion wie den Statuswechsel ausführt.
	favoritenCleanup func(ctx context.Context, tischID int) error
	err              error
}

// SetFavoritenCleanup sets the cleanup AendereStatus runs on delete; if it fails, the status change
// is skipped, mirroring the real transaction's rollback.
func (m *TischRepo) SetFavoritenCleanup(cleanup func(ctx context.Context, tischID int) error) {
	m.favoritenCleanup = cleanup
}

// SetOffenerSaldo markiert einen Tisch mit einem offenen Saldo in der offenen
// Kassensitzung (Testhilfe für den tisch_saldo_offen-Pfad).
func (m *TischRepo) SetOffenerSaldo(tischID, saldoCents int) {
	m.offeneSaldi[tischID] = saldoCents
}

func (m *TischRepo) GetTischSaldiOffeneSitzung(ctx context.Context) (map[int]int, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]int, len(m.offeneSaldi))
	maps.Copy(result, m.offeneSaldi)
	return result, nil
}

func (m *TischRepo) GetTisch(ctx context.Context, id int) (tisch.Tisch, error) {
	if m.err != nil {
		return tisch.Tisch{}, m.err
	}
	t, ok := m.tische[id]
	if !ok {
		return tisch.Tisch{}, db.ErrNotFound
	}
	return t, nil
}

func (m *TischRepo) GetAlleTische(ctx context.Context) ([]tisch.Tisch, error) {
	var result []tisch.Tisch
	for _, t := range m.tische {
		result = append(result, t)
	}
	return result, m.err
}

func (m *TischRepo) GetAktiveTische(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error) {
	var result []tisch.AktiverTisch
	for _, t := range m.tische {
		if t.Status == tisch.ActiveStatus {
			result = append(result, tisch.AktiverTisch{ID: t.ID, Name: t.Name, SaldoCents: 0})
		}
	}
	return result, m.err
}

func (m *TischRepo) CreateTisch(ctx context.Context, t tisch.Tisch) (int, error) {
	newID := len(m.tische) + 1
	t.ID = newID
	m.tische[newID] = t
	return newID, m.err
}

func (m *TischRepo) UpdateTisch(ctx context.Context, t tisch.Tisch) error {
	m.tische[t.ID] = t
	return m.err
}

// AendereStatus mirrors the real transaction: an open saldo refuses the change under saldoSchutz, and a
// failing Favoriten cleanup of a deleted Tisch skips it.
func (m *TischRepo) AendereStatus(ctx context.Context, t tisch.Tisch, saldoSchutz bool) error {
	if m.err != nil {
		return m.err
	}
	if saldoSchutz && m.offeneSaldi[t.ID] > 0 {
		return tisch.ErrOffenerSaldo
	}
	if t.Status == tisch.DeletedStatus && m.favoritenCleanup != nil {
		if err := m.favoritenCleanup(ctx, t.ID); err != nil {
			return err
		}
	}
	m.tische[t.ID] = t
	return nil
}

func (m *TischRepo) GetAktiveTischeMitFavoriten(_ context.Context, _ int, _ int) ([]tisch.AktiverTischMitFavorit, error) {
	return nil, m.err
}
