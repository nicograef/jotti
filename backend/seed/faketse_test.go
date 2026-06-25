//go:build unit

package seed

import (
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// fiskalischeTypen sind die zu signierenden Event-Typen mit ihrem erwarteten processType.
var fiskalischeTypen = map[string]string{
	string(kasse.EventTypeBestellungAufgenommenV1):   tse.ProcessTypeBestellungV1,
	string(kasse.EventTypeZahlungKassiertV1):         tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeStornierungErteiltV1):      tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeBestellungUmgebuchtV1):     tse.ProcessTypeBestellungV1,
	string(kasse.EventTypeAuszahlungGeleistetV1):     tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeDirektverkaufGetaetigtV1):  tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeDirektverkaufStorniertV1):  tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeGeldtransitGebuchtV1):      tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeDifferenzSollIstGebuchtV1): tse.ProcessTypeKassenbelegV1,
	string(kasse.EventTypeTagesabschlussErstelltV1):  tse.ProcessTypeSonstigerVorgang,
}

// tseFelder liest die TSE-Felder aus beliebigen Event-Daten (alle Event-Typen nutzen
// dieselben JSON-Keys).
type tseFelder struct {
	TxID    string         `json:"tseTxId"`
	Data    *kasse.TSEData `json:"tseData"`
	Ausfall bool           `json:"tseAusfall"`
}

// buildSignierteDaten baut das volle Szenario und signiert es mit der Fake-TSE.
func buildSignierteDaten(t *testing.T) (seedDaten, []ausfallFenster, tseSeitentabellen) {
	t.Helper()
	s := demoSzenario()
	daten, err := buildSeedDaten(s, testJetzt)
	if err != nil {
		t.Fatalf("buildSeedDaten: %v", err)
	}
	fenster := ausfallFensterAus(s, testJetzt)
	seiten, err := signiereEvents(daten.Events, fenster)
	if err != nil {
		t.Fatalf("signiereEvents: %v", err)
	}
	return daten, fenster, seiten
}

func fensterFuer(fenster []ausfallFenster, zeit time.Time) *ausfallFenster {
	for i, f := range fenster {
		if !zeit.Before(f.von) && zeit.Before(f.bis) {
			return &fenster[i]
		}
	}
	return nil
}

// TestSigniereEvents_TypAbdeckung prüft, dass genau die fiskalischen Event-Typen TSE-Felder
// tragen — außerhalb der Ausfallfenster mit vollständigen TSE-Daten samt korrektem
// processType, innerhalb nur mit der txID. Jeder fiskalische Typ kommt signiert vor.
func TestSigniereEvents_TypAbdeckung(t *testing.T) {
	daten, fenster, _ := buildSignierteDaten(t)

	signiertProTyp := map[string]int{}
	for _, ev := range daten.Events {
		felder := parseData[tseFelder](t, ev.event)
		erwarteterProcessType, fiskalisch := fiskalischeTypen[ev.event.Type]

		if !fiskalisch {
			if felder.TxID != "" || felder.Data != nil {
				t.Errorf("%s (%s v%d): nicht-fiskalischer Typ trägt TSE-Felder", ev.event.Type, ev.event.Subject, ev.event.Version)
			}
			continue
		}

		if felder.TxID == "" {
			t.Errorf("%s (%s v%d): fiskalisches Event ohne txID", ev.event.Type, ev.event.Subject, ev.event.Version)
			continue
		}
		if fensterFuer(fenster, ev.event.Time) != nil {
			if felder.Data != nil {
				t.Errorf("%s (%s v%d): im Ausfallfenster signiert", ev.event.Type, ev.event.Subject, ev.event.Version)
			}
			continue
		}
		if felder.Data == nil {
			t.Errorf("%s (%s v%d): fiskalisches Event außerhalb des Fensters ohne TSE-Daten", ev.event.Type, ev.event.Subject, ev.event.Version)
			continue
		}
		if felder.Data.ProcessType != erwarteterProcessType {
			t.Errorf("%s: processType = %q, erwartet %q", ev.event.Type, felder.Data.ProcessType, erwarteterProcessType)
		}
		signiertProTyp[ev.event.Type]++
	}

	for typ := range fiskalischeTypen {
		if signiertProTyp[typ] == 0 {
			t.Errorf("Event-Typ %s: kein signiertes Vorkommen im Szenario", typ)
		}
	}
}

