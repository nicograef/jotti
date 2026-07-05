//go:build unit

package application

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/fiskal/dsfinvk"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

const konsistenzZNr = 3

func konsistenzPosition(satz string, einzelpreis, menge int) kasse.PositionEventData {
	return kasse.PositionEventData{
		PositionID:  "p-" + satz + "-" + strconv.Itoa(einzelpreis),
		VarianteID:  101,
		ProduktName: "Produkt",
		Kategorie:   "essen",
		Steuersatz:  satz,
		Einzelpreis: einzelpreis,
		Menge:       menge,
	}
}

func konsistenzEvent(t *testing.T, id int, typ kasse.EventType, subject string, data any) event.Event {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal %s: %v", typ, err)
	}
	return event.Event{
		ID:       id,
		UserID:   7,
		UserName: "anna",
		Type:     string(typ),
		Time:     time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC).Add(time.Duration(id) * time.Minute),
		Subject:  subject,
		Version:  id,
		Data:     raw,
	}
}

// centsAusExportBetrag parst einen DSFinV-K-Betrag ("27,98", "-14,55") in Cents.
func centsAusExportBetrag(t *testing.T, betrag string) int {
	t.Helper()
	negativ := strings.HasPrefix(betrag, "-")
	teile := strings.Split(strings.TrimPrefix(betrag, "-"), ",")
	if len(teile) != 2 || len(teile[1]) != 2 {
		t.Fatalf("unerwartetes Betragsformat im Export: %q", betrag)
	}
	euro, err := strconv.Atoi(teile[0])
	if err != nil {
		t.Fatalf("unerwartetes Betragsformat im Export: %q", betrag)
	}
	cent, err := strconv.Atoi(teile[1])
	if err != nil {
		t.Fatalf("unerwartetes Betragsformat im Export: %q", betrag)
	}
	cents := euro*100 + cent
	if negativ {
		cents = -cents
	}
	return cents
}

