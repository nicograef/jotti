//go:build unit

package seed

import (
	"encoding/base64"
	"strconv"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/druckstation"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// buildDruckDaten baut das volle Szenario, spielt die Fake-TSE nach und erzeugt die
// Druckauftrags-Historie.
func buildDruckDaten(t *testing.T) (szenario, seedDaten, map[int]*tse.Signatur, []druckauftragZeile) {
	t.Helper()
	s := demoSzenario()
	daten, err := buildSeedDaten(s, testJetzt)
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
	signaturauftraege, err := buildSignaturauftraege(daten.Events, ausfallFensterAus(s, testJetzt))
	if err != nil {
		t.Fatalf("buildSignaturauftraege: %v", err)
	}
	signaturen := signaturenNachEventID(signaturauftraege)
	auftraege, err := buildDruckauftraege(s, daten.Events, signaturen, testJetzt)
	if err != nil {
		t.Fatalf("buildDruckauftraege: %v", err)
	}
	return s, daten, signaturen, auftraege
}

// TestDemoSzenario_Druckstationen prüft die Drehbuch-Konfiguration: alle fünf Stationen mit
// gültiger IPv4-Adresse, Kassenbeleg ohne Bonmodus, die übrigen mit gemischten Bonmodi.
func TestDemoSzenario_Druckstationen(t *testing.T) {
	s := demoSzenario()

	if len(s.Druckstationen) != 5 {
		t.Fatalf("%d Druckstationen, erwartet 5", len(s.Druckstationen))
	}

	bonmodi := map[druckstation.Bonmodus]int{}
	for _, st := range s.Druckstationen {
		d := druckstation.Druckstation{Kategorie: st.Kategorie, DruckerIP: st.DruckerIP, Bonmodus: st.Bonmodus}
		if err := d.Validate(); err != nil {
			t.Errorf("Druckstation %s: %v", st.Kategorie, err)
		}
		if st.DruckerIP == "" {
			t.Errorf("Druckstation %s ohne Drucker-IP", st.Kategorie)
		}
		if st.Kategorie.HatBonmodus() {
			bonmodi[st.Bonmodus]++
		}
	}

	if bonmodi[druckstation.BonmodusProPosition] == 0 || bonmodi[druckstation.BonmodusProBestellung] == 0 {
		t.Errorf("Bonmodi nicht gemischt: %v", bonmodi)
	}
}

// TestBaueDruckauftraege_StatusVerteilung prüft die Status-Dramaturgie: alle vier Status in
// beiden Bon-Arten plausibel befüllt — überwiegend gedruckt (Quittung nach der Erstellung),
// offene nur im Relay-Abholfenster, fehlgeschlagene mit ausgeschöpften Versuchen und
// Fehlertext in einem Drucker-Ausfallfenster, genau einer verworfen.
func TestBaueDruckauftraege_StatusVerteilung(t *testing.T) {
	s, _, _, auftraege := buildDruckDaten(t)
	fenster := druckerFensterAus(s, testJetzt)

	statusZahl := map[string]int{}
	bonArten := map[string]int{}
	fehlgeschlageneBonArten := map[string]int{}
	for _, a := range auftraege {
		statusZahl[a.Status]++
		bonArten[a.BonArt]++

		if a.ErstelltAm.After(testJetzt) {
			t.Errorf("Auftrag %s: erstellt_am %v liegt nach jetzt", a.Referenz, a.ErstelltAm)
		}

		switch a.Status {
		case "offen":
			if a.Versuche != 0 || a.LetzterFehler != nil || a.GedrucktAm != nil {
				t.Errorf("offener Auftrag %s: erwartet 0 Versuche, kein Fehler, kein gedruckt_am", a.Referenz)
			}
			if !a.ErstelltAm.After(testJetzt.Add(-relayAbholFenster)) {
				t.Errorf("offener Auftrag %s: erstellt_am %v außerhalb des Relay-Abholfensters", a.Referenz, a.ErstelltAm)
			}
		case "gedruckt":
			if a.GedrucktAm == nil || !a.GedrucktAm.After(a.ErstelltAm) {
				t.Errorf("gedruckter Auftrag %s: gedruckt_am fehlt oder liegt nicht nach erstellt_am", a.Referenz)
			}
		case "fehlgeschlagen", "verworfen":
			fehlgeschlageneBonArten[a.BonArt]++
			if a.Versuche != druckauftrag_repo.MaxDruckversuche {
				t.Errorf("%s Auftrag %s: %d Versuche, erwartet %d", a.Status, a.Referenz, a.Versuche, druckauftrag_repo.MaxDruckversuche)
			}
			if a.LetzterFehler == nil || *a.LetzterFehler == "" {
				t.Errorf("%s Auftrag %s ohne Fehlertext", a.Status, a.Referenz)
			}
			if a.GedrucktAm != nil {
				t.Errorf("%s Auftrag %s mit gedruckt_am", a.Status, a.Referenz)
			}
			imFenster := false
			for _, f := range fenster {
				if a.ZielIP == f.ip && !a.ErstelltAm.Before(f.von) && a.ErstelltAm.Before(f.bis) {
					imFenster = true
				}
			}
			if !imFenster {
				t.Errorf("%s Auftrag %s liegt in keinem Drucker-Ausfallfenster", a.Status, a.Referenz)
			}
		default:
			t.Errorf("Auftrag %s: unbekannter Status %q", a.Referenz, a.Status)
		}
	}

	if statusZahl["offen"] < 1 {
		t.Error("kein offener Druckauftrag")
	}
	if statusZahl["fehlgeschlagen"] < 2 {
		t.Errorf("nur %d fehlgeschlagene Druckaufträge, erwartet mehrere (≥ 2)", statusZahl["fehlgeschlagen"])
	}
	if statusZahl["verworfen"] != 1 {
		t.Errorf("%d verworfene Druckaufträge, erwartet genau 1", statusZahl["verworfen"])
	}
	rest := statusZahl["offen"] + statusZahl["fehlgeschlagen"] + statusZahl["verworfen"]
	if statusZahl["gedruckt"] <= rest {
		t.Errorf("gedruckte Aufträge (%d) nicht in der Überzahl gegenüber den übrigen (%d)", statusZahl["gedruckt"], rest)
	}
	if bonArten["arbeitsbon"] == 0 || bonArten["kassenbeleg"] == 0 {
		t.Errorf("nicht beide Bon-Arten vorhanden: %v", bonArten)
	}
	if fehlgeschlageneBonArten["arbeitsbon"] == 0 || fehlgeschlageneBonArten["kassenbeleg"] == 0 {
		t.Errorf("fehlgeschlagene Aufträge nicht in beiden Bon-Arten: %v", fehlgeschlageneBonArten)
	}
}

// TestBaueDruckauftraege_ReferenzenUndPayloads prüft die fachliche Konsistenz: Jede Referenz
// verweist auf das Event mit passendem Typ und passender ID, jede Bestellung und jeder
// Direktverkauf hat Arbeits- bzw. Abholbons, und die Payloads sind echte ESC/POS-Bytes —
// Kassenbelege inklusive Betreiber, Gesamtsumme und TSE-QR-Daten aus den Signaturspalten
// des Auftrags. Vorgänge ohne quittierte Signatur erhalten keinen Kassenbeleg-Druckauftrag.
func TestBaueDruckauftraege_ReferenzenUndPayloads(t *testing.T) {
	s, daten, signaturen, auftraege := buildDruckDaten(t)

	abholbonIP := ""
	for _, st := range s.Druckstationen {
		if st.Kategorie == druckstation.KategorieAbholbon {
			abholbonIP = st.DruckerIP
		}
	}

	referenzen := map[string]int{}
	for _, a := range auftraege {
		typ, idText, ok := strings.Cut(a.Referenz, ":")
		if !ok {
			t.Errorf("Auftrag mit unerwarteter Referenz %q", a.Referenz)
			continue
		}
		id, err := strconv.Atoi(idText)
		if err != nil || id < 1 || id > len(daten.Events) {
			t.Errorf("Referenz %q verweist auf keine gültige Event-ID", a.Referenz)
			continue
		}
		evt := daten.Events[id-1].event
		if evt.ID != id {
			t.Fatalf("Event-IDs nicht deterministisch ab 1 vergeben: Position %d trägt ID %d", id, evt.ID)
		}
		if evt.Type != typ+":v1" {
			t.Errorf("Referenz %q verweist auf Event vom Typ %s", a.Referenz, evt.Type)
		}
		referenzen[a.BonArt+"/"+a.Referenz]++

		payload, err := base64.StdEncoding.DecodeString(a.Payload)
		if err != nil || len(payload) == 0 {
			t.Errorf("Auftrag %s: Payload kein gültiges Base64 oder leer", a.Referenz)
			continue
		}

		if a.BonArt == "arbeitsbon" && a.ZielIP == abholbonIP && !strings.Contains(string(payload), "Direktverkauf") {
			t.Errorf("Abholbon %s ohne Direktverkauf-Kopfzeile", a.Referenz)
		}
	}

	belege := 0
	belegeNachsigniert := 0
	fenster := ausfallFensterAus(s, testJetzt)
	for _, ev := range daten.Events {
		evt := ev.event
		switch kasse.EventType(evt.Type) {
		case kasse.EventTypeBestellungAufgenommenV1:
			if referenzen["arbeitsbon/bestellung-aufgenommen:"+strconv.Itoa(evt.ID)] == 0 {
				t.Errorf("Bestellung %d ohne Arbeitsbon-Auftrag", evt.ID)
			}
		case kasse.EventTypeDirektverkaufGetaetigtV1:
			if referenzen["arbeitsbon/direktverkauf-getaetigt:"+strconv.Itoa(evt.ID)] != 1 {
				t.Errorf("Direktverkauf %d ohne Sammel-Abholbon", evt.ID)
			}
		}

		beleg, ok := findKassenbeleg(auftraege, evt)
		if !ok {
			continue
		}
		belege++

		signatur := signaturen[evt.ID]
		if signatur == nil {
			t.Errorf("Kassenbeleg %s zu einem Vorgang ohne quittierte Signatur", beleg.Referenz)
			continue
		}
		if fensterFuer(fenster, evt.Time) != nil {
			belegeNachsigniert++
			if beleg.ErstelltAm.Before(signatur.LogTimeEnd) {
				t.Errorf("Kassenbeleg %s: erstellt_am %v vor der nachgetragenen Signatur %v", beleg.Referenz, beleg.ErstelltAm, signatur.LogTimeEnd)
			}
		}

		payload, err := base64.StdEncoding.DecodeString(beleg.Payload)
		if err != nil {
			t.Fatalf("Kassenbeleg %s: Payload kein gültiges Base64", beleg.Referenz)
		}
		text := string(payload)
		for _, erwartet := range []string{"KASSENBELEG", s.Betreiber.Vereinsname, "GESAMT:", "Kassen-ID: " + fakeKassenSeriennummer} {
			if !strings.Contains(text, erwartet) {
				t.Errorf("Kassenbeleg %s: Payload ohne %q", beleg.Referenz, erwartet)
			}
		}
		if !strings.Contains(text, signatur.QRCodeData) {
			t.Errorf("Kassenbeleg %s: Payload ohne TSE-QR-Daten des Auftrags", beleg.Referenz)
		}
	}

	if belege == 0 {
		t.Fatal("kein Kassenbeleg-Druckauftrag im Szenario")
	}
	if belegeNachsigniert == 0 {
		t.Error("kein Kassenbeleg mit nachgetragener Signatur (Beleg zum aufgelösten TSE-Ausfallfenster fehlt)")
	}
}

// TestKassenbeleg_OhneSignaturKeinDruckauftrag prüft den Ausstehend-Fall: Zu einem Vorgang
// ohne quittierte Signatur entsteht kein Kassenbeleg-Druckauftrag — der Beleg-Abruf hätte
// „ausstehend" geantwortet.
func TestKassenbeleg_OhneSignaturKeinDruckauftrag(t *testing.T) {
	_, daten, signaturen, _ := buildDruckDaten(t)

	var unsigniert *e.Event
	for i := range daten.Events {
		evt := daten.Events[i].event
		typ := kasse.EventType(evt.Type)
		if (typ == kasse.EventTypeZahlungKassiertV1 || typ == kasse.EventTypeDirektverkaufGetaetigtV1) && signaturen[evt.ID] == nil {
			unsigniert = &evt
			break
		}
	}
	if unsigniert == nil {
		t.Fatal("kein unsignierter Zahlungs- oder Direktverkaufs-Vorgang im Szenario")
	}

	b := &bondruckBuilder{signaturen: signaturen}
	if _, ok, err := b.kassenbeleg(*unsigniert); err != nil {
		t.Fatalf("kassenbeleg: %v", err)
	} else if ok {
		t.Error("kassenbeleg lieferte einen Druckauftrag trotz fehlender Signatur")
	}
}

// findKassenbeleg sucht den Kassenbeleg-Druckauftrag zu einem Event.
func findKassenbeleg(auftraege []druckauftragZeile, evt e.Event) (druckauftragZeile, bool) {
	typ := strings.TrimSuffix(evt.Type, ":v1")
	for _, a := range auftraege {
		if a.BonArt == "kassenbeleg" && a.Referenz == typ+":"+strconv.Itoa(evt.ID) {
			return a, true
		}
	}
	return druckauftragZeile{}, false
}
