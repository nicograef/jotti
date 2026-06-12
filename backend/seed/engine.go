package seed

import (
	"fmt"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// seedEvent bündelt ein Domain-Event mit der Kassensitzungs-Zuordnung, die der Writer für die
// Persistenz im Kassenjournal benötigt.
type seedEvent struct {
	event           e.Event
	kassensitzungNr int
}

// kassensitzungZeile ist die CRUD-Zeile der kassensitzungen-Tabelle. Sie muss vor den zugehörigen
// Events angelegt werden (Fremdschlüssel kassenjournal.kassensitzung_nr → kassensitzungen.z_nr).
type kassensitzungZeile struct {
	ZNr         int
	Datum       time.Time
	Bezeichnung string
	Status      kasse.KassensitzungStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// seedDaten ist das Ergebnis der Engine: die anzulegenden Kassensitzungs-Zeilen und die Events.
type seedDaten struct {
	Kassensitzungen []kassensitzungZeile
	Events          []seedEvent
}

// buildSeedDaten übersetzt das Szenario deterministisch in Events und Kassensitzungs-Zeilen.
// jetzt ist der Bezugszeitpunkt: alle Events liegen in einem Fenster davor, sind streng monoton
// steigend und tragen je Subject lückenlose Versionen ab 1. Alle Events entstehen über die
// Domain-Konstruktoren, sodass dieselben Invarianten wie im Produktivbetrieb gelten.
func buildSeedDaten(s szenario, jetzt time.Time) (seedDaten, error) {
	variantenIdx := variantenIndex(s.Produkte)
	benutzerIdx := benutzerIndex(s.Benutzer)

	sitzung := s.Sitzung
	sitzungStart := jetzt.Add(-5 * time.Hour)
	datum := sitzungStart.Format("2006-01-02")

	// tick liefert für jedes Event einen neuen, streng monoton steigenden Zeitstempel im Fenster.
	naechsteZeit := sitzungStart
	tick := func() time.Time {
		t := naechsteZeit
		naechsteZeit = naechsteZeit.Add(6 * time.Minute)
		return t
	}

	versionen := map[string]int{}
	var events []seedEvent

	// add weist die nächste Version (je Subject ab 1) und einen Zeitstempel zu und hängt das Event an.
	add := func(evt e.Event) {
		versionen[evt.Subject]++
		evt.Version = versionen[evt.Subject]
		evt.Time = tick()
		events = append(events, seedEvent{event: evt, kassensitzungNr: sitzung.ZNr})
	}

	eroeffnerName, ok := benutzerIdx[sitzung.EroeffnetVon]
	if !ok {
		return seedDaten{}, fmt.Errorf("eröffnender Benutzer %d nicht im Szenario", sitzung.EroeffnetVon)
	}
	kassensitzungSubject := kasse.KassensitzungSubject(sitzung.ZNr)
	eroeffnet, err := kasse.NewKassensitzungEroeffnetEvent(kassensitzungSubject, sitzung.EroeffnetVon, eroeffnerName, datum, sitzung.Bezeichnung, sitzung.AnfangsbestandCents)
	if err != nil {
		return seedDaten{}, fmt.Errorf("kassensitzung eröffnen: %w", err)
	}
	add(eroeffnet)

	for _, verlauf := range sitzung.Tische {
		serviceName, ok := benutzerIdx[verlauf.ServiceUserID]
		if !ok {
			return seedDaten{}, fmt.Errorf("service-Benutzer %d nicht im Szenario", verlauf.ServiceUserID)
		}
		positionen, err := buildPositionen(verlauf.Posten, variantenIdx)
		if err != nil {
			return seedDaten{}, fmt.Errorf("tisch %d: %w", verlauf.TischID, err)
		}
		tischSubject := kasse.TischSessionSubject(sitzung.ZNr, verlauf.TischID)

		bestellung, err := kasse.NewBestellungAufgenommenEvent(tischSubject, verlauf.ServiceUserID, serviceName, positionen, "")
		if err != nil {
			return seedDaten{}, fmt.Errorf("tisch %d bestellung: %w", verlauf.TischID, err)
		}
		add(bestellung)

		if !verlauf.Ausgegeben && !verlauf.Bezahlt {
			continue
		}

		// Der Bestell-Konstruktor erzeugt die PositionIDs intern. Wir spielen das Event in den
		// Tisch-Session-Zustand ein, um genau diese IDs für Ausgabe und Zahlung zu verwenden.
		state, err := kasse.ApplyEvent(kasse.TischSession{}, bestellung)
		if err != nil {
			return seedDaten{}, fmt.Errorf("tisch %d zustand: %w", verlauf.TischID, err)
		}

		if verlauf.Ausgegeben {
			ausgabe, err := kasse.NewAusgabeBestaetigtEvent(tischSubject, verlauf.ServiceUserID, serviceName, state.AusstehendePositionen, "")
			if err != nil {
				return seedDaten{}, fmt.Errorf("tisch %d ausgabe: %w", verlauf.TischID, err)
			}
			add(ausgabe)
		}

		if verlauf.Bezahlt {
			zahlung, err := kasse.NewZahlungKassiertEvent(tischSubject, verlauf.ServiceUserID, serviceName, state.UnbezahltePositionen, state.SaldoCents, "")
			if err != nil {
				return seedDaten{}, fmt.Errorf("tisch %d zahlung: %w", verlauf.TischID, err)
			}
			add(zahlung)
		}
	}

	kassensitzungen := []kassensitzungZeile{{
		ZNr:         sitzung.ZNr,
		Datum:       sitzungStart,
		Bezeichnung: sitzung.Bezeichnung,
		Status:      kasse.KassensitzungOffen,
		CreatedAt:   sitzungStart,
		UpdatedAt:   jetzt,
	}}

	return seedDaten{Kassensitzungen: kassensitzungen, Events: events}, nil
}

// variantenIndex bildet jede Varianten-ID auf eine Position-Vorlage ab (ohne Menge und PositionID).
func variantenIndex(produkte []produkt) map[int]kasse.Position {
	idx := make(map[int]kasse.Position)
	for _, p := range produkte {
		for _, v := range p.Varianten {
			idx[v.ID] = kasse.Position{
				VarianteID:   v.ID,
				ProduktName:  p.Name,
				VarianteName: v.Name,
				Kategorie:    string(p.Kategorie),
				Steuersatz:   string(p.Steuersatz),
				Einzelpreis:  v.PreisCents,
			}
		}
	}
	return idx
}

// benutzerIndex bildet jede Benutzer-ID auf den Anzeigenamen ab.
func benutzerIndex(benutzer []benutzer) map[int]string {
	idx := make(map[int]string, len(benutzer))
	for _, b := range benutzer {
		idx[b.ID] = b.Name
	}
	return idx
}

// buildPositionen löst Bestellposten gegen den Varianten-Index zu vollständigen Positionen auf.
// Die PositionID bleibt leer — sie wird vom Bestell-Konstruktor vergeben.
func buildPositionen(posten []bestellposten, variantenIdx map[int]kasse.Position) ([]kasse.Position, error) {
	positionen := make([]kasse.Position, 0, len(posten))
	for _, p := range posten {
		vorlage, ok := variantenIdx[p.VarianteID]
		if !ok {
			return nil, fmt.Errorf("variante %d nicht im Szenario", p.VarianteID)
		}
		vorlage.Menge = p.Menge
		positionen = append(positionen, vorlage)
	}
	return positionen, nil
}
