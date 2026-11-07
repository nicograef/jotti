package kasse

import (
	"reflect"
	"testing"
)

func pos(positionID string, bestellerUserID int, bestellerName string) Position {
	return Position{PositionID: positionID, Menge: 1, BestellerUserID: bestellerUserID, BestellerName: bestellerName}
}

func TestHatEigeneOffenePositionen(t *testing.T) {
	tests := []struct {
		name    string
		session TischSession
		want    bool
	}{
		{
			name:    "eigene unbezahlte Position",
			session: TischSession{UnbezahltePositionen: []Position{pos("p2", 8, "Bert"), pos("p3", 7, "Anna")}},
			want:    true,
		},
		{
			name:    "nur fremde unbezahlte Positionen",
			session: TischSession{UnbezahltePositionen: []Position{pos("p2", 8, "Bert")}},
			want:    false,
		},
		{
			// Schichtübergabe: eine Kollegin hat die eigenen Positionen kassiert.
			name:    "keine unbezahlten Positionen",
			session: TischSession{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HatEigeneOffenePositionen(tt.session, 7); got != tt.want {
				t.Errorf("HatEigeneOffenePositionen = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeOffeneArbeitProServicekraft(t *testing.T) {
	sessions := []TischSession{
		{
			TischID: 3,
			UnbezahltePositionen: []Position{
				{PositionID: "p1", Menge: 2, EinzelpreisCents: 250, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{
			TischID: 1,
			UnbezahltePositionen: []Position{
				{PositionID: "p2", Menge: 3, EinzelpreisCents: 100, BestellerUserID: 8, BestellerName: "Bert"},
				{PositionID: "p3", Menge: 1, EinzelpreisCents: 400, BestellerUserID: 7, BestellerName: "Anna"},
				{PositionID: "p4", Menge: 1, EinzelpreisCents: 50, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{TischID: 2},
	}

	got := ComputeOffeneArbeitProServicekraft(sessions)

	// Aufsteigend nach UserID, je Servicekraft aufsteigend nach Tisch-ID; Tisch 2 ohne offene Arbeit fehlt.
	want := []OffeneArbeitServicekraft{
		{UserID: 7, UserName: "Anna", OffeneTische: []OffeneArbeitTisch{{TischID: 1, OffenCents: 450}, {TischID: 3, OffenCents: 500}}},
		{UserID: 8, UserName: "Bert", OffeneTische: []OffeneArbeitTisch{{TischID: 1, OffenCents: 300}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestComputeOffeneArbeitProServicekraft_EmptySessions(t *testing.T) {
	servicekraefte := ComputeOffeneArbeitProServicekraft(nil)

	if len(servicekraefte) != 0 {
		t.Errorf("expected no servicekraefte for no sessions, got %+v", servicekraefte)
	}
}
