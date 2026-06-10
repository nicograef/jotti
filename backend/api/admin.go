package api

import (
	"database/sql"
	"net/http"

	druckstationApp "github.com/nicograef/jotti/backend/api/druckstation/application"
	druckstationHTTP "github.com/nicograef/jotti/backend/api/druckstation/http"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/application"
	kasseHTTP "github.com/nicograef/jotti/backend/api/kasse/http"
	productApp "github.com/nicograef/jotti/backend/api/product/application"
	productHTTP "github.com/nicograef/jotti/backend/api/product/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	settingsApp "github.com/nicograef/jotti/backend/api/settings/application"
	settingsHTTP "github.com/nicograef/jotti/backend/api/settings/http"
	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	userApp "github.com/nicograef/jotti/backend/api/user/application"
	userHTTP "github.com/nicograef/jotti/backend/api/user/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
	"github.com/nicograef/jotti/backend/repository/settings_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
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

	productRepo := product_repo.NewRepository(db)
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

	tableRepo := table_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	settingsRepo := settings_repo.NewRepository(db)
	tseStore := tse_repo.NewStore(db)
	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:           tableRepo,
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		KassensitzungenRepo: kassensitzungenRepo,
	}
	r.HandleFunc("/update-tisch", tc.TischAktualisierenHandler())
	r.HandleFunc("/create-tisch", tc.TischErstellenHandler())
	r.HandleFunc("/activate-tisch", tc.TischAktivierenHandler())
	r.HandleFunc("/deactivate-tisch", tc.TischDeaktivierenHandler())
	r.HandleFunc("/delete-tisch", tc.TischLoeschenHandler())

	tq := tableHTTP.QueryHandler{}
	tq.Query = tableApp.Query{TableRepo: tableRepo, EventRepo: kassenjournalRepo, KassensitzungenRepo: kassensitzungenRepo}
	r.HandleFunc("/get-all-tische", tq.GetAllTischeHandler())

	reportingRepo := reporting_repo.NewRepository(db)
	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{
		ReportingRepo:       reportingRepo,
		KassensitzungenRepo: kassensitzungenRepo,
	}
	r.HandleFunc("/get-abrechnung", rq.GetReportingHandler())
	r.HandleFunc("/get-all-kassensitzungen", rq.GetAllKassensitzungenHandler())
	r.HandleFunc("/get-live-reporting", rq.GetLiveReportingHandler())

	kc := kasseHTTP.CommandHandler{}
	kc.Command = kasseApp.Command{
		KassenjournalRepo:   kassenjournalRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		SettingsRepo:        settingsRepo,
		NewTSEClient: func(credentials tse.Credentials) (tse.TSEClient, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
	}
	r.HandleFunc("/kassensitzung-eroeffnen", kc.KassensitzungEroeffnenHandler())
	r.HandleFunc("/geldtransit-buchen", kc.GeldtransitBuchenHandler())
	r.HandleFunc("/kassensturz-durchfuehren", kc.KassensturzDurchfuehrenHandler())
	r.HandleFunc("/tagesabschluss-erstellen", kc.TagesabschlussErstellenHandler())

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

	sq := settingsHTTP.QueryHandler{}
	sq.Query = settingsApp.Query{
		SettingsRepo:  settingsRepo,
		TSEStatusRepo: tseStore,
		NewTSEConnectionTester: func(credentials tse.Credentials) (tse.ConnectionTester, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
	}
	r.HandleFunc("/get-kassenidentitaet", sq.GetKassenidentitaetHandler())
	r.HandleFunc("/get-betreiber", sq.GetBetreiberHandler())
	r.HandleFunc("/get-bondruck-einstellungen", sq.GetBondruckEinstellungenHandler())
	r.HandleFunc("/get-tse-konfiguration", sq.GetTSEKonfigurationHandler())
	r.HandleFunc("/test-tse-verbindung", sq.TestTSEVerbindungHandler())
	r.HandleFunc("/get-tse-status", sq.GetTSEStatusHandler())

	sc := settingsHTTP.CommandHandler{}
	sc.Command = settingsApp.Command{SettingsRepo: settingsRepo}
	r.HandleFunc("/update-betreiber", sc.UpdateBetreiberHandler())
	r.HandleFunc("/update-bondruck-einstellungen", sc.UpdateBondruckEinstellungenHandler())
	r.HandleFunc("/update-tse-konfiguration", sc.UpdateTSEKonfigurationHandler())

	return r
}
