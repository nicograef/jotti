//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

type mockHistorieRepo struct {
	events []event.Event
	err    error
}

func (m *mockHistorieRepo) ReadDirektverkaufEvents(_ context.Context, _ int) ([]event.Event, error) {
	return m.events, m.err
}

func TestGetDirektverkaufHistorie_NoOpenKassensitzung_ReturnsEmpty(t *testing.T) {
	query := Query{
		EventRepo:           &mockHistorieRepo{},
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
	}

	historie, err := query.GetDirektverkaufHistorie(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 0 {
		t.Errorf("expected empty historie, got %d entries", len(historie))
	}
}

func TestGetDirektverkaufHistorie_GroupsByVerkaufMostRecentFirst(t *testing.T) {
	getaetigtA, verkaufA, _ := getaetigtEvent(t, 500, 2)    // 1000 cents
	getaetigtB, verkaufB, posB := getaetigtEvent(t, 400, 3) // 1200 cents
	subjectB := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufB)
	stornoPosB := kasse.Position{PositionID: posB, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 400, Menge: 1}
	stornoB, err := kasse.NewDirektverkaufStorniertEvent(subjectB, verkaufB, 2, "Leitung", []kasse.Position{stornoPosB}, 400, "Rückgabe")
	if err != nil {
		t.Fatalf("failed to create storno event: %v", err)
	}

	query := Query{
		EventRepo:           &mockHistorieRepo{events: []event.Event{getaetigtA, getaetigtB, stornoB}},
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	historie, err := query.GetDirektverkaufHistorie(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 2 {
		t.Fatalf("expected 2 historie entries, got %d", len(historie))
	}

	// Most recent Verkauf (B) first.
	if historie[0].VerkaufID != verkaufB {
		t.Errorf("expected first entry to be verkauf B (%s), got %s", verkaufB, historie[0].VerkaufID)
	}
	if historie[0].GesamtbetragCents != 1200 {
		t.Errorf("expected B gesamtbetragCents 1200, got %d", historie[0].GesamtbetragCents)
	}
	if historie[0].GesamtStorniertCents != 400 {
		t.Errorf("expected B gesamtStorniertCents 400, got %d", historie[0].GesamtStorniertCents)
	}
	if len(historie[0].OffenePositionen) != 1 || historie[0].OffenePositionen[0].Menge != 2 {
		t.Errorf("expected B offene Positionen with menge 2, got %+v", historie[0].OffenePositionen)
	}

	if historie[1].VerkaufID != verkaufA {
		t.Errorf("expected second entry to be verkauf A (%s), got %s", verkaufA, historie[1].VerkaufID)
	}
	if historie[1].GesamtStorniertCents != 0 {
		t.Errorf("expected A gesamtStorniertCents 0, got %d", historie[1].GesamtStorniertCents)
	}
}
