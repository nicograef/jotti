package api

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/event"
	"github.com/nicograef/jotti/backend/service/product"
	"github.com/nicograef/jotti/backend/service/table"
)

func NewApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	psq := product.QueryHandler{Query: &product.Query{Persistence: &product.Persistence{DB: db}}}
	r.HandleFunc("/get-all-products", psq.GetAllProductsHandler())

	tc := table.CommandHandler{Command: &table.Command{Persistence: &event.Persistence{DB: db}}}
	r.HandleFunc("/place-order", tc.PlaceOrderHandler())

	tq := table.QueryHandler{Query: &table.Query{EventPersistence: &event.Persistence{DB: db}, TablePersistence: &table.Persistence{DB: db}}}
	r.HandleFunc("/get-table", tq.GetTableHandler())
	r.HandleFunc("/get-all-tables", tq.GetAllTablesHandler())
	r.HandleFunc("/get-orders", tq.GetOrdersHandler())
	r.HandleFunc("/get-table-balance", tq.GetTableBalanceHandler())
	r.HandleFunc("/get-table-unpaid-products", tq.GetTableUnpaidProductsHandler())

	return r
}
