package api

import (
	"net/http"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
)

func NewServiceleitungApi(deps Deps) http.Handler {
	r := http.NewServeMux()

	tc := tischgeschaeftHTTP.CommandHandler{}
	tc.Command = tischgeschaeftApp.Command{
		TischRepo:           deps.TischRepo,
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		FavoritRepo:         deps.FavoritRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}
	r.HandleFunc("/stornierung-erteilen", tc.StornierungErteilenHandler())

	dc := direktverkaufHTTP.CommandHandler{}
	dc.Command = direktverkaufApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}
	r.HandleFunc("/direktverkauf-stornieren", dc.DirektverkaufStornierenHandler())

	return r
}
