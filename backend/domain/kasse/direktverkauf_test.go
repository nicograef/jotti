//go:build unit

package kasse

import (
	"testing"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

// verkaufPositionen extracts the fat positions (incl. server-generated PositionIDs) from a
// getaetigt event, so tests can build fat storno positions that reference real positions.
func verkaufPositionen(t *testing.T, evt e.Event) []Position {
	t.Helper()
	data := direktverkaufGetaetigtV1Data{}
	if err := e.ParseData(evt, &data, direktverkaufGetaetigtV1DataSchema); err != nil {
		t.Fatalf("failed to parse getaetigt event data: %v", err)
	}
	return fromPositionenEventData(data.Positionen)
}

// stornoPosition copies a fat position with an overridden Menge for use in a storno event.
func stornoPosition(p Position, menge int) Position {
	p.Menge = menge
	return p
}

func mengeOf(positionen []Position, positionID string) int {
	for _, p := range positionen {
		if p.PositionID == positionID {
			return p.Menge
		}
	}
	return 0
}

func TestComputeNichtStornierteVerkaufPositionen_PartialThenFullStorno(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)
	getaetigt, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
		testPosition(2, "Bratwurst", "mit Brot", "essen", 350, 1),
	}, "")
	if err != nil {
		t.Fatalf("failed to create getaetigt event: %v", err)
	}
	positionen := verkaufPositionen(t, getaetigt)
	beer, wurst := positionen[0], positionen[1]
	beerID, wurstID := beer.PositionID, wurst.PositionID

	storno1, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []Position{stornoPosition(beer, 1)}, 500, "Vertippt")
	if err != nil {
		t.Fatalf("failed to create storno1 event: %v", err)
	}

	rest, err := ComputeNichtStornierteVerkaufPositionen([]e.Event{getaetigt, storno1})
	if err != nil {
		t.Fatalf("failed to compute remaining positionen: %v", err)
	}
	if got := mengeOf(rest, beerID); got != 1 {
		t.Errorf("expected 1 Beer remaining after partial storno, got %d", got)
	}
	if got := mengeOf(rest, wurstID); got != 1 {
		t.Errorf("expected 1 Bratwurst remaining, got %d", got)
	}

	storno2, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 3, "Leitung", []Position{stornoPosition(beer, 1), stornoPosition(wurst, 1)}, 850, "Rest")
	if err != nil {
		t.Fatalf("failed to create storno2 event: %v", err)
	}

	rest2, err := ComputeNichtStornierteVerkaufPositionen([]e.Event{getaetigt, storno1, storno2})
	if err != nil {
		t.Fatalf("failed to compute remaining positionen: %v", err)
	}
	if len(rest2) != 0 {
		t.Errorf("expected empty remaining after full storno, got %+v", rest2)
	}
}

func TestBuildDirektverkaufHistorieEintrag_AggregatesStornos(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)
	getaetigt, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}, "Tresen")
	if err != nil {
		t.Fatalf("failed to create getaetigt event: %v", err)
	}
	beer := verkaufPositionen(t, getaetigt)[0]
	beerID := beer.PositionID

	storno, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []Position{stornoPosition(beer, 1)}, 500, "Rückgabe")
	if err != nil {
		t.Fatalf("failed to create storno event: %v", err)
	}

	eintrag, err := BuildDirektverkaufHistorieEintrag([]e.Event{getaetigt, storno})
	if err != nil {
		t.Fatalf("failed to build historie eintrag: %v", err)
	}

	if eintrag.VerkaufID != verkaufID {
		t.Errorf("expected verkaufId %s, got %s", verkaufID, eintrag.VerkaufID)
	}
	if eintrag.GesamtbetragCents != 1000 {
		t.Errorf("expected gesamtbetragCents 1000, got %d", eintrag.GesamtbetragCents)
	}
	if eintrag.GesamtStorniertCents != 500 {
		t.Errorf("expected gesamtStorniertCents 500, got %d", eintrag.GesamtStorniertCents)
	}
	if got := mengeOf(eintrag.OffenePositionen, beerID); got != 1 {
		t.Errorf("expected 1 Beer in offene Positionen, got %d", got)
	}
	// The original positions must stay untouched by the storno reduction.
	if got := mengeOf(eintrag.Positionen, beerID); got != 2 {
		t.Errorf("expected original Positionen to keep 2 Beer, got %d", got)
	}
}
