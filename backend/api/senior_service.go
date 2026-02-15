package api

import (
	"database/sql"
	"net/http"

	table "github.com/nicograef/jotti/backend/api/table/http"
)

func NewSeniorServiceApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/cancel-table-variants", tc.CancelTableVariantsHandler())

	return r
}
