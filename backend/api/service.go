package api

import (
	"database/sql"
	"net/http"

	product "github.com/nicograef/jotti/backend/api/product/http"
	table "github.com/nicograef/jotti/backend/api/table/http"
)

func NewServiceApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	pq := product.NewQueryHandler(db)
	r.HandleFunc("/get-active-products", pq.GetActiveProductsHandler())

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/place-table-order", tc.PlaceTableOrderHandler())

	tq := table.NewQueryHandler(db)
	r.HandleFunc("/get-table", tq.GetTableHandler())
	r.HandleFunc("/get-active-tables", tq.GetActiveTablesHandler())
	r.HandleFunc("/get-table-orders", tq.GetTableOrdersHandler())
	r.HandleFunc("/get-table-balance", tq.GetTableBalanceHandler())
	r.HandleFunc("/get-table-unpaid-products", tq.GetTableUnpaidProductsHandler())

	return r
}
