package api

import (
	"database/sql"
	"net/http"

	product "github.com/nicograef/jotti/backend/api/product/http"
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
	r.HandleFunc("/reset-password", uc.ResetPasswordHandler())

	uq := user.NewQueryHandler(db)
	r.HandleFunc("/get-all-users", uq.GetAllUsersHandler())

	pc := product.NewCommandHandler(db)
	r.HandleFunc("/create-product", pc.CreateProductHandler())
	r.HandleFunc("/update-product", pc.UpdateProductHandler())
	r.HandleFunc("/activate-product", pc.ActivateProductHandler())
	r.HandleFunc("/deactivate-product", pc.DeactivateProductHandler())

	pq := product.NewQueryHandler(db)
	r.HandleFunc("/get-all-products", pq.GetAllProductsHandler())

	tc := table.NewCommandHandler(db)
	r.HandleFunc("/update-table", tc.UpdateTableHandler())
	r.HandleFunc("/create-table", tc.CreateTableHandler())
	r.HandleFunc("/activate-table", tc.ActivateTableHandler())
	r.HandleFunc("/deactivate-table", tc.DeactivateTableHandler())

	tq := table.NewQueryHandler(db)
	r.HandleFunc("/get-all-tables", tq.GetAllTablesHandler())

	return r
}
