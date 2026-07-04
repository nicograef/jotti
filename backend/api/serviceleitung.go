package api

import (
	"database/sql"
	"net/http"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/favorit_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func NewServiceleitungApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	tableRepo := table_repo.NewRepository(db)
	kassenjournalRepo := kassenjournal_repo.NewRepository(db)
	kassensitzungenRepo := kassensitzungen_repo.NewRepository(db)
	productRepo := product_repo.NewRepository(db)
	favoritRepo := favorit_repo.NewRepository(db)
	druckstationRepo := druckstation_repo.NewRepository(db)
	tc := tischgeschaeftHTTP.CommandHandler{}
	tc.Command = tischgeschaeftApp.Command{
		TableRepo:           tableRepo,
		EventRepo:           kassenjournalRepo,
		ProductRepo:         productRepo,
		FavoritRepo:         favoritRepo,
		KassensitzungenRepo: kassensitzungenRepo,
		DruckstationRepo:    druckstationRepo,
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
