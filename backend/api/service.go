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
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/bestellung-aufgeben", tc.BestellungAufgebenHandler())
	r.HandleFunc("/zahlung-registrieren", tc.ZahlungRegistrierenHandler())
	r.HandleFunc("/produkte-liefern", tc.ProdukteLiefernHandler())

	tq := table.NewQueryHandler(db)
	r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
	r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
	r.HandleFunc("/get-tisch-state", tq.GetTischStateHandler())

	return r
}
