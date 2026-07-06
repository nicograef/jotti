//go:build unit

package seed

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// testJetzt ist ein fester Bezugszeitpunkt, damit die Tests deterministisch laufen.
var testJetzt = time.Date(2026, 6, 14, 16, 0, 0, 0, time.UTC)

func buildTestDaten(t *testing.T) (szenario, seedDaten) {
	t.Helper()
	s := demoSzenario()
	daten, err := buildSeedDaten(s, testJetzt)
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
	return s, daten
}

func eventsProSitzung(daten seedDaten) map[int][]e.Event {
	events := map[int][]e.Event{}
	for _, ev := range daten.Events {
		events[ev.kassensitzungNr] = append(events[ev.kassensitzungNr], ev.event)
	}
	return events
}

func parseData[T any](t *testing.T, evt e.Event) T {
	t.Helper()
	var data T
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		t.Fatalf("event %s v%d: data nicht parsebar: %v", evt.Subject, evt.Version, err)
	}
	return data
}

// tischStates spielt alle Tisch-Events einer Kassensitzung subject-weise nach.
func tischStates(t *testing.T, events []e.Event) map[string]kasse.TischSession {
	t.Helper()
	states := map[string]kasse.TischSession{}
	for _, evt := range events {
		if !strings.Contains(evt.Subject, "/tisch-") {
			continue
		}
		st, err := kasse.ApplyEvent(states[evt.Subject], evt)
		if err != nil {
			t.Fatalf("ApplyEvent %s v%d: %v", evt.Subject, evt.Version, err)
		}
		states[evt.Subject] = st
	}
	return states
}

func TestBuildSeedDaten_VersionenLueckenlosJeSubject(t *testing.T) {
	_, daten := buildTestDaten(t)

	hoechste := map[string]int{}
	gesehen := map[string]map[int]bool{}
	for _, ev := range daten.Events {
		subj := ev.event.Subject
		if gesehen[subj] == nil {
			gesehen[subj] = map[int]bool{}
		}
		if gesehen[subj][ev.event.Version] {
			t.Fatalf("doppelte Version %d für Subject %s", ev.event.Version, subj)
		}
		gesehen[subj][ev.event.Version] = true
		if ev.event.Version > hoechste[subj] {
			hoechste[subj] = ev.event.Version
		}
	}

	for subj, max := range hoechste {
		for v := 1; v <= max; v++ {
			if !gesehen[subj][v] {
				t.Errorf("Subject %s: Version %d fehlt (Lücke)", subj, v)
			}
		}
	}
}

func TestBuildSeedDaten_ZeitstempelMonotonImSitzungsfenster(t *testing.T) {
	s, daten := buildTestDaten(t)
	if len(daten.Events) == 0 {
		t.Fatal("keine Events erzeugt")
	}

	fenster := map[int][2]time.Time{}
	for _, sitzung := range s.Sitzungen {
		start := testJetzt.Add(-sitzung.StartVorJetzt)
		fenster[sitzung.ZNr] = [2]time.Time{start, start.Add(sitzung.Dauer)}
	}

	var vorher time.Time
	for i, ev := range daten.Events {
		ts := ev.event.Time
		f := fenster[ev.kassensitzungNr]
		if ts.Before(f[0]) || ts.After(f[1]) {
			t.Errorf("Event %d (Sitzung %d) Zeitstempel %v außerhalb des Fensters [%v, %v]",
				i, ev.kassensitzungNr, ts, f[0], f[1])
		}
		if i > 0 && !ts.After(vorher) {
			t.Errorf("Event %d Zeitstempel %v nicht streng monoton nach %v", i, ts, vorher)
		}
		vorher = ts
	}
}

func TestBuildSeedDaten_EventsValide(t *testing.T) {
	_, daten := buildTestDaten(t)
	if len(daten.Events) == 0 {
		t.Fatal("keine Events erzeugt")
	}
	for i, ev := range daten.Events {
		evt := ev.event
		if err := evt.Validate(); err != nil {
			t.Errorf("Event %d (%s v%d) ungültig: %v", i, evt.Subject, evt.Version, err)
		}
	}
}

