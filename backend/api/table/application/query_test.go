//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func TestGetAllTische(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus}}, nil)
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
	positions := []kasse.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", Einzelpreis: 350, Menge: 2},
	}
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetOffeneKassensitzung(&kasse.KassensitzungState{ZNr: 1, Status: kasse.KassensitzungStatusOffen})
	subject := kasse.TischSessionSubject(1, 1)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:            700,
		UnbezahltePositionen:  positions,
		AusstehendePositionen: positions,
		GesamtZahlungenCents:  0,
	})
	query := Query{
		TableRepo: table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus}}, nil),
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
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("expected 1 ausstehende position, got %d", len(state.AusstehendePositionen))
	}
}

func TestGetTischState_NoState(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetOffeneKassensitzung(&kasse.KassensitzungState{ZNr: 1, Status: kasse.KassensitzungStatusOffen})
	query := Query{
		TableRepo: table_repo.NewMock([]table.Tisch{{ID: 999, Name: "Tisch 999", Status: table.ActiveStatus}}, nil),
		EventRepo: eventMock,
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
	if state.AusstehendePositionen != nil {
		t.Errorf("expected nil ausstehend, got %v", state.AusstehendePositionen)
	}
}

func TestGetTischHistorie_UsesReadEventsBySubject(t *testing.T) {
	// Historie should work via ReadEventsBySubject (event stream), not ReadTableState
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetOffeneKassensitzung(&kasse.KassensitzungState{ZNr: 1, Status: kasse.KassensitzungStatusOffen})
	query := Query{
		TableRepo: table_repo.NewMock(nil, nil),
		EventRepo: eventMock,
	}

	historie, err := query.GetTischHistorie(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 0 {
		t.Errorf("expected empty historie, got %d entries", len(historie))
	}
}
