package seed

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
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

// tagesSummen sammelt die Beträge eines Betriebstags für den Tagesabschluss.
// StornierungenCents umfasst beide Storno-Arten (geldneutrale Korrektur und
// kassenwirksame Warenrücknahme); nur die Warenrücknahme mindert als negativer Umsatz
// den Tagesumsatz und den Bargeldbestand, daher wird sie zusätzlich getrennt geführt.
type tagesSummen struct {
	ZahlungenCents            int
	DirektverkaufCents        int
	DirektverkaufStornosCents int
	StornierungenCents        int
	WarenruecknahmenCents     int
	GeldtransitCents          int
}

// UmsatzGesamtCents folgt der Reporting-Definition: Zahlungen + Direktverkäufe
// − Direktverkauf-Stornos − Warenrücknahmen. Die geldneutrale Korrektur fließt
// nicht ein, die kassenwirksame Warenrücknahme mindert den Umsatz.
func (s tagesSummen) UmsatzGesamtCents() int {
	return s.ZahlungenCents + s.DirektverkaufCents - s.DirektverkaufStornosCents - s.WarenruecknahmenCents
}

// buildSeedDaten übersetzt das Szenario deterministisch in Events und Kassensitzungs-Zeilen.
// jetzt ist der Bezugszeitpunkt: Jede Sitzung belegt ihr eigenes Zeitfenster davor, die
// Zeitstempel sind global streng monoton steigend und jedes Subject trägt lückenlose
// Versionen ab 1. Alle Events entstehen über die Domain-Konstruktoren, sodass dieselben
// Invarianten wie im Produktivbetrieb gelten; abgeschlossene Sitzungen enden mit einem
// Tagesabschluss, dessen Summen aus den erzeugten Tages-Events berechnet sind.
func buildSeedDaten(s szenario, jetzt time.Time) (seedDaten, error) {
	variantenIdx := variantenIndex(s.Produkte)
	benutzerIdx := benutzerIndex(s.Benutzer)
	tischIdx := tischIndex(s.Tische)

	daten := seedDaten{}
	vorherigesEnde := time.Time{}

	for i := range s.Sitzungen {
		sitzung := &s.Sitzungen[i]
		start := sitzung.startZeit(jetzt)
		ende := start.Add(sitzung.Dauer)
		if !vorherigesEnde.IsZero() && !start.After(vorherigesEnde) {
			return seedDaten{}, fmt.Errorf("sitzung %d: Zeitfenster überlappt die vorherige Sitzung", sitzung.ZNr)
		}
		vorherigesEnde = ende

		bauer := &sitzungsBuilder{
			sitzung:       *sitzung,
			varianten:     variantenIdx,
			benutzer:      benutzerIdx,
			tische:        tischIdx,
			versionen:     map[string]int{},
			tischEvents:   map[int][]e.Event{},
			verkaufEvents: map[string][]e.Event{},
		}
		events, err := bauer.buildEvents(start, ende)
		if err != nil {
			return seedDaten{}, fmt.Errorf("sitzung %d: %w", sitzung.ZNr, err)
		}

		if err := verteileZeitstempel(events, start, ende, sitzung.Abgeschlossen, sitzung.Tagesprofil); err != nil {
			return seedDaten{}, fmt.Errorf("sitzung %d: %w", sitzung.ZNr, err)
		}

		for _, evt := range events {
			daten.Events = append(daten.Events, seedEvent{event: evt, kassensitzungNr: sitzung.ZNr})
		}

		status := kasse.KassensitzungOffen
		updatedAt := jetzt
		if sitzung.Abgeschlossen {
			status = kasse.KassensitzungAbgeschlossen
			updatedAt = ende
		}
		daten.Kassensitzungen = append(daten.Kassensitzungen, kassensitzungZeile{
			ZNr:         sitzung.ZNr,
			Datum:       start,
			Bezeichnung: sitzung.Bezeichnung,
			Status:      status,
			CreatedAt:   start,
			UpdatedAt:   updatedAt,
		})
	}

	// Kassenjournal-IDs deterministisch ab 1 vergeben; der Writer persistiert sie explizit.
	// Bondruck-Referenzen und Belegnummern verweisen auf diese IDs.
	for i := range daten.Events {
		daten.Events[i].event.ID = i + 1
	}

	return daten, nil
}

