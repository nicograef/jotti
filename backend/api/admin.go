package api

import (
	"database/sql"
	"net/http"

	druckauftragApp "github.com/nicograef/jotti/backend/api/druck/auftrag/application"
	druckauftragHTTP "github.com/nicograef/jotti/backend/api/druck/auftrag/http"
	druckstationApp "github.com/nicograef/jotti/backend/api/druck/station/application"
	druckstationHTTP "github.com/nicograef/jotti/backend/api/druck/station/http"
	exportApp "github.com/nicograef/jotti/backend/api/fiskal/export/application"
	exportHTTP "github.com/nicograef/jotti/backend/api/fiskal/export/http"
	fiskalSetupApp "github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	fiskalSetupHTTP "github.com/nicograef/jotti/backend/api/fiskal/setup/http"
	tseApp "github.com/nicograef/jotti/backend/api/fiskal/signatur/application"
	tseHTTP "github.com/nicograef/jotti/backend/api/fiskal/signatur/http"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	kasseHTTP "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	betreiberApp "github.com/nicograef/jotti/backend/api/stammdaten/betreiber/application"
	betreiberHTTP "github.com/nicograef/jotti/backend/api/stammdaten/betreiber/http"
	productApp "github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	productHTTP "github.com/nicograef/jotti/backend/api/stammdaten/produkt/http"
	tischApp "github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	tischHTTP "github.com/nicograef/jotti/backend/api/stammdaten/tisch/http"
	userApp "github.com/nicograef/jotti/backend/api/stammdaten/user/application"
	userHTTP "github.com/nicograef/jotti/backend/api/stammdaten/user/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/domain/tse"
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
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func NewAdminApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	userRepo := user_repo.NewRepository(db)
	uc := userHTTP.CommandHandler{}
	uc.Command = userApp.Command{UserRepo: userRepo}
	r.HandleFunc("/create-user", uc.CreateUserHandler())
	r.HandleFunc("/update-user", uc.UpdateUserHandler())
	r.HandleFunc("/activate-user", uc.ActivateUserHandler())
	r.HandleFunc("/deactivate-user", uc.DeactivateUserHandler())
	r.HandleFunc("/delete-user", uc.DeleteUserHandler())
	r.HandleFunc("/reset-password", uc.ResetPasswordHandler())

	uq := userHTTP.QueryHandler{}
	uq.Query = userApp.Query{UserRepo: userRepo}
	r.HandleFunc("/get-all-users", uq.GetAllUsersHandler())

	productRepo := produkt_repo.NewRepository(db)
	pc := productHTTP.CommandHandler{}
	pc.Command = productApp.Command{ProductRepo: productRepo}
	r.HandleFunc("/create-produkt", pc.CreateProductHandler())
	r.HandleFunc("/update-produkt", pc.UpdateProductHandler())
	r.HandleFunc("/create-variante", pc.CreateVariantHandler())
	r.HandleFunc("/update-variante", pc.UpdateVariantHandler())
	r.HandleFunc("/activate-variante", pc.ActivateVariantHandler())
	r.HandleFunc("/deactivate-variante", pc.DeactivateVariantHandler())
	r.HandleFunc("/delete-produkt", pc.DeleteProduktHandler())
	r.HandleFunc("/delete-variante", pc.DeleteVarianteHandler())

	pq := productHTTP.QueryHandler{}
	pq.Query = productApp.Query{ProductRepo: productRepo}
	r.HandleFunc("/get-all-produkte", pq.GetAllProductsHandler())

	tableRepo := tisch_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	betreiberRepo := betreiber_repo.NewRepository(db)
	tseStore := tse_repo.NewRepository(db)
	favoritRepo := favorit_repo.NewRepository(db)
	tc := tischHTTP.CommandHandler{}
	tc.Command = tischApp.Command{
		TableRepo:   tableRepo,
		FavoritRepo: favoritRepo,
	}
	r.HandleFunc("/update-tisch", tc.TischAktualisierenHandler())
	r.HandleFunc("/create-tisch", tc.TischErstellenHandler())
	r.HandleFunc("/activate-tisch", tc.TischAktivierenHandler())
	r.HandleFunc("/deactivate-tisch", tc.TischDeaktivierenHandler())
	r.HandleFunc("/delete-tisch", tc.TischLoeschenHandler())

	tq := tischHTTP.QueryHandler{}
	tq.Query = tischApp.Query{TableRepo: tableRepo}
	r.HandleFunc("/get-all-tische", tq.GetAllTischeHandler())

	reportingRepo := reporting_repo.NewRepository(db)
	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{
		ReportingRepo:       reportingRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		TischSessionRepo:    kassenjournalRepo,
		TischRepo:           tableRepo,
	}
	r.HandleFunc("/get-abrechnung", rq.GetReportingHandler())
	r.HandleFunc("/get-all-kassensitzungen", rq.GetAllKassensitzungenHandler())
	r.HandleFunc("/get-live-reporting", rq.GetLiveReportingHandler())

	exportHandler := exportHTTP.Handler{}
	exportHandler.Service = exportApp.Export{
		KassenjournalRepo:   kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		BetreiberRepo:       betreiberRepo,
		TSERepo:             tseStore,
		TableRepo:           tableRepo,
	}
	r.HandleFunc("/export/dsfinvk", exportHandler.ExportHandler())

	kc := kasseHTTP.CommandHandler{}
	kc.Command = kasseApp.Command{
		KassenjournalRepo:   kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		BetreiberRepo:       betreiberRepo,
		TSERepo:             tseStore,
	}
	r.HandleFunc("/kassensitzung-eroeffnen", kc.KassensitzungEroeffnenHandler())
	r.HandleFunc("/geldtransit-buchen", kc.GeldtransitBuchenHandler())
	r.HandleFunc("/kasse-abschliessen", kc.KasseAbschliessenHandler())

	kq := kasseHTTP.QueryHandler{}
	kq.Query = kasseApp.Query{KassenjournalRepo: kassenjournalRepo, KassensitzungenRepo: kassensitzungenRepo}
	r.HandleFunc("/get-offene-kassensitzung", kq.GetOffeneKassensitzungHandler())
	r.HandleFunc("/get-kassenbestand", kq.GetKassenbestandHandler())

	druckstationRepo := druckstation_repo.NewRepository(db)
	druckstationCommandHandler := druckstationHTTP.CommandHandler{}
	druckstationCommandHandler.Command = druckstationApp.Command{DruckstationRepo: druckstationRepo}
	druckstationQueryHandler := druckstationHTTP.QueryHandler{}
	druckstationQueryHandler.Query = druckstationApp.Query{DruckstationRepo: druckstationRepo}
	r.HandleFunc("/get-druckstationen", druckstationQueryHandler.GetDruckstationenHandler())
	r.HandleFunc("/update-druckstationen", druckstationCommandHandler.UpdateDruckstationenHandler())

	druckauftragRepo := druckauftrag_repo.NewRepository(db)
	druckauftragCommandHandler := druckauftragHTTP.CommandHandler{}
	druckauftragCommandHandler.Command = druckauftragApp.Command{DruckauftragRepo: druckauftragRepo}
	druckauftragQueryHandler := druckauftragHTTP.QueryHandler{}
	druckauftragQueryHandler.Query = druckauftragApp.Query{DruckauftragRepo: druckauftragRepo}
	r.HandleFunc("/get-fehlgeschlagene-druckauftraege", druckauftragQueryHandler.GetFehlgeschlageneDruckauftraegeHandler())
	r.HandleFunc("/druckauftrag-erneut-versuchen", druckauftragCommandHandler.DruckauftragErneutVersuchenHandler())
	r.HandleFunc("/druckauftrag-verwerfen", druckauftragCommandHandler.DruckauftragVerwerfenHandler())

	tseQueryHandler := tseHTTP.QueryHandler{}
	tseQueryHandler.Query = tseApp.Query{TSERepo: tseStore}
	r.HandleFunc("/get-tse-signatur-queue", tseQueryHandler.GetTSESignaturQueueHandler())
	r.HandleFunc("/get-tse-stoerungen", tseQueryHandler.GetTSEStoerungenHandler())

	sq := fiskalSetupHTTP.QueryHandler{}
	sq.Query = fiskalSetupApp.Query{
		SettingsRepo: tseStore,
		NewTSEConnectionTester: func(credentials tse.Credentials) (tse.ConnectionTester, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
		NewTSESetupClient: func(credentials tse.SetupCredentials) (tse.SetupClient, error) {
			return tse_repo.NewFiskalyTSESetupClient(cfg.FiskalyBaseURL, credentials, nil)
		},
	}
	r.HandleFunc("/get-kassenidentitaet", sq.GetKassenidentitaetHandler())
	r.HandleFunc("/get-tse-konfiguration", sq.GetTSEKonfigurationHandler())
	r.HandleFunc("/test-tse-verbindung", sq.TestTSEVerbindungHandler())
	r.HandleFunc("/tse-setup-pruefen", sq.PruefeTSESetupHandler())
	r.HandleFunc("/get-tse-status", sq.GetTSEStatusHandler())

	sc := fiskalSetupHTTP.CommandHandler{}
	sc.Command = fiskalSetupApp.Command{
		SettingsRepo:        tseStore,
		KassensitzungenRepo: kassensitzungenRepo,
		NewTSESetupClient: func(credentials tse.SetupCredentials) (tse.SetupClient, error) {
			return tse_repo.NewFiskalyTSESetupClient(cfg.FiskalyBaseURL, credentials, nil)
		},
	}
	r.HandleFunc("/update-tse-konfiguration", sc.UpdateTSEKonfigurationHandler())
	r.HandleFunc("/tse-einrichten", sc.RichteTSEEinHandler())
	r.HandleFunc("/tse-uebernehmen", sc.UebernimmTSEHandler())

	bq := betreiberHTTP.QueryHandler{}
	bq.Query = betreiberApp.Query{BetreiberRepo: betreiberRepo}
	r.HandleFunc("/get-betreiber", bq.GetBetreiberHandler())

	bc := betreiberHTTP.CommandHandler{}
	bc.Command = betreiberApp.Command{BetreiberRepo: betreiberRepo}
	r.HandleFunc("/update-betreiber", bc.UpdateBetreiberHandler())

	return r
}
