package api

import (
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
)

func NewAdminApi(deps Deps) http.Handler {
	r := http.NewServeMux()

	uc := userHTTP.CommandHandler{}
	uc.Command = userApp.Command{UserRepo: deps.UserRepo}
	r.HandleFunc("/create-user", uc.CreateUserHandler())
	r.HandleFunc("/update-user", uc.UpdateUserHandler())
	r.HandleFunc("/activate-user", uc.ActivateUserHandler())
	r.HandleFunc("/deactivate-user", uc.DeactivateUserHandler())
	r.HandleFunc("/delete-user", uc.DeleteUserHandler())
	r.HandleFunc("/reset-password", uc.ResetPasswordHandler())

	uq := userHTTP.QueryHandler{}
	uq.Query = userApp.Query{UserRepo: deps.UserRepo}
	r.HandleFunc("/get-all-users", uq.GetAllUsersHandler())

	pc := productHTTP.CommandHandler{}
	pc.Command = productApp.Command{ProductRepo: deps.ProduktRepo}
	r.HandleFunc("/create-produkt", pc.CreateProductHandler())
	r.HandleFunc("/update-produkt", pc.UpdateProductHandler())
	r.HandleFunc("/create-variante", pc.CreateVariantHandler())
	r.HandleFunc("/update-variante", pc.UpdateVariantHandler())
	r.HandleFunc("/activate-variante", pc.ActivateVariantHandler())
	r.HandleFunc("/deactivate-variante", pc.DeactivateVariantHandler())
	r.HandleFunc("/delete-produkt", pc.DeleteProduktHandler())
	r.HandleFunc("/delete-variante", pc.DeleteVarianteHandler())

	pq := productHTTP.QueryHandler{}
	pq.Query = productApp.Query{ProductRepo: deps.ProduktRepo}
	r.HandleFunc("/get-all-produkte", pq.GetAllProductsHandler())

	tc := tischHTTP.CommandHandler{}
	tc.Command = tischApp.Command{
		TableRepo:   deps.TischRepo,
		FavoritRepo: deps.FavoritRepo,
	}
	r.HandleFunc("/update-tisch", tc.TischAktualisierenHandler())
	r.HandleFunc("/create-tisch", tc.TischErstellenHandler())
	r.HandleFunc("/activate-tisch", tc.TischAktivierenHandler())
	r.HandleFunc("/deactivate-tisch", tc.TischDeaktivierenHandler())
	r.HandleFunc("/delete-tisch", tc.TischLoeschenHandler())

	tq := tischHTTP.QueryHandler{}
	tq.Query = tischApp.Query{TableRepo: deps.TischRepo}
	r.HandleFunc("/get-all-tische", tq.GetAllTischeHandler())

	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{
		ReportingRepo:       deps.ReportingRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		TischSessionRepo:    deps.KassenjournalRepo,
		TischRepo:           deps.TischRepo,
	}
	r.HandleFunc("/get-abrechnung", rq.GetReportingHandler())
	r.HandleFunc("/get-all-kassensitzungen", rq.GetAllKassensitzungenHandler())
	r.HandleFunc("/get-live-reporting", rq.GetLiveReportingHandler())

	exportHandler := exportHTTP.Handler{}
	exportHandler.Service = exportApp.Export{
		KassenjournalRepo:   deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		BetreiberRepo:       deps.BetreiberRepo,
		TSERepo:             deps.TSERepo,
		TableRepo:           deps.TischRepo,
	}
	r.HandleFunc("/export/dsfinvk", exportHandler.ExportHandler())

	kc := kasseHTTP.CommandHandler{}
	kc.Command = kasseApp.Command{
		KassenjournalRepo:   deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		BetreiberRepo:       deps.BetreiberRepo,
		TSERepo:             deps.TSERepo,
	}
	r.HandleFunc("/kassensitzung-eroeffnen", kc.KassensitzungEroeffnenHandler())
	r.HandleFunc("/geldtransit-buchen", kc.GeldtransitBuchenHandler())
	r.HandleFunc("/kasse-abschliessen", kc.KasseAbschliessenHandler())

	kq := kasseHTTP.QueryHandler{}
	kq.Query = kasseApp.Query{KassenjournalRepo: deps.KassenjournalRepo, KassensitzungenRepo: deps.KassensitzungenRepo}
	r.HandleFunc("/get-offene-kassensitzung", kq.GetOffeneKassensitzungHandler())
	r.HandleFunc("/get-kassenbestand", kq.GetKassenbestandHandler())

	druckstationCommandHandler := druckstationHTTP.CommandHandler{}
	druckstationCommandHandler.Command = druckstationApp.Command{DruckstationRepo: deps.DruckstationRepo}
	druckstationQueryHandler := druckstationHTTP.QueryHandler{}
	druckstationQueryHandler.Query = druckstationApp.Query{DruckstationRepo: deps.DruckstationRepo}
	r.HandleFunc("/get-druckstationen", druckstationQueryHandler.GetDruckstationenHandler())
	r.HandleFunc("/update-druckstationen", druckstationCommandHandler.UpdateDruckstationenHandler())

	druckauftragCommandHandler := druckauftragHTTP.CommandHandler{}
	druckauftragCommandHandler.Command = druckauftragApp.Command{DruckauftragRepo: deps.DruckauftragRepo}
	druckauftragQueryHandler := druckauftragHTTP.QueryHandler{}
	druckauftragQueryHandler.Query = druckauftragApp.Query{DruckauftragRepo: deps.DruckauftragRepo}
	r.HandleFunc("/get-fehlgeschlagene-druckauftraege", druckauftragQueryHandler.GetFehlgeschlageneDruckauftraegeHandler())
	r.HandleFunc("/druckauftrag-erneut-versuchen", druckauftragCommandHandler.DruckauftragErneutVersuchenHandler())
	r.HandleFunc("/druckauftrag-verwerfen", druckauftragCommandHandler.DruckauftragVerwerfenHandler())

	tseQueryHandler := tseHTTP.QueryHandler{}
	tseQueryHandler.Query = tseApp.Query{TSERepo: deps.TSERepo}
	r.HandleFunc("/get-tse-signatur-queue", tseQueryHandler.GetTSESignaturQueueHandler())
	r.HandleFunc("/get-tse-stoerungen", tseQueryHandler.GetTSEStoerungenHandler())

	sq := fiskalSetupHTTP.QueryHandler{}
	sq.Query = fiskalSetupApp.Query{
		SettingsRepo:           deps.TSERepo,
		NewTSEConnectionTester: deps.NewTSEConnectionTester,
		NewTSESetupClient:      deps.NewTSESetupClient,
	}
	r.HandleFunc("/get-kassenidentitaet", sq.GetKassenidentitaetHandler())
	r.HandleFunc("/get-tse-konfiguration", sq.GetTSEKonfigurationHandler())
	r.HandleFunc("/test-tse-verbindung", sq.TestTSEVerbindungHandler())
	r.HandleFunc("/tse-setup-pruefen", sq.PruefeTSESetupHandler())
	r.HandleFunc("/get-tse-status", sq.GetTSEStatusHandler())

	sc := fiskalSetupHTTP.CommandHandler{}
	sc.Command = fiskalSetupApp.Command{
		SettingsRepo:        deps.TSERepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		NewTSESetupClient:   deps.NewTSESetupClient,
	}
	r.HandleFunc("/update-tse-konfiguration", sc.UpdateTSEKonfigurationHandler())
	r.HandleFunc("/tse-einrichten", sc.RichteTSEEinHandler())
	r.HandleFunc("/tse-uebernehmen", sc.UebernimmTSEHandler())

	bq := betreiberHTTP.QueryHandler{}
	bq.Query = betreiberApp.Query{BetreiberRepo: deps.BetreiberRepo}
	r.HandleFunc("/get-betreiber", bq.GetBetreiberHandler())

	bc := betreiberHTTP.CommandHandler{}
	bc.Command = betreiberApp.Command{BetreiberRepo: deps.BetreiberRepo}
	r.HandleFunc("/update-betreiber", bc.UpdateBetreiberHandler())

	return r
}