// TestSigniereEvents_MonotonieUndValidierung prüft die globalen TSE-Zähler: streng monotone
// Transaktionsnummern und Signaturzähler (Events in Event-Reihenfolge, nachgetragene
// Signaturen in Quittier-Reihenfolge), lückenlose Nummern über beide Quellen, die feste
// Seriennummer, logTime-Paare aus den Event-Zeitstempeln und kasse.TSEData.Validate.
func TestSigniereEvents_MonotonieUndValidierung(t *testing.T) {
	daten, _, seiten := buildSignierteDaten(t)

	vergeben := map[int]bool{}
	pruefe := func(quelle string, d kasse.TSEData, vorherTx, vorherSig int) (int, int) {
		t.Helper()
		if err := d.Validate(); err != nil {
			t.Errorf("%s: TSEData ungültig: %v", quelle, err)
		}
		if d.SerialNumberTSE != fakeTSESeriennummer {
			t.Errorf("%s: Seriennummer %q, erwartet %q", quelle, d.SerialNumberTSE, fakeTSESeriennummer)
		}
		if !strings.HasPrefix(d.QRCodeData, "V0;") {
			t.Errorf("%s: QR-Code-Daten ohne V0-Präfix: %q", quelle, d.QRCodeData)
		}
		if d.TransactionNumber <= vorherTx {
			t.Errorf("%s: Transaktionsnummer %d nicht streng monoton nach %d", quelle, d.TransactionNumber, vorherTx)
		}
		if d.SignatureCounter <= vorherSig {
			t.Errorf("%s: Signaturzähler %d nicht streng monoton nach %d", quelle, d.SignatureCounter, vorherSig)
		}
		if vergeben[d.TransactionNumber] {
			t.Errorf("%s: Transaktionsnummer %d doppelt vergeben", quelle, d.TransactionNumber)
		}
		vergeben[d.TransactionNumber] = true
		return d.TransactionNumber, d.SignatureCounter
	}

	vorherTx, vorherSig := 0, 0
	for _, ev := range daten.Events {
		felder := parseData[tseFelder](t, ev.event)
		if felder.Data == nil {
			continue
		}
		quelle := ev.event.Subject + " v" + ev.event.Type
		if felder.Data.LogTimeStart != ev.event.Time.UTC().Format(time.RFC3339) {
			t.Errorf("%s: logTimeStart %q entspricht nicht dem Event-Zeitstempel %v", quelle, felder.Data.LogTimeStart, ev.event.Time)
		}
		vorherTx, vorherSig = pruefe(quelle, *felder.Data, vorherTx, vorherSig)
	}

	// processType der nachgetragenen Signaturen liefert der zugehörige Auftrag.
	processTypeProTxID := map[string]string{}
	for _, a := range seiten.Auftraege {
		processTypeProTxID[a.TxID] = a.ProcessType
	}

	vorherTx, vorherSig = 0, 0
	for _, sig := range seiten.Signaturen {
		d := kasse.TSEData{
			TransactionNumber: sig.TransaktionNummer,
			SignatureCounter:  sig.SignaturZaehler,
			SerialNumberTSE:   sig.TSESeriennummer,
			LogTimeStart:      sig.LogTimeStart.UTC().Format(time.RFC3339),
			LogTimeEnd:        sig.LogTimeEnd.UTC().Format(time.RFC3339),
			Signature:         sig.Signatur,
			ProcessType:       processTypeProTxID[sig.TxID],
			QRCodeData:        sig.QRCodeData,
		}
		vorherTx, vorherSig = pruefe("signatur "+sig.TxID, d, vorherTx, vorherSig)
	}

	for nr := 1; nr <= len(vergeben); nr++ {
		if !vergeben[nr] {
			t.Errorf("Transaktionsnummer %d fehlt (Lücke)", nr)
		}
	}
}

