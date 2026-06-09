//go:build unit

package kasse

import (
	"testing"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestNewDirektverkaufGetaetigtEvent_SingleEventWithConsistentSum(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)
	positionen := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
		testPosition(2, "Bratwurst", "mit Brot", "essen", 350, 1),
	}

	event, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "TestUser", positionen, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.Type != string(EventTypeDirektverkaufGetaetigtV1) {
		t.Errorf("expected type %s, got %s", EventTypeDirektverkaufGetaetigtV1, event.Type)
	}
	if event.Subject != subject {
		t.Errorf("expected subject %s, got %s", subject, event.Subject)
	}

	data := direktverkaufGetaetigtV1Data{}
	if err := e.ParseData(event, &data, direktverkaufGetaetigtV1DataSchema); err != nil {
		t.Fatalf("failed to parse event data: %v", err)
	}

	if data.VerkaufID != verkaufID {
		t.Errorf("expected verkaufId %s, got %s", verkaufID, data.VerkaufID)
	}

	wantSum := 500*2 + 350*1
	if data.GesamtbetragCents != wantSum {
		t.Errorf("expected gesamtbetragCents %d, got %d", wantSum, data.GesamtbetragCents)
	}

	if len(data.Positionen) != 2 {
		t.Fatalf("expected 2 positionen, got %d", len(data.Positionen))
	}
	for _, p := range data.Positionen {
		if _, err := uuid.Parse(p.PositionID); err != nil {
			t.Errorf("expected server-generated UUID position ID, got %q", p.PositionID)
		}
	}
}

func TestNewDirektverkaufGetaetigtEvent_RejectsEmptyPositionen(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)

	_, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "TestUser", []Position{}, "")
	if err == nil {
		t.Fatal("expected error for empty positionen, got nil")
	}
}

func TestNewDirektverkaufStorniertEvent_ValidatesAndStoresPositionen(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)
	posID := uuid.New().String()
	positionen := []Position{{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 500, Menge: 2}}

	event, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", positionen, 1000, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.Type != string(EventTypeDirektverkaufStorniertV1) {
		t.Errorf("expected type %s, got %s", EventTypeDirektverkaufStorniertV1, event.Type)
	}
	if event.Subject != subject {
		t.Errorf("expected subject %s, got %s", subject, event.Subject)
	}

	data := direktverkaufStorniertV1Data{}
	if err := e.ParseData(event, &data, direktverkaufStorniertV1DataSchema); err != nil {
		t.Fatalf("failed to parse event data: %v", err)
	}

	if data.VerkaufID != verkaufID {
		t.Errorf("expected verkaufId %s, got %s", verkaufID, data.VerkaufID)
	}
	if _, err := uuid.Parse(data.StornierungID); err != nil {
		t.Errorf("expected server-generated UUID stornierungId, got %q", data.StornierungID)
	}
	if len(data.Positionen) != 1 || data.Positionen[0].PositionID != posID || data.Positionen[0].Menge != 2 {
		t.Errorf("expected one Position {%s, 2}, got %+v", posID, data.Positionen)
	}
	if data.Positionen[0].ProduktName != "Cola" {
		t.Errorf("expected fat position to preserve ProduktName 'Cola', got %q", data.Positionen[0].ProduktName)
	}
	if data.GesamtStornierungCents != 1000 {
		t.Errorf("expected gesamtStornierungCents 1000, got %d", data.GesamtStornierungCents)
	}
}

func TestNewDirektverkaufStorniertEvent_RejectsShortKommentar(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)

	_, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []Position{{PositionID: uuid.New().String(), VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 500, Menge: 1}}, 500, "ab")
	if err == nil {
		t.Fatal("expected error for kommentar shorter than 3 characters, got nil")
	}
}

func TestNewDirektverkaufStorniertEvent_RejectsEmptyPositionen(t *testing.T) {
	verkaufID := uuid.New().String()
	subject := DirektverkaufSubject(1, verkaufID)

	_, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []Position{}, 0, "Rückgabe")
	if err == nil {
		t.Fatal("expected error for empty positionen, got nil")
	}
}
