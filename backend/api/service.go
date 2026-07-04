package api

import (
	"database/sql"
	"net/http"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/direktverkauf/http"
	productApp "github.com/nicograef/jotti/backend/api/product/application"
	productHTTP "github.com/nicograef/jotti/backend/api/product/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/favorit_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
	"github.com/nicograef/jotti/backend/repository/settings_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

func NewServiceApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	productRepo := product_repo.NewRepository(db)
	pq := productHTTP.QueryHandler{}
	pq.Query = productApp.Query{ProductRepo: productRepo}
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tableRepo := table_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	favoritRepo := favorit_repo.NewRepository(db)
	druckstationRepo := druckstation_repo.NewRepository(db)
	druckauftragRepo := druckauftrag_repo.NewRepository(db)
	settingsRepo := settings_repo.NewRepository(db)

	tseRepo := tse_repo.NewRepository(db)

	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:           tableRepo,
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		FavoritRepo:         favoritRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepoTableAdapter{repo: druckstationRepo},
		DruckauftragRepo:    druckauftragRepo,
		SettingsRepo:        settingsRepo,
		TSERepo:             tseRepo,
	}
	r.HandleFunc("/bestellung-aufnehmen", tc.BestellungAufnehmenHandler())
	r.HandleFunc("/bestellung-umbuchen", tc.BestellungUmbuchenHandler())
	r.HandleFunc("/zahlung-kassieren", tc.ZahlungKassierenHandler())
	r.HandleFunc("/beleg-drucken", tc.KassenbelegDruckenHandler())
	r.HandleFunc("/ausgabe-bestaetigen", tc.AusgabeBestaetigenHandler())
	r.HandleFunc("/favorit-hinzufuegen", tc.FavoritHinzufuegenHandler())
	r.HandleFunc("/favorit-entfernen", tc.FavoritEntfernenHandler())

	dc := direktverkaufHTTP.CommandHandler{}
	dc.Command = direktverkaufApp.Command{
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepoTableAdapter{repo: druckstationRepo},
	}
	r.HandleFunc("/direktverkauf-taetigen", dc.DirektverkaufTaetigenHandler())

	dq := direktverkaufHTTP.QueryHandler{}
	dq.Query = direktverkaufApp.Query{
		EventRepo:           kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
	}
	r.HandleFunc("/get-direktverkauf-historie", dq.GetDirektverkaufHistorieHandler())

	tq := tableHTTP.QueryHandler{}
	tq.Query = tableApp.Query{TableRepo: tableRepo, EventRepo: kassenjournalRepo, FavoritRepo: favoritRepo, KassensitzungenRepo: kassensitzungenRepo}
	r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
	r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
	r.HandleFunc("/get-tisch-state", tq.GetTischStateHandler())
	r.HandleFunc("/get-aktive-tische-mit-favoriten", tq.GetAktiveTischeMitFavoritenHandler())
	r.HandleFunc("/get-meine-tische-state", tq.GetMeineTischeStateHandler())

	reportingRepo := reporting_repo.NewRepository(db)
	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{
		ReportingRepo:       reportingRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		TischSessionRepo:    kassenjournalRepo,
		TischRepo:           tableRepo,
	}
	r.HandleFunc("/get-eigene-uebersicht", rq.GetEigeneUebersichtHandler())

	return r
}