// sitzungsBuilder baut die Event-Folge einer Kassensitzung auf. Er hält dafür den laufenden
// Kassenbestand, die Tagessummen und die bisherigen Events je Tisch bzw. Direktverkauf,
// um Folge-Aktionen (Zahlung, Storno) gegen den tatsächlichen Zustand aufzulösen.
type sitzungsBuilder struct {
	sitzung   kassensitzungDrehbuch
	varianten map[int]kasse.Position
	benutzer  map[int]string
	tische    map[int]tisch

	versionen       map[string]int
	events          []e.Event
	tischEvents     map[int][]e.Event
	verkaufEvents   map[string][]e.Event
	bestandCents    int
	summen          tagesSummen
	kassensturzDone bool
}

func (b *sitzungsBuilder) buildEvents(start, ende time.Time) ([]e.Event, error) {
	name, err := b.benutzerName(b.sitzung.EroeffnetVon)
	if err != nil {
		return nil, err
	}
	subject := kasse.KassensitzungSubject(b.sitzung.ZNr)
	eroeffnet, err := kasse.NewKassensitzungEroeffnetEvent(subject, b.sitzung.EroeffnetVon, name,
		start.Format("2006-01-02"), b.sitzung.Bezeichnung, b.sitzung.AnfangsbestandCents)
	if err != nil {
		return nil, fmt.Errorf("kassensitzung eröffnen: %w", err)
	}
	b.add(eroeffnet)
	b.bestandCents += b.sitzung.AnfangsbestandCents

	for i, a := range b.sitzung.Aktionen {
		if err := b.process(a); err != nil {
			return nil, fmt.Errorf("aktion %d (%T): %w", i, a, err)
		}
	}

	if b.sitzung.Abgeschlossen {
		if err := b.schliesseTagAb(start, ende); err != nil {
			return nil, err
		}
	}

	return b.events, nil
}

func (b *sitzungsBuilder) process(a aktion) error {
	switch a := a.(type) {
	case bestellen:
		return b.bestellen(a)
	case kassieren:
		return b.kassieren(a)
	case stornieren:
		return b.stornieren(a)
	case umbuchen:
		return b.umbuchen(a)
	case direktverkauf:
		return b.direktverkauf(a)
	case direktverkaufStorno:
		return b.direktverkaufStorno(a)
	case geldtransit:
		return b.geldtransit(a)
	case kassensturz:
		return b.kassensturz(a)
	default:
		return fmt.Errorf("unbekannter Aktionstyp %T", a)
	}
}

func (b *sitzungsBuilder) bestellen(a bestellen) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	positionen, err := buildPositionen(a.Posten, b.varianten)
	if err != nil {
		return err
	}
	evt, err := kasse.NewBestellungAufgenommenEvent(b.tischSubject(a.Tisch), a.User, name, uuid.New().String(), positionen, a.Kommentar)
	if err != nil {
		return err
	}
	b.addTisch(a.Tisch, evt)
	return nil
}

func (b *sitzungsBuilder) kassieren(a kassieren) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	state, err := b.tischState(a.Tisch)
	if err != nil {
		return err
	}
	auswahl, err := selectPositionen(state.UnbezahltePositionen, a.Posten)
	if err != nil {
		return fmt.Errorf("unbezahlte Positionen: %w", err)
	}
	betrag := summeCents(auswahl)
	evt, err := kasse.NewZahlungKassiertEvent(b.tischSubject(a.Tisch), a.User, name, auswahl, betrag, "")
	if err != nil {
		return err
	}
	b.addTisch(a.Tisch, evt)
	b.bestandCents += betrag
	b.summen.ZahlungenCents += betrag
	return nil
}

