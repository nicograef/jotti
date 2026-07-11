//go:build unit

package kasse

import "testing"

func pos(positionID string, bestellerUserID int, bestellerName string) Position {
	return Position{PositionID: positionID, Menge: 1, BestellerUserID: bestellerUserID, BestellerName: bestellerName}
}

func TestComputeEigeneArbeitAnTisch(t *testing.T) {
	tests := []struct {
		name          string
		session       TischSession
		userID        int
		wantUnbezahlt int
		wantOffen     int
		wantErledigt  bool
	}{
		{
			name: "offene eigene unbezahlte Positionen",
			session: TischSession{
				UnbezahltePositionen: []Position{
					pos("p1", 7, "Anna"),
					pos("p2", 8, "Bert"),
					pos("p3", 7, "Anna"),
				},
			},
			userID:        7,
			wantUnbezahlt: 2,
			wantOffen:     2,
			wantErledigt:  false,
		},
		{
			name: "erledigt ohne eigene unbezahlte Positionen",
			session: TischSession{
				UnbezahltePositionen: []Position{pos("p2", 8, "Bert")},
			},
			userID:       7,
			wantErledigt: true,
		},
		{
			// Schichtübergabe: eine Kollegin hat die eigenen Positionen kassiert,
			// sie sind aus der Unbezahlt-Liste verschwunden.
			name:         "schichtübergabe erledigt",
			session:      TischSession{UnbezahltePositionen: nil},
			userID:       7,
			wantErledigt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arbeit := ComputeEigeneArbeitAnTisch(tt.session, tt.userID)
			if arbeit.AnzahlUnbezahlt != tt.wantUnbezahlt {
				t.Errorf("AnzahlUnbezahlt = %d, want %d", arbeit.AnzahlUnbezahlt, tt.wantUnbezahlt)
			}
			if arbeit.AnzahlOffen != tt.wantOffen {
				t.Errorf("AnzahlOffen = %d, want %d", arbeit.AnzahlOffen, tt.wantOffen)
			}
			if arbeit.Erledigt != tt.wantErledigt {
				t.Errorf("Erledigt = %v, want %v", arbeit.Erledigt, tt.wantErledigt)
			}
		})
	}
}

func TestComputeOffeneArbeit_OffenCents(t *testing.T) {
	sessions := []TischSession{
		{
			TischID: 3,
			UnbezahltePositionen: []Position{
				{PositionID: "p1", Menge: 2, EinzelpreisCents: 250, BestellerUserID: 7},
			},
		},
		{
			TischID: 1,
			UnbezahltePositionen: []Position{
				{PositionID: "p2", Menge: 1, EinzelpreisCents: 400, BestellerUserID: 7},
				{PositionID: "p3", Menge: 3, EinzelpreisCents: 100, BestellerUserID: 8},
			},
		},
	}

	// Einzeltisch: nur eigene Positionen zählen (2 × 250 = 500 Cent).
	arbeit := ComputeEigeneArbeitAnTisch(sessions[0], 7)
	if arbeit.OffenCents != 500 {
		t.Errorf("OffenCents = %d, want 500", arbeit.OffenCents)
	}

	// Rollup reicht OffenCents je Tisch durch (Tisch 1: 400, Tisch 3: 500).
	rollup := ComputeOffeneArbeitRollup(sessions, 7)
	if len(rollup.OffeneTische) != 2 {
		t.Fatalf("expected 2 offene Tische, got %d: %+v", len(rollup.OffeneTische), rollup.OffeneTische)
	}
	if rollup.OffeneTische[0].TischID != 1 || rollup.OffeneTische[0].OffenCents != 400 {
		t.Errorf("Tisch 1: got %+v, want OffenCents 400", rollup.OffeneTische[0])
	}
	if rollup.OffeneTische[1].TischID != 3 || rollup.OffeneTische[1].OffenCents != 500 {
		t.Errorf("Tisch 3: got %+v, want OffenCents 500", rollup.OffeneTische[1])
	}
}

