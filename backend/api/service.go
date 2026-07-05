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
	productApp "github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	productHTTP "github.com/nicograef/jotti/backend/api/stammdaten/produkt/http"
	tischApp "github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	tischHTTP "github.com/nicograef/jotti/backend/api/stammdaten/tisch/http"
)

func NewServiceApi(deps Deps) http.Handler {
	r := http.NewServeMux()

	pq := productHTTP.QueryHandler{}
	pq.Query = productApp.Query{ProduktRepo: deps.ProduktRepo}
	r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler())

	tc := tischgeschaeftHTTP.CommandHandler{}
	tc.Command = tischgeschaeftApp.Command{
		TischRepo:           deps.TischRepo,
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		FavoritRepo:         deps.FavoritRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}
	r.HandleFunc("/bestellung-aufnehmen", tc.BestellungAufnehmenHandler())
	r.HandleFunc("/bestellung-umbuchen", tc.BestellungUmbuchenHandler())
	r.HandleFunc("/zahlung-kassieren", tc.ZahlungKassierenHandler())
	r.HandleFunc("/ausgabe-bestaetigen", tc.AusgabeBestaetigenHandler())

	bc := belegHTTP.CommandHandler{}
	bc.Command = belegApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
		DruckauftragRepo:    deps.DruckauftragRepo,
		BetreiberRepo:       deps.BetreiberRepo,
		TSERepo:             deps.TSERepo,
	}
	r.HandleFunc("/beleg-drucken", bc.KassenbelegDruckenHandler())

	stc := tischHTTP.CommandHandler{}
	stc.Command = tischApp.Command{
		TischRepo:   deps.TischRepo,
		FavoritRepo: deps.FavoritRepo,
	}
	r.HandleFunc("/favorit-hinzufuegen", stc.FavoritHinzufuegenHandler())
	r.HandleFunc("/favorit-entfernen", stc.FavoritEntfernenHandler())

	dc := direktverkaufHTTP.CommandHandler{}
	dc.Command = direktverkaufApp.Command{
		EventRepo:           deps.KassenjournalRepo,
		ProduktRepo:         deps.ProduktRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		DruckstationRepo:    deps.DruckstationRepo,
	}
	r.HandleFunc("/direktverkauf-taetigen", dc.DirektverkaufTaetigenHandler())

	dq := direktverkaufHTTP.QueryHandler{}
	dq.Query = direktverkaufApp.Query{
		EventRepo:           deps.KassenjournalRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
	}
	r.HandleFunc("/get-direktverkauf-historie", dq.GetDirektverkaufHistorieHandler())

	tq := tischgeschaeftHTTP.QueryHandler{}
	tq.Query = tischgeschaeftApp.Query{TischRepo: deps.TischRepo, EventRepo: deps.KassenjournalRepo, FavoritRepo: deps.FavoritRepo, KassensitzungenRepo: deps.KassensitzungenRepo}
	r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
	r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
	r.HandleFunc("/get-tisch-state", tq.GetTischStateHandler())
	r.HandleFunc("/get-aktive-tische-mit-favoriten", tq.GetAktiveTischeMitFavoritenHandler())
	r.HandleFunc("/get-meine-tische-state", tq.GetMeineTischeStateHandler())

	rq := reportingHTTP.QueryHandler{}
	rq.Query = reportingApp.Query{
		ReportingRepo:       deps.ReportingRepo,
		KassensitzungenRepo: deps.KassensitzungenRepo,
		TischSessionRepo:    deps.KassenjournalRepo,
		TischRepo:           deps.TischRepo,
	}
	r.HandleFunc("/get-eigene-uebersicht", rq.GetEigeneUebersichtHandler())

	return r
}
