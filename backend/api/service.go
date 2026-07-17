package api

import (
	"net/http"

	belegApp "github.com/nicograef/jotti/backend/api/druck/beleg/application"
	belegHTTP "github.com/nicograef/jotti/backend/api/druck/beleg/http"
	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	direktverkaufHTTP "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/http"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	tischgeschaeftHTTP "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/http"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	reportingHTTP "github.com/nicograef/jotti/backend/api/reporting/http"
	produktApp "github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	produktHTTP "github.com/nicograef/jotti/backend/api/stammdaten/produkt/http"
	tischApp "github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	tischHTTP "github.com/nicograef/jotti/backend/api/stammdaten/tisch/http"
)

func NewServiceApi(deps Deps) (http.Handler, []string) {
	r := newRouteMux()

	pq := produktHTTP.QueryHandler{Query: produktApp.Query{ProduktRepo: deps.ProduktRepo}}
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tc := tischgeschaeftHTTP.CommandHandler{Command: tischgeschaeftApp.Command{
		TischRepo:           deps.TischRepo,
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		FavoritRepo:         deps.FavoritRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}}
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

	stc := tischHTTP.CommandHandler{Command: tischApp.Command{
		TischRepo:   deps.TischRepo,
		FavoritRepo: deps.FavoritRepo,
	}}
	r.HandleFunc("/favorit-hinzufuegen", stc.FavoritHinzufuegenHandler())
	r.HandleFunc("/favorit-entfernen", stc.FavoritEntfernenHandler())

	dc := direktverkaufHTTP.CommandHandler{Command: direktverkaufApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}}
	r.HandleFunc("/direktverkauf-taetigen", dc.DirektverkaufTaetigenHandler())

	dq := direktverkaufHTTP.QueryHandler{Query: direktverkaufApp.Query{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}}
	r.HandleFunc("/get-direktverkauf-historie", dq.GetDirektverkaufHistorieHandler())

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

	rq := reportingHTTP.QueryHandler{Query: reportingApp.Query{
		ReportingRepo:       deps.ReportingRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		TischSessionRepo:    deps.KassenjournalRepo,
		TischRepo:           deps.TischRepo,
	}}
	r.HandleFunc("/get-eigene-uebersicht", rq.GetEigeneUebersichtHandler())

	return r.Handler(), r.Paths()
}
