package api

import (
	"net/http"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	pfandrueckgabeApp "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/application"
	pfandrueckgabeHTTP "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
)

func NewServiceleitungApi(deps Deps) (http.Handler, []string) {
	r := newRouteMux()

	tc := tischgeschaeftHTTP.CommandHandler{Command: tischgeschaeftApp.Command{
		TischRepo:           deps.TischRepo,
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}}
	r.HandleFunc("/stornierung-erteilen", tc.StornierungErteilenHandler())

	dvCommand := direktverkaufApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}
	dc := direktverkaufHTTP.CommandHandler{Command: dvCommand}
	r.HandleFunc("/direktverkauf-stornieren", dc.DirektverkaufStornierenHandler())

	pc := pfandrueckgabeHTTP.CommandHandler{Command: pfandrueckgabeApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		TischRepo:           deps.TischRepo,
	}}
	r.HandleFunc("/pfandrueckgabe-stornieren", pc.PfandrueckgabeStornierenHandler())

	return r.Handler(), r.Paths()
}
