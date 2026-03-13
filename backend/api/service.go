package api

import (
	"database/sql"
	"net/http"

	productHTTP "github.com/nicograef/jotti/backend/api/product/http"
	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	"github.com/nicograef/jotti/backend/repository/event_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func NewServiceApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	productRepo := product_repo.NewRepository(db)
	pq := productHTTP.QueryHandler{}
	pq.ProductRepo = productRepo
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tableRepo := table_repo.NewRepository(db)
	eventRepo := event_repo.NewRepository(db)
	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:   tableRepo,
		EventRepo:   eventRepo,
		ProductRepo: productRepo,
	}
	r.HandleFunc("/bestellung-aufgeben", tc.BestellungAufgebenHandler())
	r.HandleFunc("/zahlung-registrieren", tc.ZahlungRegistrierenHandler())
	r.HandleFunc("/produkte-liefern", tc.ProdukteLiefernHandler())

	tq := tableHTTP.QueryHandler{}
	tq.Query = tableApp.Query{TableRepo: tableRepo, EventRepo: eventRepo}
	r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
	r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
	r.HandleFunc("/get-tisch-state", tq.GetTischStateHandler())

	return r
}