// TestSigniereEvents_Ausfallfenster prüft die Paarigkeit: Jedes fiskalische Event im
// Ausfallfenster hat genau einen Nachsignier-Auftrag (Beginn = Event-Zeitstempel) und das
// TSEAusfall-Flag, wo der Event-Typ es vorsieht. Erledigte Aufträge tragen eine nachgetragene
// Signatur-Zeile, dauerhafte Fehlschläge einen Fehlertext; die Statusverteilung stimmt
// (überwiegend erledigt, einzelne fehlgeschlagen, genau einer verworfen, offene am Sonntag).
func TestSigniereEvents_Ausfallfenster(t *testing.T) {
	daten, fenster, seiten := buildSignierteDaten(t)

	auftragProTxID := map[string]nachsignierAuftragZeile{}
	for _, a := range seiten.Auftraege {
		if _, doppelt := auftragProTxID[a.TxID]; doppelt {
			t.Errorf("Nachsignier-Auftrag %s doppelt", a.TxID)
		}
		auftragProTxID[a.TxID] = a
	}
	signaturProTxID := map[string]tseSignaturZeile{}
	for _, sig := range seiten.Signaturen {
		signaturProTxID[sig.TxID] = sig
	}

	ausfallEvents := 0
	for _, ev := range daten.Events {
		felder := parseData[tseFelder](t, ev.event)
		if felder.TxID == "" {
			continue // nicht fiskalisch
		}
		if fensterFuer(fenster, ev.event.Time) == nil {
			if _, hat := auftragProTxID[felder.TxID]; hat {
				t.Errorf("Event %s außerhalb des Fensters hat einen Nachsignier-Auftrag", felder.TxID)
			}
			continue
		}

		ausfallEvents++
		auftrag, ok := auftragProTxID[felder.TxID]
		if !ok {
			t.Errorf("Ausfall-Event %s (%s) ohne Nachsignier-Auftrag", felder.TxID, ev.event.Type)
			continue
		}
		if !auftrag.ErstelltAm.Equal(ev.event.Time) {
			t.Errorf("Auftrag %s: erstellt_am %v ≠ Event-Zeitstempel %v", felder.TxID, auftrag.ErstelltAm, ev.event.Time)
		}
		typ := ev.event.Type
		if typ == string(kasse.EventTypeZahlungKassiertV1) || typ == string(kasse.EventTypeDirektverkaufGetaetigtV1) {
			if !felder.Ausfall {
				t.Errorf("Event %s (%s) im Ausfallfenster ohne TSEAusfall-Flag", felder.TxID, typ)
			}
		}
	}
	if ausfallEvents != len(seiten.Auftraege) {
		t.Errorf("%d Ausfall-Events, aber %d Nachsignier-Aufträge", ausfallEvents, len(seiten.Auftraege))
	}

	statusZahl := map[string]int{}
	for _, a := range seiten.Auftraege {
		statusZahl[a.Status]++
		f := fensterFuer(fenster, a.ErstelltAm)
		if f == nil {
			t.Errorf("Auftrag %s: erstellt_am %v liegt in keinem Ausfallfenster", a.TxID, a.ErstelltAm)
			continue
		}
		_, hatSignatur := signaturProTxID[a.TxID]

		switch a.Status {
		case "offen":
			if f.aufgeloest {
				t.Errorf("offener Auftrag %s in aufgelöstem Fenster", a.TxID)
			}
			if a.Versuche != 0 || a.LetzterFehler != nil || a.ErledigtAm != nil || hatSignatur {
				t.Errorf("offener Auftrag %s: erwartet 0 Versuche, kein Fehler, keine Erledigung, keine Signatur", a.TxID)
			}
		case "erledigt":
			if !f.aufgeloest {
				t.Errorf("erledigter Auftrag %s in offenem Fenster", a.TxID)
			}
			if !hatSignatur {
				t.Errorf("erledigter Auftrag %s ohne nachgetragene Signatur-Zeile", a.TxID)
			}
			if a.ErledigtAm == nil || a.ErledigtAm.Before(f.bis) {
				t.Errorf("erledigter Auftrag %s: erledigt_am fehlt oder liegt vor dem Fensterende", a.TxID)
			}
		case "fehlgeschlagen", "verworfen":
			if !f.aufgeloest {
				t.Errorf("%s Auftrag %s in offenem Fenster", a.Status, a.TxID)
			}
			if a.LetzterFehler == nil || *a.LetzterFehler == "" {
				t.Errorf("%s Auftrag %s ohne Fehlertext", a.Status, a.TxID)
			}
			if a.ErledigtAm != nil || hatSignatur {
				t.Errorf("%s Auftrag %s: erwartet keine Erledigung und keine Signatur", a.Status, a.TxID)
			}
		default:
			t.Errorf("Auftrag %s: unbekannter Status %q", a.TxID, a.Status)
		}
	}

	if statusZahl["offen"] < 1 {
		t.Error("kein offener Nachsignier-Auftrag (Sonntags-Aussetzer fehlt)")
	}
	if statusZahl["fehlgeschlagen"] < 2 {
		t.Errorf("nur %d fehlgeschlagene Aufträge, erwartet einzelne (≥ 2)", statusZahl["fehlgeschlagen"])
	}
	if statusZahl["verworfen"] != 1 {
		t.Errorf("%d verworfene Aufträge, erwartet genau 1", statusZahl["verworfen"])
	}
	rest := statusZahl["offen"] + statusZahl["fehlgeschlagen"] + statusZahl["verworfen"]
	if statusZahl["erledigt"] <= rest {
		t.Errorf("erledigte Aufträge (%d) nicht in der Überzahl gegenüber den übrigen (%d)", statusZahl["erledigt"], rest)
	}
	if len(seiten.Signaturen) != statusZahl["erledigt"] {
		t.Errorf("%d Signatur-Zeilen, aber %d erledigte Aufträge", len(seiten.Signaturen), statusZahl["erledigt"])
	}
}
