//go:build unit

package kasse

import "testing"

func TestComputeEigeneArbeitAnTisch_OffeneEigeneArbeit(t *testing.T) {
	session := TischSession{
		AusstehendePositionen: []Position{
			{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
		},
		UnbezahltePositionen: []Position{
			{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			{PositionID: "p3", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
		},
	}

	arbeit := ComputeEigeneArbeitAnTisch(session, 7)

	if arbeit.AnzahlAusstehend != 1 {
		t.Errorf("expected 1 ausstehend, got %d", arbeit.AnzahlAusstehend)
	}
	if arbeit.AnzahlUnbezahlt != 2 {
		t.Errorf("expected 2 unbezahlt, got %d", arbeit.AnzahlUnbezahlt)
	}
	// p1 (ausstehend ∪ unbezahlt) and p3 (unbezahlt) -> 2 distinct positions.
	if arbeit.AnzahlOffen != 2 {
		t.Errorf("expected 2 offen (union), got %d", arbeit.AnzahlOffen)
	}
	if arbeit.Erledigt {
		t.Errorf("expected not erledigt with open own positions")
	}
}

func TestComputeEigeneArbeitAnTisch_ErledigtWhenNoOwnPositions(t *testing.T) {
	session := TischSession{
		AusstehendePositionen: []Position{
			{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
		},
		UnbezahltePositionen: []Position{
			{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
		},
	}

	arbeit := ComputeEigeneArbeitAnTisch(session, 7)

	if !arbeit.Erledigt {
		t.Errorf("expected erledigt for servicekraft without own positions")
	}
	if arbeit.AnzahlOffen != 0 {
		t.Errorf("expected 0 offen, got %d", arbeit.AnzahlOffen)
	}
}

// Schichtübergabe: a colleague delivered and paid Anna's positions, so they no
// longer appear in either list — Anna is "erledigt" at this table.
func TestComputeEigeneArbeitAnTisch_SchichtuebergabeErledigt(t *testing.T) {
	session := TischSession{
		AusstehendePositionen: nil,
		UnbezahltePositionen:  nil,
	}

	arbeit := ComputeEigeneArbeitAnTisch(session, 7)

	if !arbeit.Erledigt {
		t.Errorf("expected erledigt after colleague handled all own positions")
	}
}

func TestComputeEigeneArbeitAnTisch_OnlyAusstehendNotErledigt(t *testing.T) {
	session := TischSession{
		AusstehendePositionen: []Position{
			{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
		},
		UnbezahltePositionen: nil,
	}

	arbeit := ComputeEigeneArbeitAnTisch(session, 7)

	if arbeit.Erledigt {
		t.Errorf("expected not erledigt with one ausstehende own position")
	}
	if arbeit.AnzahlAusstehend != 1 || arbeit.AnzahlUnbezahlt != 0 || arbeit.AnzahlOffen != 1 {
		t.Errorf("unexpected counts: %+v", arbeit)
	}
}
