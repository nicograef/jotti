package api

import (
	"database/sql"
	"net/http"

	product "github.com/nicograef/jotti/backend/api/product/http"
	reporting "github.com/nicograef/jotti/backend/api/reporting/http"
	table "github.com/nicograef/jotti/backend/api/table/http"
	user "github.com/nicograef/jotti/backend/api/user/http"
)

func NewAdminApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	uc := user.NewCommandHandler(db)
	r.HandleFunc("/create-user", uc.CreateUserHandler())
	r.HandleFunc("/update-user", uc.UpdateUserHandler())
	r.HandleFunc("/activate-user", uc.ActivateUserHandler())
	r.HandleFunc("/deactivate-user", uc.DeactivateUserHandler())
	r.HandleFunc("/delete-user", uc.DeleteUserHandler())
	r.HandleFunc("/reset-password", uc.ResetPasswordHandler())

	uq := user.NewQueryHandler(db)
	r.HandleFunc("/get-all-users", uq.GetAllUsersHandler())

	pc := product.NewCommandHandler(db)
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

	pq := product.NewQueryHandler(db)
	r.HandleFunc("/get-all-produkte", pq.GetAllProductsHandler())

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/update-tisch", tc.TischAktualisierenHandler())
	r.HandleFunc("/create-tisch", tc.TischErstellenHandler())
	r.HandleFunc("/activate-tisch", tc.TischAktivierenHandler())
	r.HandleFunc("/deactivate-tisch", tc.TischDeaktivierenHandler())
	r.HandleFunc("/delete-tisch", tc.TischLoeschenHandler())

	tq := table.NewQueryHandler(db)
	r.HandleFunc("/get-all-tische", tq.GetAllTischeHandler())

	rq := reporting.NewQueryHandler(db)
	r.HandleFunc("/get-dashboard", rq.GetDashboardHandler())
	r.HandleFunc("/get-tagesabrechnung", rq.GetTagesabrechnungHandler())

	return r
}
