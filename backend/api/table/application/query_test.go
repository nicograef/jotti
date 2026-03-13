//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/event_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func TestGetAllTische(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{table.Tisch{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus}}, nil)
	query := Query{TableRepo: repo}

	tische, err := query.GetAllTische(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tische) != 1 {
		t.Fatalf("expected 1 tisch, got %d", len(tische))
	}
	if tische[0].Name != "Tisch 1" {
		t.Errorf("expected name 'Tisch 1', got %s", tische[0].Name)
	}
}

func TestGetTischState(t *testing.T) {
	positions := []table.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", Einzelpreis: 350, Menge: 2},
	}
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(1, table.TischState{
		SaldoCents:             700,
		UnbezahltePositionen:   positions,
		UngeliefertePositionen: positions,
		GesamtZahlungenCents:   0,
	})
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: eventMock,
	}

	state, err := query.GetTischState(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 700 {
		t.Errorf("expected saldo 700, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Errorf("expected menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if len(state.UngeliefertePositionen) != 1 {
		t.Fatalf("expected 1 ungelieferte position, got %d", len(state.UngeliefertePositionen))
	}
}

func TestGetTischState_NoState(t *testing.T) {
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	state, err := query.GetTischState(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Errorf("expected saldo 0, got %d", state.SaldoCents)
	}
	if state.UnbezahltePositionen != nil {
		t.Errorf("expected nil unbezahlt, got %v", state.UnbezahltePositionen)
	}
	if state.UngeliefertePositionen != nil {
		t.Errorf("expected nil ungeliefert, got %v", state.UngeliefertePositionen)
	}
}

func TestGetTischHistorie_UsesReadEventsBySubject(t *testing.T) {
	// Historie should work via ReadEventsBySubject (event stream), not ReadTableState
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	historie, err := query.GetTischHistorie(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 0 {
		t.Errorf("expected empty historie, got %d entries", len(historie))
	}
}
