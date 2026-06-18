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

func TestComputeOffeneArbeitRollup_MehrereTischeUndServicekraefte(t *testing.T) {
	sessions := []TischSession{
		{
			TischID: 3,
			AusstehendePositionen: []Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			},
			UnbezahltePositionen: []Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{
			TischID: 1,
			AusstehendePositionen: []Position{
				{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
			},
			UnbezahltePositionen: []Position{
				{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
				{PositionID: "p3", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{
			// Anna hat hier alles erledigt; dieser Tisch darf nicht erscheinen.
			TischID: 2,
			AusstehendePositionen: []Position{
				{PositionID: "p4", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
			},
		},
	}

	rollup := ComputeOffeneArbeitRollup(sessions, 7)

	if rollup.Erledigt {
		t.Errorf("expected Anna to have open work")
	}
	if len(rollup.OffeneTische) != 2 {
		t.Fatalf("expected 2 open tische for Anna, got %d: %+v", len(rollup.OffeneTische), rollup.OffeneTische)
	}
	// Aufsteigend nach TischID sortiert: Tisch 1 zuerst, dann Tisch 3.
	if rollup.OffeneTische[0].TischID != 1 || rollup.OffeneTische[1].TischID != 3 {
		t.Errorf("expected tische sorted by id [1 3], got %+v", rollup.OffeneTische)
	}
	if rollup.OffeneTische[0].AnzahlAusstehend != 0 || rollup.OffeneTische[0].AnzahlUnbezahlt != 1 || rollup.OffeneTische[0].AnzahlOffen != 1 {
		t.Errorf("unexpected counts for tisch 1: %+v", rollup.OffeneTische[0])
	}
	if rollup.OffeneTische[1].AnzahlAusstehend != 1 || rollup.OffeneTische[1].AnzahlUnbezahlt != 1 || rollup.OffeneTische[1].AnzahlOffen != 1 {
		t.Errorf("unexpected counts for tisch 3: %+v", rollup.OffeneTische[1])
	}
}

// Schichtübergabe im Rollup: hat eine Kollegin alle eigenen Positionen
// ausgegeben und kassiert, gilt die Servicekraft über alle Tische als erledigt.
func TestComputeOffeneArbeitRollup_AllesErledigt(t *testing.T) {
	sessions := []TischSession{
		{
			TischID: 1,
			AusstehendePositionen: []Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
			},
			UnbezahltePositionen: []Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
			},
		},
		{TischID: 2},
	}

	rollup := ComputeOffeneArbeitRollup(sessions, 7)

	if !rollup.Erledigt {
		t.Errorf("expected erledigt for servicekraft without any open work")
	}
	if len(rollup.OffeneTische) != 0 {
		t.Errorf("expected no open tische, got %+v", rollup.OffeneTische)
	}
}

func TestComputeOffeneArbeitRollup_EmptySessions(t *testing.T) {
	rollup := ComputeOffeneArbeitRollup(nil, 7)

	if !rollup.Erledigt {
		t.Errorf("expected erledigt for no sessions")
	}
	if len(rollup.OffeneTische) != 0 {
		t.Errorf("expected empty open tische, got %+v", rollup.OffeneTische)
	}
}
