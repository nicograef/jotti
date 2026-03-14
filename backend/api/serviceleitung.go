package api

import (
	"database/sql"
	"net/http"

	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	"github.com/nicograef/jotti/backend/repository/event_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func NewServiceleitungApi(db *sql.DB) http.Handler {
	r := http.NewServeMux()

	tableRepo := table_repo.NewRepository(db)
	eventRepo := event_repo.NewRepository(db)
	productRepo := product_repo.NewRepository(db)
	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:   tableRepo,
		EventRepo:   eventRepo,
		ProductRepo: productRepo,
	}
	r.HandleFunc("/stornierung-erteilen", tc.StornierungErteilenHandler())

	return r
}