// stornieren bildet eine „Stornieren"-Aktion wie im Produktivbetrieb ab: das Routing
// teilt die ausgewählten Positionen nach Bezahlstatus auf — unbezahlte Mengen werden
// geldneutral korrigiert (bestellung-korrigiert), bezahlte Mengen je begleichender
// Zahlung als kassenwirksame Warenrücknahme zurückgenommen (stornierung-erteilt). Nur
// die Warenrücknahme mindert den Bargeldbestand; die Korrektur ist geldneutral.
func (b *sitzungsBuilder) stornieren(a stornieren) error {
	if len(a.Posten) == 0 {
		return fmt.Errorf("stornieren ohne Posten")
	}
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	events := b.tischEvents[a.Tisch]
	kandidaten, err := kasse.ComputeNichtStorniertePositionen(events)
	if err != nil {
		return err
	}
	// Jüngste Positionen zuerst, damit die Varianten-Auswahl die letzte Bestellung trifft.
	slices.Reverse(kandidaten)
	auswahl, err := selectPositionen(kandidaten, a.Posten)
	if err != nil {
		return fmt.Errorf("nicht stornierte Positionen: %w", err)
	}

	refs := make([]kasse.PositionRef, len(auswahl))
	for i, p := range auswahl {
		refs[i] = kasse.PositionRef{PositionID: p.PositionID, Menge: p.Menge}
	}
	aufteilung, ok := kasse.ComputeStornoAufteilung(events, refs)
	if !ok {
		return fmt.Errorf("stornieren: ausgewählte Positionen nicht stornierbar (Tisch %d)", a.Tisch)
	}

	subject := b.tischSubject(a.Tisch)
	if len(aufteilung.Korrektur) > 0 {
		evt, err := kasse.NewBestellungKorrigiertEvent(subject, a.User, name, aufteilung.Korrektur, aufteilung.KorrekturCents, a.Kommentar)
		if err != nil {
			return err
		}
		b.addTisch(a.Tisch, evt)
		b.summen.StornierungenCents += aufteilung.KorrekturCents
	}
	for _, wr := range aufteilung.Warenruecknahmen {
		evt, err := kasse.NewStornierungErteiltEvent(subject, a.User, name, wr.ZahlungID, wr.Positionen, wr.GesamtCents, a.Kommentar)
		if err != nil {
			return err
		}
		b.addTisch(a.Tisch, evt)
		b.bestandCents -= wr.GesamtCents
		b.summen.StornierungenCents += wr.GesamtCents
		b.summen.WarenruecknahmenCents += wr.GesamtCents
	}
	return nil
}

// umbuchen erzeugt das verknüpfte, geldneutrale Umbuchungs-Event-Paar (Abgang auf
// dem Quelltisch, Zugang auf dem Zieltisch) mit den Standard-Kommentaren — wie
// BestellungUmbuchen im Produktivbetrieb. Eine Umbuchung ist kein Storno und fließt
// daher nicht in die Storno-Summe ein.
func (b *sitzungsBuilder) umbuchen(a umbuchen) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	quellTisch, err := b.tischName(a.VonTisch)
	if err != nil {
		return err
	}
	zielTisch, err := b.tischName(a.NachTisch)
	if err != nil {
		return err
	}
	state, err := b.tischState(a.VonTisch)
	if err != nil {
		return err
	}
	auswahl, err := selectPositionen(state.UnbezahltePositionen, a.Posten)
	if err != nil {
		return fmt.Errorf("unbezahlte Positionen: %w", err)
	}
	betrag := summeCents(auswahl)

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(b.sitzung.ZNr, a.VonTisch, a.NachTisch, a.User, name,
		auswahl, betrag, "Umbuchung auf "+zielTisch, "Umbuchung von "+quellTisch, "")
	if err != nil {
		return err
	}
	b.addTisch(a.VonTisch, quellEvent)
	b.addTisch(a.NachTisch, zielEvent)
	return nil
}