func TestBuildSeedDaten_EventGroessenordnung(t *testing.T) {
	_, daten := buildTestDaten(t)
	events := eventsProSitzung(daten)

	grenzen := map[int][2]int{
		1: {140, 180},
		2: {630, 770},
		3: {110, 150},
	}
	for znr, g := range grenzen {
		anzahl := len(events[znr])
		if anzahl < g[0] || anzahl > g[1] {
			t.Errorf("Sitzung %d: %d Events, erwartet %d–%d", znr, anzahl, g[0], g[1])
		}
	}
}

// TestBuildSeedDaten_SonntagsTischZustaende prüft, dass der offene Sonntag alle Tisch-Zustände
// abdeckt: leer, frisch bestellt, teilgeliefert, teilbezahlt, Warenrücknahme nach Bezahlung,
// abgeschlossen.
func TestBuildSeedDaten_SonntagsTischZustaende(t *testing.T) {
	_, daten := buildTestDaten(t)
	states := tischStates(t, eventsProSitzung(daten)[3])
	subject := func(tischID int) string { return kasse.TischSessionSubject(3, tischID) }

	// Leere Tische: keinerlei Events.
	for _, tischID := range []int{3, 10, 12, 13, 15} {
		if _, ok := states[subject(tischID)]; ok {
			t.Errorf("Tisch %d: Events vorhanden, erwartet leer", tischID)
		}
	}

	// Tisch 14: frisch bestellt — nichts ausgegeben, nichts bezahlt.
	t14 := states[subject(14)]
	if t14.SaldoCents <= 0 || len(t14.AusstehendePositionen) == 0 || t14.GesamtZahlungenCents != 0 {
		t.Errorf("Tisch 14 nicht frisch bestellt: saldo=%d ausstehend=%d zahlungen=%d",
			t14.SaldoCents, len(t14.AusstehendePositionen), t14.GesamtZahlungenCents)
	}

	// Tisch 8: teilgeliefert — offener Saldo, ausstehende und bereits ausgegebene Positionen.
	t8 := states[subject(8)]
	if t8.SaldoCents <= 0 || len(t8.AusstehendePositionen) == 0 || len(t8.AusstehendePositionen) >= len(t8.UnbezahltePositionen) {
		t.Errorf("Tisch 8 nicht teilgeliefert: saldo=%d ausstehend=%d unbezahlt=%d",
			t8.SaldoCents, len(t8.AusstehendePositionen), len(t8.UnbezahltePositionen))
	}

	// Tisch 2: teilbezahlt — Zahlungen geleistet, aber noch offener Saldo.
	t2 := states[subject(2)]
	if t2.GesamtZahlungenCents <= 0 || t2.SaldoCents <= 0 {
		t.Errorf("Tisch 2 nicht teilbezahlt: saldo=%d zahlungen=%d", t2.SaldoCents, t2.GesamtZahlungenCents)
	}

	// Tisch 7: Storno nach Bezahlung (Warenrücknahme) — der offene Betrag bleibt 0.
	if t7 := states[subject(7)]; t7.SaldoCents != 0 {
		t.Errorf("Tisch 7 Saldo = %d, erwartet 0 (Warenrücknahme nach Bezahlung)", t7.SaldoCents)
	}

	// Tisch 9: Storno nach Bezahlung (Warenrücknahme) — ausgeglichen.
	if t9 := states[subject(9)]; t9.SaldoCents != 0 {
		t.Errorf("Tisch 9 Saldo = %d, erwartet 0 (Warenrücknahme nach Bezahlung)", t9.SaldoCents)
	}

	// Abgeschlossene Tische: Bestellungen − Korrekturen − Zahlungen = 0.
	for _, tischID := range []int{1, 5, 9, 18, 19} {
		if st := states[subject(tischID)]; st.SaldoCents != 0 {
			t.Errorf("Tisch %d Saldo = %d, erwartet 0 (abgeschlossen)", tischID, st.SaldoCents)
		}
	}
}

