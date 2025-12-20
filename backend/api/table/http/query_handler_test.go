//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/table"
)

type mockQuery struct {
	table   table.Table
	order   table.Order
	product table.OrderProduct
	balance int
	err     error
}

func (m mockQuery) GetTable(ctx context.Context, id int) (table.Table, error) {
	return m.table, m.err
}
func (m mockQuery) GetAllTables(ctx context.Context) ([]table.Table, error) {
	return []table.Table{m.table}, m.err
}
func (m mockQuery) GetActiveTables(ctx context.Context) ([]table.Table, error) {
	return []table.Table{m.table}, m.err
}
func (m mockQuery) GetTableOrders(ctx context.Context, tableID int) ([]table.Order, error) {
	return []table.Order{m.order}, m.err
}
func (m mockQuery) GetTablePayments(ctx context.Context, tableID int) ([]table.Payment, error) {
	return []table.Payment{}, m.err
}
func (m mockQuery) GetTableBalance(ctx context.Context, tableID int) (int, error) {
	return m.balance, m.err
}
func (m mockQuery) GetTableUnpaidProducts(ctx context.Context, tableID int) ([]table.OrderProduct, error) {
	return []table.OrderProduct{m.product}, m.err
}

func TestGetAllTablesHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetAllTablesHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
