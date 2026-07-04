package api

import (
	"database/sql"
	"net/http"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	tableApp "github.com/nicograef/jotti/backend/api/table/application"
	tableHTTP "github.com/nicograef/jotti/backend/api/table/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/settings_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

func NewServiceleitungApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	tableRepo := table_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	productRepo := product_repo.NewRepository(db)
	settingsRepo := settings_repo.NewRepository(db)
	tseRepo := tse_repo.NewRepository(db)
	tc := tableHTTP.CommandHandler{}
	tc.Command = tableApp.Command{
		TableRepo:           tableRepo,
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		SettingsRepo:        settingsRepo,
		TSERepo:             tseRepo,
	}
	r.HandleFunc("/stornierung-erteilen", tc.StornierungErteilenHandler())

	dc := direktverkaufHTTP.CommandHandler{}
	dc.Command = direktverkaufApp.Command{
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		KassensitzungenRepo: kassensitzungenRepo,
	}
	r.HandleFunc("/direktverkauf-stornieren", dc.DirektverkaufStornierenHandler())

	return r
}
