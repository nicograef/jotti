//go:build unit

package seed

import (
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// fiskalischeTypen sind die signaturpflichtigen Event-Typen mit ihrem erwarteten processType.
// Die Sitzungseröffnung ist datenabhängig signaturpflichtig (Anfangsbestand > 0) — im
// Demo-Szenario eröffnen alle Sitzungen mit Anfangsbestand, sie zählt daher dazu.
var fiskalischeTypen = map[string]string{
	string(kasse.EventTypeBestellungAufgenommenV1):   tse.ProcessTypeBestellungV1,
	string(kasse.EventTypeZahlungKassiertV1):         tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeStornierungErteiltV1):      tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeBestellungKorrigiertV1):    tse.ProcessTypeBestellungV1,
	string(kasse.EventTypeBestellungUmgebuchtV1):     tse.ProcessTypeBestellungV1,
	string(kasse.EventTypeDirektverkaufGetaetigtV1):  tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeDirektverkaufStorniertV1):  tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeGeldtransitGebuchtV1):      tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeDifferenzSollIstGebuchtV1): tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeTagesabschlussErstelltV1):  tse.ProcessTypeSonstigerVorgang,
	string(kasse.EventTypeKassensitzungEroeffnetV1):  tse.ProcessTypeKassenbelegV1,
}

// buildSignierteDaten baut das volle Szenario und spielt Outbox und Signatur-Worker nach.
func buildSignierteDaten(t *testing.T) (seedDaten, []ausfallFenster, []signaturauftragZeile) {
	t.Helper()
	s := demoSzenario()
	daten, err := buildSeedDaten(s, testJetzt)
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
	fenster := ausfallFensterAus(s, testJetzt)
	auftraege, err := baueSignaturauftraege(daten.Events, fenster)
	if err != nil {
		t.Fatalf("baueSignaturauftraege: %v", err)
	}
	return daten, fenster, auftraege
}

func fensterFuer(fenster []ausfallFenster, zeit time.Time) *ausfallFenster {
	for i, f := range fenster {
		if !zeit.Before(f.von) && zeit.Before(f.bis) {
			return &fenster[i]
		}
	}
	return nil
}

// auftragProEventID indiziert die Aufträge nach Event-ID und prüft dabei die
// Genau-einmal-Eigenschaft (event_id UNIQUE).
func auftragProEventID(t *testing.T, auftraege []signaturauftragZeile) map[int]signaturauftragZeile {
	t.Helper()
	auftragProEvent := map[int]signaturauftragZeile{}
	for _, a := range auftraege {
		if _, doppelt := auftragProEvent[a.EventID]; doppelt {
			t.Errorf("Event %d hat mehr als einen Signaturauftrag", a.EventID)
		}
		auftragProEvent[a.EventID] = a
	}
	return auftragProEvent
}

// TestBaueSignaturauftraege_TypAbdeckung prüft die Paarigkeit von Events und Aufträgen:
// Genau die fiskalischen Event-Typen erhalten genau einen Auftrag mit korrektem processType
// und Beginn = Event-Zeitstempel; nicht-fiskalische Typen erhalten keinen. Jeder fiskalische
// Typ kommt im Szenario vor.
func TestBaueSignaturauftraege_TypAbdeckung(t *testing.T) {
	daten, _, auftraege := buildSignierteDaten(t)
	auftragProEvent := auftragProEventID(t, auftraege)

	auftraegeProTyp := map[string]int{}
	for _, ev := range daten.Events {
		auftrag, hatAuftrag := auftragProEvent[ev.event.ID]
		erwarteterProcessType, fiskalisch := fiskalischeTypen[ev.event.Type]

		if !fiskalisch {
			if hatAuftrag {
				t.Errorf("%s (%s v%d): nicht-fiskalischer Typ hat einen Signaturauftrag", ev.event.Type, ev.event.Subject, ev.event.Version)
			}
			continue
		}

		if !hatAuftrag {
			t.Errorf("%s (%s v%d): fiskalisches Event ohne Signaturauftrag", ev.event.Type, ev.event.Subject, ev.event.Version)
			continue
		}
		if auftrag.ProcessType != erwarteterProcessType {
			t.Errorf("%s: processType = %q, erwartet %q", ev.event.Type, auftrag.ProcessType, erwarteterProcessType)
		}
		if auftrag.ProcessData == "" {
			t.Errorf("%s: Auftrag ohne processData-Snapshot", ev.event.Type)
		}
		if !auftrag.ErstelltAm.Equal(ev.event.Time) {
			t.Errorf("%s: erstellt_am %v ≠ Event-Zeitstempel %v", ev.event.Type, auftrag.ErstelltAm, ev.event.Time)
		}
		auftraegeProTyp[ev.event.Type]++
	}

	if len(auftraege) != len(auftragProEvent) {
		t.Errorf("%d Aufträge, aber %d Event-IDs — Aufträge ohne Event?", len(auftraege), len(auftragProEvent))
	}
	for typ := range fiskalischeTypen {
		if auftraegeProTyp[typ] == 0 {
			t.Errorf("Event-Typ %s: kein Signaturauftrag im Szenario", typ)
		}
	}
}

