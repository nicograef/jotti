package kasse

import (
	"testing"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

// verkaufPositionen extracts the fat positions (incl. server-generated PositionIDs) from a getaetigt event.
func verkaufPositionen(t *testing.T, evt e.Event) []Position {
	t.Helper()
	data := DirektverkaufGetaetigtV1Data{}
	if err := e.ParseData(evt, &data, direktverkaufGetaetigtV1DataSchema); err != nil {
		t.Fatalf("failed to parse getaetigt event data: %v", err)
	}
	return PositionenFromEventData(data.Positionen)
}

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

func TestBuildDirektverkaufHistorieEintrag_PartialThenFullStorno(t *testing.T) {
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

	storno1, err := NewDirektverkaufStorniertEvent(subject, uuid.NewString(), verkaufID, 2, "Leitung", []Position{stornoPosition(beer, 1)}, 500, "Vertippt")
	if err != nil {
		t.Fatalf("failed to create storno1 event: %v", err)
	}

	eintrag, err := BuildDirektverkaufHistorieEintrag([]e.Event{getaetigt, storno1})
	if err != nil {
		t.Fatalf("failed to replay direktverkauf: %v", err)
	}
	rest := eintrag.OffenePositionen
	if got := mengeOf(rest, beerID); got != 1 {
		t.Errorf("expected 1 Beer remaining after partial storno, got %d", got)
	}
	if got := mengeOf(rest, wurstID); got != 1 {
		t.Errorf("expected 1 Bratwurst remaining, got %d", got)
	}

	storno2, err := NewDirektverkaufStorniertEvent(subject, uuid.NewString(), verkaufID, 3, "Leitung", []Position{stornoPosition(beer, 1), stornoPosition(wurst, 1)}, 850, "Rest")
	if err != nil {
		t.Fatalf("failed to create storno2 event: %v", err)
	}

	eintrag2, err := BuildDirektverkaufHistorieEintrag([]e.Event{getaetigt, storno1, storno2})
	if err != nil {
		t.Fatalf("failed to replay direktverkauf: %v", err)
	}
	rest2 := eintrag2.OffenePositionen
	if len(rest2) != 0 {
		t.Errorf("expected empty remaining after full storno, got %+v", rest2)
	}
}

func TestBuildDirektverkaufHistorieEintrag_Positionskommentar(t *testing.T) {
	verkaufID := uuid.NewString()
	subject := DirektverkaufSubject(1, verkaufID)
	wurst := testPosition(2, "Bratwurst", "mit Brot", "essen", 350, 2)
	wurst.Kommentar = "1 davon ohne Senf"
	v2, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", []Position{
		wurst,
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}, "")
	if err != nil {
		t.Fatalf("failed to create getaetigt event: %v", err)
	}
	eintrag, err := BuildDirektverkaufHistorieEintrag([]e.Event{v2})
	if err != nil {
		t.Fatalf("failed to replay v2 direktverkauf: %v", err)
	}
	if got := eintrag.Positionen[0].Kommentar; got != "1 davon ohne Senf" {
		t.Errorf("v2 Positionen[0].Kommentar = %q, want %q", got, "1 davon ohne Senf")
	}
	if got := eintrag.Positionen[1].Kommentar; got != "" {
		t.Errorf("v2 Positionen[1].Kommentar = %q, want empty", got)
	}

	v1ID := uuid.NewString()
	v1, err := e.New(1, "User", string(EventTypeDirektverkaufGetaetigtV1), DirektverkaufSubject(1, v1ID), DirektverkaufGetaetigtV1Data{
		VerkaufID: v1ID,
		Positionen: []PositionEventData{{
			PositionID: uuid.NewString(), VarianteID: 2, ProduktName: "Bratwurst", VarianteName: "mit Brot",
			Kategorie: "essen", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1,
		}},
		GesamtbetragCents: 350,
	})
	if err != nil {
		t.Fatalf("failed to create v1 event: %v", err)
	}
	eintragV1, err := BuildDirektverkaufHistorieEintrag([]e.Event{v1})
	if err != nil {
		t.Fatalf("failed to replay v1 direktverkauf: %v", err)
	}
	if got := eintragV1.Positionen[0].Kommentar; got != "" {
		t.Errorf("v1 Positionen[0].Kommentar = %q, want empty", got)
	}
	if got := eintragV1.VerkaufID; got != v1ID {
		t.Errorf("v1 VerkaufID = %q, want %q", got, v1ID)
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

	storno, err := NewDirektverkaufStorniertEvent(subject, uuid.NewString(), verkaufID, 2, "Leitung", []Position{stornoPosition(beer, 1)}, 500, "Rückgabe")
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
	if got := mengeOf(eintrag.Positionen, beerID); got != 2 {
		t.Errorf("expected original Positionen to keep 2 Beer, got %d", got)
	}
	if len(eintrag.Stornierungen) != 1 {
		t.Fatalf("expected 1 stornierung in eintrag, got %d", len(eintrag.Stornierungen))
	}
	if eintrag.Stornierungen[0].StornierungID == "" {
		t.Error("expected stornierungId to be set")
	}
	if eintrag.Stornierungen[0].GesamtStornierungCents != 500 {
		t.Errorf("expected stornierung betrag 500, got %d", eintrag.Stornierungen[0].GesamtStornierungCents)
	}
	if !eintrag.Stornierungen[0].StorniertAm.Equal(storno.Time) {
		t.Errorf("expected storniertAm %s, got %s", storno.Time, eintrag.Stornierungen[0].StorniertAm)
	}
}
