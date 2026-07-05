//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

func TestGetAllTische(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus}}, nil)
	query := Query{TischRepo: repo}

	tische, err := query.GetAllTische(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tische) != 1 {
		t.Fatalf("expected 1 tisch, got %d", len(tische))
	}
	if tische[0].Name != "Tisch 1" {
		t.Errorf("expected name 'Tisch 1', got %s", tische[0].Name)
	}
}
