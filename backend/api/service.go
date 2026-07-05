package api

import (
	"database/sql"
	"net/http"

	belegApp "github.com/nicograef/jotti/backend/api/druck/beleg/application"
	belegHTTP "github.com/nicograef/jotti/backend/api/druck/beleg/http"
	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	productApp "github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	productHTTP "github.com/nicograef/jotti/backend/api/stammdaten/produkt/http"
	tischApp "github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	tischHTTP "github.com/nicograef/jotti/backend/api/stammdaten/tisch/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/betreiber_repo"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/favorit_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

func NewServiceApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	productRepo := produkt_repo.NewRepository(db)
	pq := productHTTP.QueryHandler{}
	pq.Query = productApp.Query{ProductRepo: productRepo}
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tableRepo := tisch_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	favoritRepo := favorit_repo.NewRepository(db)
	druckstationRepo := druckstation_repo.NewRepository(db)
	druckauftragRepo := druckauftrag_repo.NewRepository(db)
	betreiberRepo := betreiber_repo.NewRepository(db)

	tseRepo := tse_repo.NewRepository(db)

	tc := tischgeschaeftHTTP.CommandHandler{}
	tc.Command = tischgeschaeftApp.Command{
		TableRepo:           tableRepo,
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		FavoritRepo:         favoritRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepo,
	}
	r.HandleFunc("/bestellung-aufnehmen", tc.BestellungAufnehmenHandler())
	r.HandleFunc("/bestellung-umbuchen", tc.BestellungUmbuchenHandler())
	r.HandleFunc("/zahlung-kassieren", tc.ZahlungKassierenHandler())
	r.HandleFunc("/ausgabe-bestaetigen", tc.AusgabeBestaetigenHandler())

	bc := belegHTTP.CommandHandler{}
	bc.Command = belegApp.Command{
		EventRepo:           kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepo,
		DruckauftragRepo:    druckauftragRepo,
		BetreiberRepo:       betreiberRepo,
		TSERepo:             tseRepo,
	}
	r.HandleFunc("/beleg-drucken", bc.KassenbelegDruckenHandler())

	stc := tischHTTP.CommandHandler{}
	stc.Command = tischApp.Command{
		TableRepo:   tableRepo,
		FavoritRepo: favoritRepo,
	}
	r.HandleFunc("/favorit-hinzufuegen", stc.FavoritHinzufuegenHandler())
	r.HandleFunc("/favorit-entfernen", stc.FavoritEntfernenHandler())

	dc := direktverkaufHTTP.CommandHandler{}
	dc.Command = direktverkaufApp.Command{
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepo,
	}
	r.HandleFunc("/direktverkauf-taetigen", dc.DirektverkaufTaetigenHandler())

	dq := direktverkaufHTTP.QueryHandler{}
	dq.Query = direktverkaufApp.Query{
		EventRepo:           kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
	}
	r.HandleFunc("/get-direktverkauf-historie", dq.GetDirektverkaufHistorieHandler())

	tq := tischgeschaeftHTTP.QueryHandler{}
	tq.Query = tischgeschaeftApp.Query{TableRepo: tableRepo, EventRepo: kassenjournalRepo, FavoritRepo: favoritRepo, KassensitzungenRepo: kassensitzungenRepo}
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
