//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/admin/table/domain"
)

type mockQueryRepo struct {
	err   error
	table domain.Table
}

func (m *mockQueryRepo) GetAllTables(ctx context.Context) ([]domain.Table, error) {
	return []domain.Table{m.table}, m.err
}

func TestGetAllTables(t *testing.T) {
	query := Query{TableRepo: &mockQueryRepo{
		table: domain.Table{ID: 1, Name: "Table 1", Status: domain.ActiveStatus},
	}}

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
