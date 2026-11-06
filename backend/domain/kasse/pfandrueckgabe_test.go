package kasse

import (
	"testing"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestBuildPfandrueckgabeHistorieEintrag_AggregatesStornos(t *testing.T) {
	rueckgabeID := uuid.New().String()
	subject := PfandrueckgabeSubject(1, rueckgabeID)
	rueckgabe, err := NewPfandZurueckgegebenEvent(subject, rueckgabeID, 7, 1, "User", []Position{
		testPosition(9, "Pfand", "Glas", "getraenk", 200, 3),
	}, "drei Gläser")
	if err != nil {
		t.Fatalf("rueckgabe event: %v", err)
	}
	data := PfandZurueckgegebenV1Data{}
	if err := e.ParseData(rueckgabe, &data, pfandZurueckgegebenV1DataSchema); err != nil {
		t.Fatalf("rueckgabe data: %v", err)
	}
	glas := fromPositionenEventData(data.Positionen)[0]

	storno1, err := NewPfandrueckgabeStorniertEvent(subject, uuid.New().String(), rueckgabeID, 7, 2, "Leitung", []Position{stornoPosition(glas, 1)}, 200, "Vertippt")
	if err != nil {
		t.Fatalf("storno1 event: %v", err)
	}
	storno2, err := NewPfandrueckgabeStorniertEvent(subject, uuid.New().String(), rueckgabeID, 7, 2, "Leitung", []Position{stornoPosition(glas, 1)}, 200, "Nochmal")
	if err != nil {
		t.Fatalf("storno2 event: %v", err)
	}

	nachErster, err := BuildPfandrueckgabeHistorieEintrag([]e.Event{rueckgabe, storno1})
	if err != nil {
		t.Fatalf("historie nach erster Stornierung: %v", err)
	}
	if got := mengeOf(nachErster.OffenePositionen, glas.PositionID); got != 2 {
		t.Errorf("offen nach erster Stornierung = %d, erwartet 2", got)
	}

	eintrag, err := BuildPfandrueckgabeHistorieEintrag([]e.Event{rueckgabe, storno1, storno2})
	if err != nil {
		t.Fatalf("historie nach zweiter Stornierung: %v", err)
	}
	if got := mengeOf(eintrag.OffenePositionen, glas.PositionID); got != 1 {
		t.Errorf("offen nach zweiter Stornierung = %d, erwartet 1", got)
	}
	if got := mengeOf(eintrag.Positionen, glas.PositionID); got != 3 {
		t.Errorf("Positionen = %d, erwartet unverändert 3", got)
	}
	if eintrag.RueckgabeID != rueckgabeID {
		t.Errorf("rueckgabeId = %q, erwartet %q", eintrag.RueckgabeID, rueckgabeID)
	}
	if eintrag.TischID != 7 {
		t.Errorf("tischId = %d, erwartet 7", eintrag.TischID)
	}
	if eintrag.GesamtbetragCents != 600 {
		t.Errorf("gesamtbetragCents = %d, erwartet 600", eintrag.GesamtbetragCents)
	}
	if eintrag.GesamtStorniertCents != 400 {
		t.Errorf("gesamtStorniertCents = %d, erwartet 400", eintrag.GesamtStorniertCents)
	}
	if len(eintrag.Stornierungen) != 2 {
		t.Fatalf("stornierungen = %d, erwartet 2", len(eintrag.Stornierungen))
	}
	if eintrag.Kommentar != "drei Gläser" || eintrag.UserName != "User" {
		t.Errorf("kommentar/user = %q/%q", eintrag.Kommentar, eintrag.UserName)
	}
}