// TestBuildSeedDaten_AbgeschlosseneTageAusgeglichen prüft die Tisch-Saldo-Sperre des
// Produktivbetriebs: An abgeschlossenen Tagen enden alle Tische mit Saldo 0.
func TestBuildSeedDaten_AbgeschlosseneTageAusgeglichen(t *testing.T) {
	_, daten := buildTestDaten(t)
	events := eventsProSitzung(daten)

	for _, znr := range []int{1, 2} {
		for subj, st := range tischStates(t, events[znr]) {
			if st.SaldoCents != 0 {
				t.Errorf("Sitzung %d, %s: Saldo = %d, erwartet 0", znr, subj, st.SaldoCents)
			}
		}
	}
}

// TestBuildSeedDaten_TagesabschlussSummen prüft die Summen der Tagesabschluss-Events gegen
// eine unabhängige Aggregation der erzeugten Tages-Events (Reporting-Formel).
func TestBuildSeedDaten_TagesabschlussSummen(t *testing.T) {
	s, daten := buildTestDaten(t)
	events := eventsProSitzung(daten)

	for _, sitzung := range s.Sitzungen {
		if !sitzung.Abgeschlossen {
			continue
		}
		tagesEvents := events[sitzung.ZNr]
		letztes := tagesEvents[len(tagesEvents)-1]
		if letztes.Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
			t.Fatalf("Sitzung %d: letztes Event ist %s, erwartet Tagesabschluss", sitzung.ZNr, letztes.Type)
		}
		abschluss := parseData[kasse.TagesabschlussErstelltV1Data](t, letztes)

		var zahlungen, warenruecknahmen, korrekturen, dv, dvStorno, transit int
		for _, evt := range tagesEvents {
			switch evt.Type {
			case string(kasse.EventTypeZahlungKassiertV1):
				zahlungen += parseData[kasse.ZahlungKassiertV1Data](t, evt).GesamtZahlungCents
			case string(kasse.EventTypeStornierungErteiltV1):
				warenruecknahmen += parseData[kasse.StornierungErteiltV1Data](t, evt).GesamtStornierungCents
			case string(kasse.EventTypeBestellungKorrigiertV1):
				korrekturen += parseData[kasse.BestellungKorrigiertV1Data](t, evt).GesamtCents
			case string(kasse.EventTypeDirektverkaufGetaetigtV1):
				dv += parseData[kasse.DirektverkaufGetaetigtV1Data](t, evt).GesamtbetragCents
			case string(kasse.EventTypeDirektverkaufStorniertV1):
				dvStorno += parseData[kasse.DirektverkaufStorniertV1Data](t, evt).GesamtStornierungCents
			case string(kasse.EventTypeGeldtransitGebuchtV1):
				data := parseData[kasse.GeldtransitGebuchtV1Data](t, evt)
				if data.Richtung == "einlage" {
					transit += data.BetragCents
				} else {
					transit -= data.BetragCents
				}
			}
		}

		// Die kassenwirksame Warenrücknahme mindert den Umsatz, die geldneutrale
		// Korrektur nicht; StornierungCents umfasst beide Storno-Arten.
		umsatz := zahlungen + dv - dvStorno - warenruecknahmen
		if abschluss.UmsatzGesamtCents != umsatz {
			t.Errorf("Sitzung %d: UmsatzGesamtCents = %d, unabhängig aggregiert %d", sitzung.ZNr, abschluss.UmsatzGesamtCents, umsatz)
		}
		if abschluss.StornierungCents != warenruecknahmen+korrekturen {
			t.Errorf("Sitzung %d: StornierungCents = %d, unabhängig aggregiert %d", sitzung.ZNr, abschluss.StornierungCents, warenruecknahmen+korrekturen)
		}
		if abschluss.GeldtransitCents != transit {
			t.Errorf("Sitzung %d: GeldtransitCents = %d, unabhängig aggregiert %d", sitzung.ZNr, abschluss.GeldtransitCents, transit)
		}

		start := testJetzt.Add(-sitzung.StartVorJetzt)
		if !abschluss.ZeitraumVon.Equal(start) || !abschluss.ZeitraumBis.Equal(start.Add(sitzung.Dauer)) {
			t.Errorf("Sitzung %d: Zeitraum [%v, %v] entspricht nicht dem Sitzungsfenster", sitzung.ZNr, abschluss.ZeitraumVon, abschluss.ZeitraumBis)
		}
	}
}