func (b *sitzungsBuilder) direktverkauf(a direktverkauf) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	positionen, err := buildPositionen(a.Posten, b.varianten)
	if err != nil {
		return err
	}
	subject := kasse.DirektverkaufSubject(b.sitzung.ZNr, a.VerkaufID)
	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, a.VerkaufID, a.User, name, positionen, a.Kommentar)
	if err != nil {
		return err
	}
	b.addVerkauf(a.VerkaufID, evt)
	betrag := summeCents(positionen)
	b.bestandCents += betrag
	b.summen.DirektverkaufCents += betrag
	return nil
}

func (b *sitzungsBuilder) direktverkaufStorno(a direktverkaufStorno) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	if len(b.verkaufEvents[a.VerkaufID]) == 0 {
		return fmt.Errorf("direktverkauf %s nicht im Drehbuch", a.VerkaufID)
	}
	kandidaten, err := kasse.ComputeNichtStornierteVerkaufPositionen(b.verkaufEvents[a.VerkaufID])
	if err != nil {
		return err
	}
	auswahl, err := selectPositionen(kandidaten, a.Posten)
	if err != nil {
		return fmt.Errorf("nicht stornierte Positionen: %w", err)
	}
	betrag := summeCents(auswahl)
	subject := kasse.DirektverkaufSubject(b.sitzung.ZNr, a.VerkaufID)
	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, a.VerkaufID, a.User, name, auswahl, betrag, a.Kommentar)
	if err != nil {
		return err
	}
	b.addVerkauf(a.VerkaufID, evt)
	b.bestandCents -= betrag
	b.summen.DirektverkaufStornosCents += betrag
	return nil
}

func (b *sitzungsBuilder) geldtransit(a geldtransit) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	evt, err := kasse.NewGeldtransitGebuchtEvent(kasse.KassensitzungSubject(b.sitzung.ZNr), a.User, name,
		uuid.New().String(), a.Richtung, a.BetragCents, a.Kommentar)
	if err != nil {
		return err
	}
	b.add(evt)
	if a.Richtung == "einlage" {
		b.bestandCents += a.BetragCents
		b.summen.GeldtransitCents += a.BetragCents
	} else {
		b.bestandCents -= a.BetragCents
		b.summen.GeldtransitCents -= a.BetragCents
	}
	return nil
}

// kassensturz schreibt den Soll/Ist-Vergleich und bei Differenz ≠ 0 die Differenz-Buchung —
// das Zwei-Event-Muster aus KasseAbschliessen im Produktivbetrieb.
func (b *sitzungsBuilder) kassensturz(a kassensturz) error {
	name, err := b.benutzerName(a.User)
	if err != nil {
		return err
	}
	subject := kasse.KassensitzungSubject(b.sitzung.ZNr)
	soll := b.bestandCents
	ist := soll - a.DifferenzCents
	evt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, a.User, name, soll, ist, a.DifferenzCents)
	if err != nil {
		return err
	}
	b.add(evt)

	if a.DifferenzCents != 0 {
		diff, err := kasse.NewDifferenzSollIstGebuchtEvent(subject, a.User, name, a.DifferenzCents)
		if err != nil {
			return err
		}
		b.add(diff)
		// Differenz = Soll − Ist; die Bargeldwirkung ist Ist − Soll (Fehlbetrag mindert den Bestand).
		b.bestandCents -= a.DifferenzCents
	}

	b.kassensturzDone = true
	return nil
}

