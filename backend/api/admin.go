package api

import (
	"database/sql"
	"net/http"

	druckerApp "github.com/nicograef/jotti/backend/api/drucker/application"
	druckerHTTP "github.com/nicograef/jotti/backend/api/drucker/http"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/application"
	kasseHTTP "github.com/nicograef/jotti/backend/api/kasse/http"
	productApp "github.com/nicograef/jotti/backend/api/product/application"
	productHTTP "github.com/nicograef/jotti/backend/api/product/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	userApp "github.com/nicograef/jotti/backend/api/user/application"
	userHTTP "github.com/nicograef/jotti/backend/api/user/http"
	"github.com/nicograef/jotti/backend/repository/drucker_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func NewAdminApi(db *sql.DB) http.Handler {
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
	r.HandleFunc("/activate-produkt", pc.ActivateProductHandler())
	r.HandleFunc("/deactivate-produkt", pc.DeactivateProductHandler())
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
	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:   tableRepo,
		EventRepo:   kassenjournalRepo,
		ProductRepo: productRepo,
	}
	r.HandleFunc("/update-tisch", tc.TischAktualisierenHandler())
	r.HandleFunc("/create-tisch", tc.TischErstellenHandler())
	r.HandleFunc("/activate-tisch", tc.TischAktivierenHandler())
	r.HandleFunc("/deactivate-tisch", tc.TischDeaktivierenHandler())
	r.HandleFunc("/delete-tisch", tc.TischLoeschenHandler())

	tq := tableHTTP.QueryHandler{}
	tq.Query = tableApp.Query{TableRepo: tableRepo, EventRepo: kassenjournalRepo}
	r.HandleFunc("/get-all-tische", tq.GetAllTischeHandler())

	reportingRepo := reporting_repo.NewRepository(db)
	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{ReportingRepo: reportingRepo, KasseRepo: kassenjournalRepo}
	r.HandleFunc("/get-abrechnung", rq.GetReportingHandler())

	kh := kasseHTTP.Handler{}
	kh.Command = kasseApp.Command{KassenRepo: kassenjournalRepo}
	kh.Query = kasseApp.Query{KassenRepo: kassenjournalRepo}
	r.HandleFunc("/kassensitzung-eroeffnen", kh.KassensitzungEroeffnenHandler())
	r.HandleFunc("/anfangsbestand-setzen", kh.AnfangsbestandSetzenHandler())
	r.HandleFunc("/kassenbewegung-buchen", kh.KassenbewegungBuchenHandler())
	r.HandleFunc("/kassensturz-durchfuehren", kh.KassensturzDurchfuehrenHandler())
	r.HandleFunc("/tagesabschluss-erstellen", kh.TagesabschlussErstellenHandler())
	r.HandleFunc("/get-offene-kassensitzung", kh.GetOffeneKassensitzungHandler())
	r.HandleFunc("/get-kassenbestand", kh.GetKassenbestandHandler())

	druckerRepo := drucker_repo.NewRepository(db)
	dc := druckerHTTP.CommandHandler{}
	dc.Command = druckerApp.Command{DruckerRepo: druckerRepo}
	dq := druckerHTTP.QueryHandler{}
	dq.Query = druckerApp.Query{DruckerRepo: druckerRepo}
	r.HandleFunc("/get-drucker-konfiguration", dq.GetDruckerConfigHandler())
	r.HandleFunc("/update-drucker-konfiguration", dc.UpdateDruckerConfigHandler())

	return r
}
