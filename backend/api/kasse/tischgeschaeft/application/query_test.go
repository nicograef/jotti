//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func TestGetTischState(t *testing.T) {
	positions := []kasse.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", Einzelpreis: 350, Menge: 2, BestellerUserID: 5, BestellerName: "Anna"},
	}
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	subject := kasse.TischSessionSubject(1, 1)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:            700,
		UnbezahltePositionen:  positions,
		AusstehendePositionen: positions,
		GesamtZahlungenCents:  0,
	})
	query := Query{
		TableRepo:           table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus}}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	state, err := query.GetTischState(context.Background(), 1, 5)
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
	if state.UnbezahltePositionen[0].BestellerName != "Anna" {
		t.Errorf("expected besteller Anna, got %q", state.UnbezahltePositionen[0].BestellerName)
	}
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("expected 1 ausstehende position, got %d", len(state.AusstehendePositionen))
	}
	// Anna hat eigene offene Positionen an diesem Tisch -> nicht erledigt.
	if state.FuerMichErledigt {
		t.Errorf("expected FuerMichErledigt false for besteller with open positions")
	}

	// Eine andere Servicekraft (ohne eigene Positionen) sieht den Tisch als erledigt.
	stateAndere, err := query.GetTischState(context.Background(), 1, 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !stateAndere.FuerMichErledigt {
		t.Errorf("expected FuerMichErledigt true for servicekraft without own positions")
	}
}

func TestGetTischState_NoState(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	query := Query{
		TableRepo:           table_repo.NewMock([]table.Tisch{{ID: 999, Name: "Tisch 999", Status: table.ActiveStatus}}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	state, err := query.GetTischState(context.Background(), 999, 1)
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

func TestGetTischHistorie_ReturnsEmptyForTischWithNoEvents(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	query := Query{
		TableRepo:           table_repo.NewMock(nil, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	historie, err := query.GetTischHistorie(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 0 {
		t.Errorf("expected empty historie, got %d entries", len(historie))
	}
}
