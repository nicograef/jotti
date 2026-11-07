package api

import (
	"net/http"

	belegApp "github.com/nicograef/jotti/backend/api/druck/beleg/application"
	belegHTTP "github.com/nicograef/jotti/backend/api/druck/beleg/http"
	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	pfandrueckgabeApp "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/application"
	pfandrueckgabeHTTP "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	produktHTTP "github.com/nicograef/jotti/backend/api/stammdaten/produkt/http"
	tischHTTP "github.com/nicograef/jotti/backend/api/stammdaten/tisch/http"
)

func NewServiceApi(deps Deps) (http.Handler, []string) {
	r := newRouteMux()

	pq := produktHTTP.QueryHandler{Query: deps.produktQuery()}
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProdukteHandler())

	tc := tischgeschaeftHTTP.CommandHandler{Command: deps.tischgeschaeftCommand()}
	r.HandleFunc("/bestellung-aufnehmen", tc.BestellungAufnehmenHandler())
	r.HandleFunc("/bestellung-umbuchen", tc.BestellungUmbuchenHandler())
	r.HandleFunc("/zahlung-kassieren", tc.ZahlungKassierenHandler())

	bc := belegHTTP.CommandHandler{Command: belegApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
		DruckauftragRepo:    deps.DruckauftragRepo,
		BetreiberRepo:       deps.BetreiberRepo,
		TSERepo:             deps.TSERepo,
	}}
	r.HandleFunc("/beleg-drucken", bc.KassenbelegDruckenHandler())

	stc := tischHTTP.CommandHandler{Command: deps.tischCommand()}
	r.HandleFunc("/favorit-hinzufuegen", stc.FavoritHinzufuegenHandler())
	r.HandleFunc("/favorit-entfernen", stc.FavoritEntfernenHandler())

	dc := direktverkaufHTTP.CommandHandler{Command: deps.direktverkaufCommand()}
	r.HandleFunc("/direktverkauf-taetigen", dc.DirektverkaufTaetigenHandler())

	pc := pfandrueckgabeHTTP.CommandHandler{Command: deps.pfandrueckgabeCommand()}
	r.HandleFunc("/pfand-zurueckgeben", pc.PfandZurueckgebenHandler())

	dq := direktverkaufHTTP.QueryHandler{Query: direktverkaufApp.Query{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}}
	r.HandleFunc("/get-direktverkauf-historie", dq.GetDirektverkaufHistorieHandler())

	pfq := pfandrueckgabeHTTP.QueryHandler{Query: pfandrueckgabeApp.Query{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}}
	r.HandleFunc("/get-pfandrueckgabe-historie", pfq.GetPfandrueckgabeHistorieHandler())

	tq := tischgeschaeftHTTP.QueryHandler{Query: tischgeschaeftApp.Query{
		TischRepo:           deps.TischRepo,
		EventRepo:           deps.KassenjournalRepo,
		FavoritRepo:         deps.FavoritRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}}
	r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
	r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
	r.HandleFunc("/get-tisch-state", tq.GetTischStateHandler())
	r.HandleFunc("/get-aktive-tische-mit-favoriten", tq.GetAktiveTischeMitFavoritenHandler())
	r.HandleFunc("/get-meine-tische-state", tq.GetMeineTischeStateHandler())

	rq := reportingHTTP.QueryHandler{Query: deps.reportingQuery()}
	r.HandleFunc("/get-eigene-uebersicht", rq.GetEigeneUebersichtHandler())

	return r.Handler(), r.Paths()
}
