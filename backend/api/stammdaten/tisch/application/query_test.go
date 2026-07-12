//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

func TestGetAllTische(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus}}, nil)
	query := Query{TischRepo: repo}

	tische, err := query.GetAllTische(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tische) != 1 {
		t.Fatalf("expected 1 tisch, got %d", len(tische))
	}
	if tische[0].Tisch.Name != "Tisch 1" {
		t.Errorf("expected name 'Tisch 1', got %s", tische[0].Tisch.Name)
	}
	if tische[0].SaldoCents != 0 {
		t.Errorf("expected saldoCents 0 without open saldo, got %d", tische[0].SaldoCents)
	}
}

// TestGetAllTische_SaldoAusOffenerSitzung deckt die saldoCents-Projektion gegen
// die tisch_sessions der offenen Kassensitzung ab: nur der Tisch mit offenem
// Saldo trägt den Betrag, alle anderen bleiben bei 0.
func TestGetAllTische_SaldoAusOffenerSitzung(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{
		{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus},
		{ID: 2, Name: "Tisch 2", Status: tisch.ActiveStatus},
	}, nil)
	repo.SetOffenerSaldo(1, 9850)
	query := Query{TischRepo: repo}

	tische, err := query.GetAllTische(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	saldi := make(map[int]int, len(tische))
	for _, t := range tische {
		saldi[t.Tisch.ID] = t.SaldoCents
	}
	if saldi[1] != 9850 {
		t.Errorf("expected tisch 1 saldoCents 9850, got %d", saldi[1])
	}
	if saldi[2] != 0 {
		t.Errorf("expected tisch 2 saldoCents 0, got %d", saldi[2])
	}
}
