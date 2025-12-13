package api

import (
	"database/sql"
	"net/http"

	product "github.com/nicograef/jotti/backend/admin/product/http"
	table "github.com/nicograef/jotti/backend/admin/table/http"
	"github.com/nicograef/jotti/backend/admin/user"
)

func NewApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	uc := user.CommandHandler{Command: &user.Command{Persistence: &user.Persistence{DB: db}}}
	r.HandleFunc("/create-user", uc.CreateUserHandler())
	r.HandleFunc("/update-user", uc.UpdateUserHandler())
	r.HandleFunc("/activate-user", uc.ActivateUserHandler())
	r.HandleFunc("/deactivate-user", uc.DeactivateUserHandler())
	r.HandleFunc("/reset-password", uc.ResetPasswordHandler())
	uq := user.QueryHandler{Query: &user.Query{Persistence: &user.Persistence{DB: db}}}
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