// TestBuildSeedDaten_Umbuchungspaar prüft, dass die Umbuchung als verknüpftes
// bestellung-umgebucht-Paar (Abgang/Zugang) mit gemeinsamer UmbuchungID, den
// Standard-Kommentaren und identischen Positionen (Mengen/Preise) erzeugt wird.
func TestBuildSeedDaten_Umbuchungspaar(t *testing.T) {
	_, daten := buildTestDaten(t)

	anzahl := 0
	for i, ev := range daten.Events {
		if ev.event.Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
			continue
		}
		abgang := parseData[kasse.BestellungUmgebuchtV1Data](t, ev.event)
		if !strings.HasPrefix(abgang.Kommentar, "Umbuchung auf Tisch ") {
			continue
		}
		anzahl++

		if i+1 >= len(daten.Events) {
			t.Fatal("Umbuchungs-Abgang ist das letzte Event, Zugang fehlt")
		}
		naechstes := daten.Events[i+1]
		if naechstes.event.Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
			t.Fatalf("auf Umbuchungs-Abgang folgt %s, erwartet bestellung-umgebucht", naechstes.event.Type)
		}
		zugang := parseData[kasse.BestellungUmgebuchtV1Data](t, naechstes.event)
		if !strings.HasPrefix(zugang.Kommentar, "Umbuchung von Tisch ") {
			t.Errorf("Zugangs-Kommentar %q ohne Umbuchungs-Präfix", zugang.Kommentar)
		}
		if zugang.UmbuchungID != abgang.UmbuchungID {
			t.Errorf("Umbuchung: Zugang/Abgang ohne gemeinsame UmbuchungID (%q vs %q)", zugang.UmbuchungID, abgang.UmbuchungID)
		}
		if zugang.GesamtCents != abgang.GesamtCents {
			t.Errorf("Umbuchung: Zugang %d Cent ≠ Abgang %d Cent", zugang.GesamtCents, abgang.GesamtCents)
		}

		// Identische Positionen: gleiche Varianten, Mengen und Einzelpreise.
		type posKey struct{ variante, menge, preis int }
		abgangPositionen := map[posKey]int{}
		for _, p := range abgang.Positionen {
			abgangPositionen[posKey{p.VarianteID, p.Menge, p.EinzelpreisCents}]++
		}
		for _, p := range zugang.Positionen {
			key := posKey{p.VarianteID, p.Menge, p.EinzelpreisCents}
			if abgangPositionen[key] == 0 {
				t.Errorf("Umbuchung: Position %+v des Zugangs fehlt im Abgang", key)
			}
			abgangPositionen[key]--
		}
	}

	if anzahl != 1 {
		t.Errorf("%d Umbuchungen gefunden, erwartet 1", anzahl)
	}
}

// TestBuildSeedDaten_DirektverkaufFesteSubjects prüft, dass alle Direktverkäufe die festen
// Subject-UUIDs aus dem Drehbuch tragen und mindestens ein Storno existiert.
func TestBuildSeedDaten_DirektverkaufFesteSubjects(t *testing.T) {
	_, daten := buildTestDaten(t)

	erwartet := map[string]bool{}
	for nr := 1; nr <= 4; nr++ {
		erwartet[kasse.DirektverkaufSubject(1, dvID(1, nr))] = true
	}
	for nr := 1; nr <= 10; nr++ {
		erwartet[kasse.DirektverkaufSubject(2, dvID(2, nr))] = true
	}
	for nr := 1; nr <= 4; nr++ {
		erwartet[kasse.DirektverkaufSubject(3, dvID(3, nr))] = true
	}

	gesehen := map[string]bool{}
	stornos := 0
	for _, ev := range daten.Events {
		if !strings.Contains(ev.event.Subject, "/direktverkauf-") {
			continue
		}
		if !erwartet[ev.event.Subject] {
			t.Errorf("Direktverkauf-Subject %s nicht im Drehbuch", ev.event.Subject)
		}
		gesehen[ev.event.Subject] = true
		if ev.event.Type == string(kasse.EventTypeDirektverkaufStorniertV1) {
			stornos++
		}
	}

	for subj := range erwartet {
		if !gesehen[subj] {
			t.Errorf("Direktverkauf-Subject %s fehlt", subj)
		}
	}
	if stornos < 1 {
		t.Error("kein Direktverkauf-Storno gefunden, erwartet mindestens 1")
	}
}

