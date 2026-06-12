//go:build unit

package seed

import (
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

func TestBuildSeedDaten_VersionenLueckenlosJeSubject(t *testing.T) {
	daten, err := buildSeedDaten(phase1Szenario(), time.Now().UTC())
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}

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

func TestBuildSeedDaten_ZeitstempelMonotonImFenster(t *testing.T) {
	jetzt := time.Now().UTC()
	daten, err := buildSeedDaten(phase1Szenario(), jetzt)
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
	if len(daten.Events) == 0 {
		t.Fatal("keine Events erzeugt")
	}

	fensterStart := jetzt.Add(-5 * time.Hour)
	var vorher time.Time
	for i, ev := range daten.Events {
		ts := ev.event.Time
		if ts.Before(fensterStart) || ts.After(jetzt) {
			t.Errorf("Event %d Zeitstempel %v außerhalb des Fensters [%v, %v]", i, ts, fensterStart, jetzt)
		}
		if i > 0 && !ts.After(vorher) {
			t.Errorf("Event %d Zeitstempel %v nicht streng monoton nach %v", i, ts, vorher)
		}
		vorher = ts
	}
}

func TestBuildSeedDaten_EventsValide(t *testing.T) {
	daten, err := buildSeedDaten(phase1Szenario(), time.Now().UTC())
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
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

func TestBuildSeedDaten_TischZustaendeKonsistent(t *testing.T) {
	s := phase1Szenario()
	daten, err := buildSeedDaten(s, time.Now().UTC())
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}

	// Tisch-Subjects einzeln nachspielen.
	states := map[string]kasse.TischSession{}
	eventsProSubject := map[string]int{}
	for _, ev := range daten.Events {
		subj := ev.event.Subject
		if !strings.Contains(subj, "/tisch-") {
			continue
		}
		eventsProSubject[subj]++
		st, err := kasse.ApplyEvent(states[subj], ev.event)
		if err != nil {
			t.Fatalf("ApplyEvent %s v%d: %v", subj, ev.event.Version, err)
		}
		states[subj] = st
	}

	znr := s.Sitzung.ZNr

	// Tisch 1: bestellt, ausgegeben und bezahlt → drei Events, Saldo 0, keine offenen Positionen.
	t1Subject := kasse.TischSessionSubject(znr, 1)
	if got := eventsProSubject[t1Subject]; got != 3 {
		t.Errorf("Tisch 1: %d Events, erwartet 3 (Bestellung+Ausgabe+Zahlung)", got)
	}
	t1 := states[t1Subject]
	if t1.SaldoCents != 0 {
		t.Errorf("Tisch 1 Saldo = %d, erwartet 0", t1.SaldoCents)
	}
	if len(t1.UnbezahltePositionen) != 0 {
		t.Errorf("Tisch 1: %d unbezahlte Positionen, erwartet 0", len(t1.UnbezahltePositionen))
	}
	if len(t1.AusstehendePositionen) != 0 {
		t.Errorf("Tisch 1: %d ausstehende Positionen, erwartet 0", len(t1.AusstehendePositionen))
	}

	// Tisch 2: ausgegeben, aber unbezahlt → Saldo > 0, keine ausstehenden, aber unbezahlte Positionen.
	t2 := states[kasse.TischSessionSubject(znr, 2)]
	if t2.SaldoCents <= 0 {
		t.Errorf("Tisch 2 Saldo = %d, erwartet > 0", t2.SaldoCents)
	}
	if len(t2.AusstehendePositionen) != 0 {
		t.Errorf("Tisch 2: %d ausstehende Positionen, erwartet 0 (ausgegeben)", len(t2.AusstehendePositionen))
	}
	if len(t2.UnbezahltePositionen) == 0 {
		t.Error("Tisch 2: keine unbezahlten Positionen, erwartet > 0")
	}

	// Tisch 3: nur bestellt → ausstehende Positionen vorhanden, Saldo > 0.
	t3 := states[kasse.TischSessionSubject(znr, 3)]
	if len(t3.AusstehendePositionen) == 0 {
		t.Error("Tisch 3: keine ausstehenden Positionen, erwartet > 0 (nicht ausgegeben)")
	}
	if t3.SaldoCents <= 0 {
		t.Errorf("Tisch 3 Saldo = %d, erwartet > 0", t3.SaldoCents)
	}
}