// schliesseTagAb prüft die Produktiv-Invarianten (Kassensturz erfolgt, alle Tische
// ausgeglichen) und hängt den Tagesabschluss mit den berechneten Summen ans Fensterende.
func (b *sitzungsBuilder) schliesseTagAb(start, ende time.Time) error {
	if !b.kassensturzDone {
		return fmt.Errorf("tagesabschluss ohne vorherigen Kassensturz")
	}
	for tischID := range b.tischEvents {
		state, err := b.tischState(tischID)
		if err != nil {
			return err
		}
		if state.SaldoCents != 0 {
			return fmt.Errorf("tagesabschluss: Tisch %d hat offenen Saldo %d", tischID, state.SaldoCents)
		}
	}

	name, err := b.benutzerName(b.sitzung.EroeffnetVon)
	if err != nil {
		return err
	}
	evt, err := kasse.NewTagesabschlussErstelltEvent(kasse.KassensitzungSubject(b.sitzung.ZNr),
		b.sitzung.EroeffnetVon, name, b.sitzung.ZNr, start, ende,
		b.summen.UmsatzGesamtCents(), b.summen.StornierungenCents,
		b.summen.GeldtransitCents)
	if err != nil {
		return fmt.Errorf("tagesabschluss: %w", err)
	}
	b.add(evt)
	return nil
}

// add weist dem Event die nächste Version seines Subjects zu, hängt es an die Sitzung an
// und gibt das versionierte Event zurück. Der Zeitstempel wird später über
// verteileZeitstempel vergeben.
func (b *sitzungsBuilder) add(evt e.Event) e.Event {
	b.versionen[evt.Subject]++
	evt.Version = b.versionen[evt.Subject]
	b.events = append(b.events, evt)
	return evt
}

func (b *sitzungsBuilder) addTisch(tischID int, evt e.Event) {
	b.tischEvents[tischID] = append(b.tischEvents[tischID], b.add(evt))
}

func (b *sitzungsBuilder) addVerkauf(verkaufID string, evt e.Event) {
	b.verkaufEvents[verkaufID] = append(b.verkaufEvents[verkaufID], b.add(evt))
}

// tischState spielt die bisherigen Events des Tischs in den Projektions-Zustand ein.
func (b *sitzungsBuilder) tischState(tischID int) (kasse.TischSession, error) {
	state := kasse.TischSession{}
	for _, evt := range b.tischEvents[tischID] {
		var err error
		state, err = kasse.ApplyEvent(state, evt)
		if err != nil {
			return kasse.TischSession{}, fmt.Errorf("tisch %d zustand: %w", tischID, err)
		}
	}
	return state, nil
}

func (b *sitzungsBuilder) tischSubject(tischID int) string {
	return kasse.TischSessionSubject(b.sitzung.ZNr, tischID)
}

func (b *sitzungsBuilder) benutzerName(userID int) (string, error) {
	name, ok := b.benutzer[userID]
	if !ok {
		return "", fmt.Errorf("benutzer %d nicht im Szenario", userID)
	}
	return name, nil
}

func (b *sitzungsBuilder) tischName(tischID int) (string, error) {
	t, ok := b.tische[tischID]
	if !ok {
		return "", fmt.Errorf("tisch %d nicht im Szenario", tischID)
	}
	return t.Name, nil
}

// verteileZeitstempel verteilt die Event-Zeitstempel streng monoton über das Sitzungsfenster.
// Das optionale Tagesprofil staucht die Verteilung zu Stoßzeiten; abgeschlossene Sitzungen
// enden exakt am Fensterende (Tagesabschluss), offene lassen das Fensterende frei.
func verteileZeitstempel(events []e.Event, start, ende time.Time, abgeschlossen bool, profil []profilPunkt) error {
	if err := validateProfil(profil); err != nil {
		return err
	}
	n := len(events)
	if n == 0 {
		return nil
	}
	dauer := ende.Sub(start)
	nenner := float64(n)
	if abgeschlossen && n > 1 {
		nenner = float64(n - 1)
	}
	for i := range events {
		anteil := warp(float64(i)/nenner, profil)
		events[i].Time = start.Add(time.Duration(anteil * float64(dauer)))
	}
	return nil
}

