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

func TestGetTischSaldo(t *testing.T) {
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(1, table.TischState{
		SaldoCents:           700,
		GesamtZahlungenCents: 0,
	})
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: eventMock,
	}

	saldo, err := query.GetTischSaldo(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if saldo != 700 {
		t.Errorf("expected saldo 700, got %d", saldo)
	}
}

func TestGetTischSaldo_NoState(t *testing.T) {
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	saldo, err := query.GetTischSaldo(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if saldo != 0 {
		t.Errorf("expected saldo 0, got %d", saldo)
	}
}

func TestGetTischUnbezahlt(t *testing.T) {
	positions := []table.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", Einzelpreis: 350, Menge: 2},
	}
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(1, table.TischState{
		SaldoCents:           700,
		UnbezahltePositionen: positions,
	})
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: eventMock,
	}

	unbezahlt, err := query.GetTischUnbezahlt(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(unbezahlt) != 1 {
		t.Fatalf("expected 1 position, got %d", len(unbezahlt))
	}
	if unbezahlt[0].Menge != 2 {
		t.Errorf("expected menge 2, got %d", unbezahlt[0].Menge)
	}
}

func TestGetTischUnbezahlt_NoState(t *testing.T) {
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	unbezahlt, err := query.GetTischUnbezahlt(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if unbezahlt != nil {
		t.Errorf("expected nil, got %v", unbezahlt)
	}
}

func TestGetTischUngeliefert(t *testing.T) {
	positions := []table.Position{
		{PositionID: "p1", ProduktName: "Pommes", VarianteName: "groß", Einzelpreis: 500, Menge: 1},
	}
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(1, table.TischState{
		SaldoCents:             500,
		UngeliefertePositionen: positions,
	})
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: eventMock,
	}

	ungeliefert, err := query.GetTischUngeliefert(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ungeliefert) != 1 {
		t.Fatalf("expected 1 position, got %d", len(ungeliefert))
	}
	if ungeliefert[0].ProduktName != "Pommes" {
		t.Errorf("expected ProduktName 'Pommes', got %s", ungeliefert[0].ProduktName)
	}
}

func TestGetTischUngeliefert_NoState(t *testing.T) {
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	ungeliefert, err := query.GetTischUngeliefert(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ungeliefert != nil {
		t.Errorf("expected nil, got %v", ungeliefert)
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
