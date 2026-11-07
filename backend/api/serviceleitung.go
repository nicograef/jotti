package api

import (
	"net/http"

	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	pfandrueckgabeHTTP "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/http"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
)

func NewServiceleitungApi(deps Deps) (http.Handler, []string) {
	r := newRouteMux()

	tc := tischgeschaeftHTTP.CommandHandler{Command: deps.tischgeschaeftCommand()}
	r.HandleFunc("/stornierung-erteilen", tc.StornierungErteilenHandler())

	dc := direktverkaufHTTP.CommandHandler{Command: deps.direktverkaufCommand()}
	r.HandleFunc("/direktverkauf-stornieren", dc.DirektverkaufStornierenHandler())

	pc := pfandrueckgabeHTTP.CommandHandler{Command: deps.pfandrueckgabeCommand()}
	r.HandleFunc("/pfandrueckgabe-stornieren", pc.PfandrueckgabeStornierenHandler())

	return r.Handler(), r.Paths()
}
