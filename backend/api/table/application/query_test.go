//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func TestGetAllTables(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{table.Table{ID: 1, Name: "Table 1", Status: table.ActiveStatus}}, nil)
	query := Query{TableRepo: repo}

	tables, err := query.GetAllTables(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}
	if tables[0].Name != "Table 1" {
		t.Errorf("expected name 'Table 1', got %s", tables[0].Name)
	}
}