// warp bildet den Event-Anteil u über die stückweise lineare Tagesprofil-Kurve auf den
// Zeit-Anteil ab; (0,0) und (1,1) sind implizite Stützstellen.
func warp(u float64, profil []profilPunkt) float64 {
	vorher := profilPunkt{EventAnteil: 0, ZeitAnteil: 0}
	for _, p := range profil {
		if u <= p.EventAnteil {
			f := (u - vorher.EventAnteil) / (p.EventAnteil - vorher.EventAnteil)
			return vorher.ZeitAnteil + f*(p.ZeitAnteil-vorher.ZeitAnteil)
		}
		vorher = p
	}
	f := (u - vorher.EventAnteil) / (1 - vorher.EventAnteil)
	return vorher.ZeitAnteil + f*(1-vorher.ZeitAnteil)
}

func validateProfil(profil []profilPunkt) error {
	vorher := profilPunkt{EventAnteil: 0, ZeitAnteil: 0}
	for _, p := range profil {
		if p.EventAnteil <= vorher.EventAnteil || p.ZeitAnteil <= vorher.ZeitAnteil || p.EventAnteil >= 1 || p.ZeitAnteil >= 1 {
			return fmt.Errorf("tagesprofil nicht streng monoton steigend in (0,1)")
		}
		vorher = p
	}
	return nil
}

// selectPositionen wählt aus den verfügbaren Positionen die angeforderten Mengen je Variante
// aus (Mengen werden über mehrere Positionen derselben Variante hinweg aufgefüllt).
// Leere posten = alle verfügbaren Positionen vollständig.
func selectPositionen(verfuegbar []kasse.Position, posten []bestellposten) ([]kasse.Position, error) {
	if len(posten) == 0 {
		if len(verfuegbar) == 0 {
			return nil, fmt.Errorf("keine Positionen verfügbar")
		}
		return slices.Clone(verfuegbar), nil
	}

	var auswahl []kasse.Position
	for _, p := range posten {
		rest := p.Menge
		for _, v := range verfuegbar {
			if v.VarianteID != p.VarianteID || rest == 0 {
				continue
			}
			gewaehlt := v
			gewaehlt.Menge = min(rest, v.Menge)
			auswahl = append(auswahl, gewaehlt)
			rest -= gewaehlt.Menge
		}
		if rest > 0 {
			return nil, fmt.Errorf("variante %d: Menge %d nicht verfügbar", p.VarianteID, p.Menge)
		}
	}
	return auswahl, nil
}

func summeCents(positionen []kasse.Position) int {
	summe := 0
	for _, p := range positionen {
		summe += p.EinzelpreisCents * p.Menge
	}
	return summe
}

// variantenIndex bildet jede Varianten-ID auf eine Position-Vorlage ab (ohne Menge und PositionID).
func variantenIndex(produkte []produkt) map[int]kasse.Position {
	idx := make(map[int]kasse.Position)
	for _, p := range produkte {
		for _, v := range p.Varianten {
			idx[v.ID] = kasse.Position{
				VarianteID:       v.ID,
				ProduktName:      p.Name,
				VarianteName:     v.Name,
				Kategorie:        string(p.Kategorie),
				Steuersatz:       string(p.Steuersatz),
				EinzelpreisCents: v.PreisCents,
			}
		}
	}
	return idx
}

// benutzerIndex bildet jede Benutzer-ID auf den Username ab (eingefroren in den Events).
func benutzerIndex(benutzer []benutzer) map[int]string {
	idx := make(map[int]string, len(benutzer))
	for _, b := range benutzer {
		idx[b.ID] = b.Username
	}
	return idx
}

// tischIndex bildet jede Tisch-ID auf den Stammdatensatz ab.
func tischIndex(tische []tisch) map[int]tisch {
	idx := make(map[int]tisch, len(tische))
	for _, t := range tische {
		idx[t.ID] = t
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