// TestBaueSignaturauftraege_MonotonieUndFormat prüft die Signaturen: global streng monotone,
// lückenlose Transaktionsnummern und streng monotone Signaturzähler in Quittier-Reihenfolge,
// die feste Seriennummer, das V0-QR-Format und plausible logTime-Paare. Erledigt und
// Signatur bedingen einander.
func TestBaueSignaturauftraege_MonotonieUndFormat(t *testing.T) {
	_, _, auftraege := buildSignierteDaten(t)

	vergeben := map[int]bool{}
	vorherTx, vorherSig := 0, 0
	for _, a := range auftraege {
		if (a.Status == tse.StatusErledigt) != (a.Signatur != nil) {
			t.Errorf("Auftrag %s: Status %q und Signatur %v widersprechen sich", a.TxID, a.Status, a.Signatur != nil)
		}
		if a.Signatur == nil {
			continue
		}

		sig := a.Signatur
		if sig.TSESeriennummer != fakeTSESeriennummer {
			t.Errorf("Auftrag %s: Seriennummer %q, erwartet %q", a.TxID, sig.TSESeriennummer, fakeTSESeriennummer)
		}
		if !strings.HasPrefix(sig.QRCodeData, "V0;") {
			t.Errorf("Auftrag %s: QR-Code-Daten ohne V0-Präfix: %q", a.TxID, sig.QRCodeData)
		}
		if !sig.LogTimeEnd.After(sig.LogTimeStart) {
			t.Errorf("Auftrag %s: logTimeEnd %v nicht nach logTimeStart %v", a.TxID, sig.LogTimeEnd, sig.LogTimeStart)
		}
		if a.ErledigtAm == nil || a.ErledigtAm.Before(sig.LogTimeEnd) {
			t.Errorf("Auftrag %s: erledigt_am fehlt oder liegt vor logTimeEnd", a.TxID)
		}
		if sig.TransaktionNummer <= vorherTx {
			t.Errorf("Auftrag %s: Transaktionsnummer %d nicht streng monoton nach %d", a.TxID, sig.TransaktionNummer, vorherTx)
		}
		if sig.SignaturZaehler <= vorherSig {
			t.Errorf("Auftrag %s: Signaturzähler %d nicht streng monoton nach %d", a.TxID, sig.SignaturZaehler, vorherSig)
		}
		if vergeben[sig.TransaktionNummer] {
			t.Errorf("Auftrag %s: Transaktionsnummer %d doppelt vergeben", a.TxID, sig.TransaktionNummer)
		}
		vergeben[sig.TransaktionNummer] = true
		vorherTx, vorherSig = sig.TransaktionNummer, sig.SignaturZaehler
	}

	for nr := 1; nr <= len(vergeben); nr++ {
		if !vergeben[nr] {
			t.Errorf("Transaktionsnummer %d fehlt (Lücke)", nr)
		}
	}
}

