package api

import (
	"database/sql"
	"net/http"

	table "github.com/nicograef/jotti/backend/api/table/http"
)

func NewServiceleitungApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/produkte-stornieren", tc.ProdukteStornierenHandler())

	return r
}