// TestBuildSeedDaten_Kassenfuehrung prüft Geldtransit in beide Richtungen sowie den
// Kassensturz: Der Soll-Bestand entspricht einer unabhängigen Aggregation der vorherigen
// Events (Kassenbestand-Formel), die Differenz erzeugt die zugehörige Differenz-Buchung.
func TestBuildSeedDaten_Kassenfuehrung(t *testing.T) {
	_, daten := buildTestDaten(t)
	events := eventsProSitzung(daten)

	richtungen := map[string]bool{}
	for _, tagesEvents := range events {
		for _, evt := range tagesEvents {
			if evt.Type == string(kasse.EventTypeGeldtransitGebuchtV1) {
				richtungen[parseData[kasse.GeldtransitGebuchtV1Data](t, evt).Richtung] = true
			}
		}
	}
	if !richtungen["entnahme"] || !richtungen["einlage"] {
		t.Errorf("Geldtransit-Richtungen %v, erwartet einlage und entnahme", richtungen)
	}

	// Kassensturz Samstag: Soll-Bestand unabhängig aggregieren (Kassenbestand-Formel).
	bestand := 0
	var sturz *kasse.KassensturzDurchgefuehrtV1Data
	var differenz *kasse.DifferenzSollIstGebuchtV1Data
	for _, evt := range events[2] {
		if evt.Type == string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
			data := parseData[kasse.KassensturzDurchgefuehrtV1Data](t, evt)
			sturz = &data
			continue
		}
		if evt.Type == string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
			data := parseData[kasse.DifferenzSollIstGebuchtV1Data](t, evt)
			differenz = &data
			continue
		}
		if sturz != nil {
			continue // nur Events vor dem Kassensturz zählen für den Soll-Bestand
		}
		switch evt.Type {
		case string(kasse.EventTypeKassensitzungEroeffnetV1):
			bestand += parseData[kasse.KassensitzungEroeffnetV1Data](t, evt).BetragCents
		case string(kasse.EventTypeZahlungKassiertV1):
			bestand += parseData[kasse.ZahlungKassiertV1Data](t, evt).GesamtZahlungCents
		case string(kasse.EventTypeStornierungErteiltV1):
			// Kassenwirksame Warenrücknahme: Bar-Rückgabe mindert den Bestand.
			bestand -= parseData[kasse.StornierungErteiltV1Data](t, evt).GesamtStornierungCents
		case string(kasse.EventTypeDirektverkaufGetaetigtV1):
			bestand += parseData[kasse.DirektverkaufGetaetigtV1Data](t, evt).GesamtbetragCents
		case string(kasse.EventTypeDirektverkaufStorniertV1):
			bestand -= parseData[kasse.DirektverkaufStorniertV1Data](t, evt).GesamtStornierungCents
		case string(kasse.EventTypeGeldtransitGebuchtV1):
			data := parseData[kasse.GeldtransitGebuchtV1Data](t, evt)
			if data.Richtung == "einlage" {
				bestand += data.BetragCents
			} else {
				bestand -= data.BetragCents
			}
		}
	}

	if sturz == nil {
		t.Fatal("kein Kassensturz am Samstag gefunden")
	}
	if sturz.SollBestandCents != bestand {
		t.Errorf("Kassensturz Soll = %d, unabhängig aggregiert %d", sturz.SollBestandCents, bestand)
	}
	if sturz.DifferenzCents == 0 {
		t.Error("Kassensturz Samstag ohne Differenz, erwartet Soll/Ist-Differenz")
	}
	if sturz.SollBestandCents-sturz.IstBestandCents != sturz.DifferenzCents {
		t.Errorf("Kassensturz: Soll %d − Ist %d ≠ Differenz %d", sturz.SollBestandCents, sturz.IstBestandCents, sturz.DifferenzCents)
	}
	if differenz == nil {
		t.Fatal("keine Differenz-Buchung am Samstag gefunden")
	}
	if differenz.BetragCents != sturz.DifferenzCents {
		t.Errorf("Differenz-Buchung %d ≠ Kassensturz-Differenz %d", differenz.BetragCents, sturz.DifferenzCents)
	}
}