// Gemeinsamer Testfall für die B9-Invariante: Die USt-Aufschlüsselung des
// Reportings (berechneUmsatzProSteuersatz auf den Brutto-Positionszeilen)
// muss für dieselbe Sitzung exakt die Summen der businesscases.csv des
// DSFinV-K-Exports ergeben — mit Kombi-Positionen, Warenrücknahme,
// Teilzahlungen und Direktverkauf inkl. Storno. Die krummen Beträge sind
// bewusst gewählt, damit Zeilen- und Aggregatbasis unterschiedlich runden.
func TestUmsatzProSteuersatz_KonsistentMitDSFinVKBusinesscases(t *testing.T) {
	subjectTisch := kasse.TischSessionSubject(konsistenzZNr, 42)

	zahlung1 := kasse.ZahlungKassiertV1Data{
		ZahlungID:          "z1",
		GesamtZahlungCents: 2238,
		Positionen: []kasse.PositionEventData{
			konsistenzPosition("kombi", 1005, 1),
			konsistenzPosition("regel", 450, 2),
			konsistenzPosition("befreit", 333, 1),
		},
	}
	zahlung2 := kasse.ZahlungKassiertV1Data{
		ZahlungID:          "z2",
		GesamtZahlungCents: 1470,
		Positionen: []kasse.PositionEventData{
			konsistenzPosition("kombi", 1005, 1),
			konsistenzPosition("ermaessigt", 155, 3),
		},
	}
	warenruecknahme := kasse.StornierungErteiltV1Data{
		StornierungID:          "s1",
		ZahlungID:              "z1",
		GesamtStornierungCents: 1455,
		Kommentar:              "Warenruecknahme",
		Positionen: []kasse.PositionEventData{
			konsistenzPosition("kombi", 1005, 1),
			konsistenzPosition("regel", 450, 1),
		},
	}
	direktverkauf := kasse.DirektverkaufGetaetigtV1Data{
		VerkaufID:         "d1",
		GesamtbetragCents: 880,
		Positionen: []kasse.PositionEventData{
			konsistenzPosition("kombi", 335, 2),
			konsistenzPosition("ermaessigt", 210, 1),
		},
	}
	direktverkaufStorno := kasse.DirektverkaufStorniertV1Data{
		StornierungID:          "ds1",
		VerkaufID:              "d1",
		GesamtStornierungCents: 335,
		Kommentar:              "Fehlbuchung",
		Positionen: []kasse.PositionEventData{
			konsistenzPosition("kombi", 335, 1),
		},
	}

	events := []event.Event{
		konsistenzEvent(t, 1, kasse.EventTypeZahlungKassiertV1, subjectTisch, zahlung1),
		konsistenzEvent(t, 2, kasse.EventTypeZahlungKassiertV1, subjectTisch, zahlung2),
		konsistenzEvent(t, 3, kasse.EventTypeStornierungErteiltV1, subjectTisch, warenruecknahme),
		konsistenzEvent(t, 4, kasse.EventTypeDirektverkaufGetaetigtV1, kasse.DirektverkaufSubject(konsistenzZNr, "d1"), direktverkauf),
		konsistenzEvent(t, 5, kasse.EventTypeDirektverkaufStorniertV1, kasse.DirektverkaufSubject(konsistenzZNr, "d1"), direktverkaufStorno),
	}

	// Reporting-Seite: Brutto-Positionszeilen wie in GetUmsatzPositionszeilen
	// (eine Zeile je Position, Stornos negativ), dann die Aufschlüsselung.
	var zeilen []reporting.UmsatzSteuersatz
	addZeilen := func(positionen []kasse.PositionEventData, vorzeichen int) {
		for _, p := range positionen {
			zeilen = append(zeilen, reporting.UmsatzSteuersatz{
				Satz:        steuer.Steuersatz(p.Steuersatz),
				BruttoCents: vorzeichen * p.Einzelpreis * p.Menge,
			})
		}
	}
	addZeilen(zahlung1.Positionen, 1)
	addZeilen(zahlung2.Positionen, 1)
	addZeilen(warenruecknahme.Positionen, -1)
	addZeilen(direktverkauf.Positionen, 1)
	addZeilen(direktverkaufStorno.Positionen, -1)

	aufschluesselung := berechneUmsatzProSteuersatz(zeilen)

	// Export-Seite: businesscases.csv derselben Events, summiert je UST_SCHLUESSEL.
	snapshot := dsfinvk.Snapshot{
		KasseSeriennummer: "11111111-2222-3333-4444-555555555555",
		Erstellung:        time.Date(2026, 6, 16, 20, 0, 0, 0, time.UTC),
		KassensitzungNr:   konsistenzZNr,
		Betreiber:         betreiber.Betreiber{Vereinsname: "TSV Beispiel", Strasse: "Hauptstr. 1", Plz: "12345", Ort: "Musterdorf"},
		Tischnamen:        map[int]string{42: "Tisch 42"},
	}
	archive, err := dsfinvk.Map(snapshot, events, nil)
	if err != nil {
		t.Fatalf("dsfinvk.Map() error = %v", err)
	}

	var businesscases dsfinvk.Table
	for _, table := range archive.Tables() {
		if table.File == "businesscases.csv" {
			businesscases = table
		}
	}
	if businesscases.File == "" {
		t.Fatal("businesscases.csv fehlt im Archiv")
	}

	type betrag struct{ brutto, netto, ust int }
	exportSummen := map[string]betrag{}
	for _, record := range businesscases.Records {
		schluessel := record[6]
		summe := exportSummen[schluessel]
		summe.brutto += centsAusExportBetrag(t, record[7])
		summe.netto += centsAusExportBetrag(t, record[8])
		summe.ust += centsAusExportBetrag(t, record[9])
		exportSummen[schluessel] = summe
	}

	// UST_SCHLUESSEL laut DSFinV-K Anlage 2: 1 = regel, 2 = ermaessigt, 6 = befreit.
	schluesselFuerSatz := map[steuer.Steuersatz]string{
		steuer.RegelSteuersatz:      "1",
		steuer.ErmaessigtSteuersatz: "2",
		steuer.BefreitSteuersatz:    "6",
	}

	if len(aufschluesselung) != len(exportSummen) {
		t.Fatalf("Reporting hat %d Steuersätze, Export %d: %+v vs %+v", len(aufschluesselung), len(exportSummen), aufschluesselung, exportSummen)
	}

	gesamtBrutto := 0
	for _, eintrag := range aufschluesselung {
		gesamtBrutto += eintrag.BruttoCents
		if eintrag.NettoCents+eintrag.SteuerCents != eintrag.BruttoCents {
			t.Errorf("%s: Netto %d + Steuer %d != Brutto %d", eintrag.Satz, eintrag.NettoCents, eintrag.SteuerCents, eintrag.BruttoCents)
		}
		export, ok := exportSummen[schluesselFuerSatz[eintrag.Satz]]
		if !ok {
			t.Errorf("%s: fehlt in businesscases.csv", eintrag.Satz)
			continue
		}
		if eintrag.BruttoCents != export.brutto || eintrag.NettoCents != export.netto || eintrag.SteuerCents != export.ust {
			t.Errorf("%s: Reporting {%d %d %d} != businesscases.csv {%d %d %d}",
				eintrag.Satz, eintrag.BruttoCents, eintrag.NettoCents, eintrag.SteuerCents,
				export.brutto, export.netto, export.ust)
		}
	}

	// Σ(Brutto je Steuersatz) == Gesamtumsatz der Sitzung
	// (2238 + 1470 - 1455 + 880 - 335).
	if gesamtBrutto != 2798 {
		t.Errorf("Σ Brutto je Steuersatz = %d, want 2798", gesamtBrutto)
	}
}
