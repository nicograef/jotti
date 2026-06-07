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