// TestBaueSignaturauftraege_Ausfallfenster prüft die Dramaturgie der Ausfallfenster: Events
// außerhalb der Fenster werden prompt quittiert (logTime = Event-Zeit), Events in aufgelösten
// Fenstern verspätet nach Fensterende (mit Fehlversuchen samt Fenster-Grund), einzelne
// scheitern dauerhaft (genau einer verworfen), Events im offenen Fenster bleiben offen ohne
// Signatur. Die Statusverteilung stimmt (überwiegend erledigt).
func TestBaueSignaturauftraege_Ausfallfenster(t *testing.T) {
	daten, fenster, auftraege := buildSignierteDaten(t)
	auftragProEvent := auftragProEventID(t, auftraege)

	eventZeit := map[int]time.Time{}
	for _, ev := range daten.Events {
		eventZeit[ev.event.ID] = ev.event.Time
	}

	statusZahl := map[string]int{}
	for _, a := range auftraege {
		statusZahl[a.Status]++
		f := fensterFuer(fenster, a.ErstelltAm)

		if f == nil {
			if a.Status != tse.StatusErledigt {
				t.Errorf("Auftrag %s außerhalb der Fenster hat Status %q, erwartet erledigt", a.TxID, a.Status)
				continue
			}
			if a.Versuche != 0 || a.LetzterFehler != nil {
				t.Errorf("Auftrag %s außerhalb der Fenster: erwartet 0 Versuche und keinen Fehler", a.TxID)
			}
			if !a.Signatur.LogTimeStart.Equal(a.ErstelltAm) {
				t.Errorf("Auftrag %s: prompte Signatur mit logTimeStart %v ≠ Event-Zeit %v", a.TxID, a.Signatur.LogTimeStart, a.ErstelltAm)
			}
			continue
		}

		switch a.Status {
		case tse.StatusOffen:
			if f.aufgeloest {
				t.Errorf("offener Auftrag %s in aufgelöstem Fenster", a.TxID)
			}
			if a.Versuche != 0 || a.LetzterFehler != nil || a.ErledigtAm != nil || a.Signatur != nil {
				t.Errorf("offener Auftrag %s: erwartet 0 Versuche, kein Fehler, keine Erledigung, keine Signatur", a.TxID)
			}
		case tse.StatusErledigt:
			if !f.aufgeloest {
				t.Errorf("erledigter Auftrag %s in offenem Fenster", a.TxID)
			}
			if a.Signatur.LogTimeStart.Before(f.bis) {
				t.Errorf("Auftrag %s: nachsignierte Signatur mit logTimeStart %v vor dem Fensterende %v", a.TxID, a.Signatur.LogTimeStart, f.bis)
			}
			if a.Versuche > 0 && (a.LetzterFehler == nil || *a.LetzterFehler != f.grund) {
				t.Errorf("Auftrag %s: Fehlversuche ohne Fenster-Grund als letzten Fehler", a.TxID)
			}
		case tse.StatusFehlgeschlagen, tse.StatusVerworfen:
			if !f.aufgeloest {
				t.Errorf("%s Auftrag %s in offenem Fenster", a.Status, a.TxID)
			}
			if a.Versuche != tse_repo.MaxSignaturVersuche {
				t.Errorf("%s Auftrag %s: %d Versuche, erwartet das Maximum %d", a.Status, a.TxID, a.Versuche, tse_repo.MaxSignaturVersuche)
			}
			if a.LetzterFehler == nil || *a.LetzterFehler == "" {
				t.Errorf("%s Auftrag %s ohne Fehlertext", a.Status, a.TxID)
			}
			if a.ErledigtAm != nil || a.Signatur != nil {
				t.Errorf("%s Auftrag %s: erwartet keine Erledigung und keine Signatur", a.Status, a.TxID)
			}
		default:
			t.Errorf("Auftrag %s: unerwarteter Status %q", a.TxID, a.Status)
		}
	}

	// Jedes fiskalische Event in einem offenen Fenster muss als offener Auftrag erscheinen.
	for eventID, a := range auftragProEvent {
		f := fensterFuer(fenster, eventZeit[eventID])
		if f != nil && !f.aufgeloest && a.Status != tse.StatusOffen {
			t.Errorf("Auftrag %s im offenen Fenster hat Status %q", a.TxID, a.Status)
		}
	}

	if statusZahl[tse.StatusOffen] < 1 {
		t.Error("kein offener Signaturauftrag (Sonntags-Aussetzer fehlt)")
	}
	if statusZahl[tse.StatusFehlgeschlagen] < 2 {
		t.Errorf("nur %d fehlgeschlagene Aufträge, erwartet einzelne (≥ 2)", statusZahl[tse.StatusFehlgeschlagen])
	}
	if statusZahl[tse.StatusVerworfen] != 1 {
		t.Errorf("%d verworfene Aufträge, erwartet genau 1", statusZahl[tse.StatusVerworfen])
	}
	rest := statusZahl[tse.StatusOffen] + statusZahl[tse.StatusFehlgeschlagen] + statusZahl[tse.StatusVerworfen]
	if statusZahl[tse.StatusErledigt] <= rest {
		t.Errorf("erledigte Aufträge (%d) nicht in der Überzahl gegenüber den übrigen (%d)", statusZahl[tse.StatusErledigt], rest)
	}
}
