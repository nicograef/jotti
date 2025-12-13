package api

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/admin/product"
	"github.com/nicograef/jotti/backend/admin/table"
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

	pc := product.CommandHandler{Command: &product.Command{Persistence: &product.Persistence{DB: db}}}
	r.HandleFunc("/create-product", pc.CreateProductHandler())
	r.HandleFunc("/update-product", pc.UpdateProductHandler())
	r.HandleFunc("/activate-product", pc.ActivateProductHandler())
	r.HandleFunc("/deactivate-product", pc.DeactivateProductHandler())

	pq := product.QueryHandler{Query: &product.Query{Persistence: &product.Persistence{DB: db}}}
	r.HandleFunc("/get-all-products", pq.GetAllProductsHandler())

	tc := table.CommandHandler{Command: &table.Command{Persistence: &table.Persistence{DB: db}}}
	r.HandleFunc("/update-table", tc.UpdateTableHandler())
	r.HandleFunc("/create-table", tc.CreateTableHandler())
	r.HandleFunc("/activate-table", tc.ActivateTableHandler())
	r.HandleFunc("/deactivate-table", tc.DeactivateTableHandler())

	tq := table.QueryHandler{Query: &table.Query{Persistence: &table.Persistence{DB: db}}}
	r.HandleFunc("/get-all-tables", tq.GetAllTablesHandler())
	return r
}