func TestComputeOffeneArbeitRollup(t *testing.T) {
	sessions := []TischSession{
		{
			TischID:              3,
			UnbezahltePositionen: []Position{pos("p1", 7, "Anna")},
		},
		{
			TischID: 1,
			UnbezahltePositionen: []Position{
				pos("p2", 8, "Bert"),
				pos("p3", 7, "Anna"),
			},
		},
	}

	tests := []struct {
		name         string
		userID       int
		wantTischIDs []int
		wantErledigt bool
	}{
		{
			name:         "offene Arbeit an mehreren Tischen, aufsteigend sortiert",
			userID:       7,
			wantTischIDs: []int{1, 3},
			wantErledigt: false,
		},
		{
			name:         "servicekraft nur an einem Tisch offen",
			userID:       8,
			wantTischIDs: []int{1},
			wantErledigt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollup := ComputeOffeneArbeitRollup(sessions, tt.userID)
			if rollup.Erledigt != tt.wantErledigt {
				t.Errorf("Erledigt = %v, want %v", rollup.Erledigt, tt.wantErledigt)
			}
			if len(rollup.OffeneTische) != len(tt.wantTischIDs) {
				t.Fatalf("got %d offene Tische, want %d: %+v", len(rollup.OffeneTische), len(tt.wantTischIDs), rollup.OffeneTische)
			}
			for i, wantID := range tt.wantTischIDs {
				if rollup.OffeneTische[i].TischID != wantID {
					t.Errorf("OffeneTische[%d].TischID = %d, want %d", i, rollup.OffeneTische[i].TischID, wantID)
				}
				if rollup.OffeneTische[i].AnzahlUnbezahlt != 1 || rollup.OffeneTische[i].AnzahlOffen != 1 {
					t.Errorf("OffeneTische[%d] unexpected counts: %+v", i, rollup.OffeneTische[i])
				}
			}
		})
	}
}

func TestComputeOffeneArbeitRollup_AllesErledigt(t *testing.T) {
	// Nur fremde Positionen -> die Servicekraft ist überall fertig.
	sessions := []TischSession{
		{TischID: 1, UnbezahltePositionen: []Position{pos("p1", 8, "Bert")}},
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

func TestComputeOffeneArbeitProServicekraft(t *testing.T) {
	sessions := []TischSession{
		{
			TischID:              3,
			UnbezahltePositionen: []Position{pos("p1", 7, "Anna")},
		},
		{
			TischID: 1,
			UnbezahltePositionen: []Position{
				pos("p2", 8, "Bert"),
				pos("p3", 7, "Anna"),
			},
		},
	}

	servicekraefte := ComputeOffeneArbeitProServicekraft(sessions)

	if len(servicekraefte) != 2 {
		t.Fatalf("expected 2 servicekraefte with open work, got %d: %+v", len(servicekraefte), servicekraefte)
	}
	// Aufsteigend nach UserID: Anna (7) zuerst, dann Bert (8).
	anna := servicekraefte[0]
	if anna.UserID != 7 || anna.UserName != "Anna" {
		t.Errorf("expected first entry Anna (7), got %+v", anna)
	}
	if len(anna.OffeneTische) != 2 || anna.OffeneTische[0].TischID != 1 || anna.OffeneTische[1].TischID != 3 {
		t.Errorf("expected Anna open at tische [1 3], got %+v", anna.OffeneTische)
	}
	bert := servicekraefte[1]
	if bert.UserID != 8 || bert.UserName != "Bert" {
		t.Errorf("expected second entry Bert (8), got %+v", bert)
	}
	if len(bert.OffeneTische) != 1 || bert.OffeneTische[0].TischID != 1 {
		t.Errorf("expected Bert open at tisch [1], got %+v", bert.OffeneTische)
	}
}

func TestComputeOffeneArbeitProServicekraft_EmptySessions(t *testing.T) {
	servicekraefte := ComputeOffeneArbeitProServicekraft(nil)

	if len(servicekraefte) != 0 {
		t.Errorf("expected no servicekraefte for no sessions, got %+v", servicekraefte)
	}
}
