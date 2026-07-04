package seed

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// kassenbelegJederNte steuert, wie oft Gäste einen Kassenbeleg verlangt haben: Jede n-te
// Zahlung bzw. jeder n-te Direktverkauf erhält einen Kassenbeleg-Druckauftrag.
const kassenbelegJederNte = 5

// belegVerlangtNach ist der Abstand zwischen dem Vorgang und dem verlangten Kassenbeleg.
const belegVerlangtNach = 30 * time.Second

// druckQuittungNach ist die Zeit von der Auftragserstellung bis zur Gedruckt-Quittung des Relays.
const druckQuittungNach = 4 * time.Second

// relayAbholFenster: Jüngere Aufträge hat das Relay noch nicht abgeholt — sie bleiben offen.
const relayAbholFenster = 15 * time.Minute

// druckauftragZeile ist die zu persistierende Zeile der druckauftraege-Tabelle.
type druckauftragZeile struct {
	ZielIP        string
	Payload       string // Base64-kodierte ESC/POS-Bytes
	Status        string
	BonArt        string
	Referenz      string
	Versuche      int
	LetzterFehler *string
	ErstelltAm    time.Time
	GedrucktAm    *time.Time
}

// druckerFenster ist ein Drucker-Ausfallfenster mit absoluten Zeiten je Drucker-IP.
type druckerFenster struct {
	ip         string
	von, bis   time.Time
	fehlertext string
}

// druckerFensterAus übersetzt die Drucker-Ausfälle des Drehbuchs in absolute Zeitfenster.
func druckerFensterAus(s szenario, jetzt time.Time) []druckerFenster {
	ipProKategorie := make(map[druckstation.Kategorie]string, len(s.Druckstationen))
	for _, st := range s.Druckstationen {
		ipProKategorie[st.Kategorie] = st.DruckerIP
	}

	var fenster []druckerFenster
	for i := range s.Sitzungen {
		sitzung := &s.Sitzungen[i]
		start := sitzung.startZeit(jetzt)
		for _, a := range sitzung.DruckerAusfaelle {
			fenster = append(fenster, druckerFenster{
				ip:         ipProKategorie[a.Kategorie],
				von:        start.Add(a.NachStart),
				bis:        start.Add(a.NachStart + a.Dauer),
				fehlertext: a.Fehlertext,
			})
		}
	}
	return fenster
}

// baueDruckauftraege baut die Druckauftrags-Historie zum Szenario: Arbeits- und Abholbons
// entstehen über die produktive Bondruck-Policy aus jeder Bestellung und jedem Direktverkauf,
// Kassenbelege (inklusive TSE-Abschnitt aus den Signaturspalten des Auftrags) für jede n-te
// Zahlung über den produktiven ESC/POS-Formatter — nur für Vorgänge mit quittierter
// Signatur, denn bei ausstehender Signatur entsteht im neuen Modell kein Druckauftrag.
// Der Status ergibt sich aus den Drucker-Ausfallfenstern des Drehbuchs (fehlgeschlagen, der
// erste Fehlschlag verworfen), dem Relay-Abholfenster vor „jetzt" (offen) und sonst der
// Gedruckt-Quittung kurz nach der Erstellung.
func baueDruckauftraege(s szenario, events []seedEvent, signaturen map[int]*tse.Signatur, jetzt time.Time) ([]druckauftragZeile, error) {
	stationen := make(map[string]bondruckApp.Druckstation, len(s.Druckstationen))
	for _, st := range s.Druckstationen {
		stationen[string(st.Kategorie)] = bondruckApp.Druckstation{IP: st.DruckerIP, Bonmodus: string(st.Bonmodus)}
	}

	b := &bondruckBauer{
		betreiber:       s.Betreiber,
		stationen:       stationen,
		fenster:         druckerFensterAus(s, jetzt),
		signaturen:      signaturen,
		jetzt:           jetzt,
		ersteBestellung: map[string]time.Time{},
	}

	var zeilen []druckauftragZeile
	for i := range events {
		evt := events[i].event

		for _, auftrag := range bondruckApp.CreateArbeitsbonAuftraegeFromEvent(evt, stationen) {
			zeilen = append(zeilen, b.zeile(auftrag, evt.Time))
		}

		switch kasse.EventType(evt.Type) {
		case kasse.EventTypeBestellungAufgenommenV1:
			if _, ok := b.ersteBestellung[evt.Subject]; !ok {
				b.ersteBestellung[evt.Subject] = evt.Time
			}
		case kasse.EventTypeZahlungKassiertV1, kasse.EventTypeDirektverkaufGetaetigtV1:
			b.belegZaehler++
			if b.belegZaehler%kassenbelegJederNte != 0 {
				continue
			}
			zeile, ok, err := b.kassenbeleg(evt)
			if err != nil {
				return nil, fmt.Errorf("event %s v%d: kassenbeleg: %w", evt.Subject, evt.Version, err)
			}
			if !ok {
				continue
			}
			zeilen = append(zeilen, zeile)
		}
	}

	return zeilen, nil
}

// bondruckBauer hält die Drucker-Ausfallfenster, die quittierten Signaturen je Event-ID,
// den Beleg-Zähler und die erste Bestellzeit je Subject (für die „Erste Bestellung"-Zeile
// auf dem Kassenbeleg).
type bondruckBauer struct {
	betreiber       settings.Betreiber
	stationen       map[string]bondruckApp.Druckstation
	fenster         []druckerFenster
	signaturen      map[int]*tse.Signatur
	jetzt           time.Time
	ersteBestellung map[string]time.Time

	belegZaehler      int
	verworfenVergeben bool
}

// zeile versieht einen Auftrag der produktiven Policy mit Status und historischen Zeitstempeln.
func (b *bondruckBauer) zeile(auftrag druckauftrag_repo.NeuerDruckauftrag, erstellt time.Time) druckauftragZeile {
	z := druckauftragZeile{
		ZielIP:     auftrag.ZielIP,
		Payload:    auftrag.Payload,
		BonArt:     auftrag.BonArt,
		Referenz:   auftrag.Referenz,
		ErstelltAm: erstellt,
	}
	b.setzeStatus(&z)
	return z
}

// setzeStatus bestimmt den Status: In einem Drucker-Ausfallfenster scheitert der Auftrag nach
// den maximalen Zustellversuchen (den ersten Fehlschlag insgesamt hat der Admin verworfen),
// die jüngsten Aufträge hat das Relay noch nicht abgeholt (offen), alle übrigen sind gedruckt.
func (b *bondruckBauer) setzeStatus(z *druckauftragZeile) {
	for _, f := range b.fenster {
		if z.ZielIP != f.ip || z.ErstelltAm.Before(f.von) || !z.ErstelltAm.Before(f.bis) {
			continue
		}
		z.Status = "fehlgeschlagen"
		if !b.verworfenVergeben {
			z.Status = "verworfen"
			b.verworfenVergeben = true
		}
		z.Versuche = druckauftrag_repo.MaxDruckversuche
		fehler := f.fehlertext
		z.LetzterFehler = &fehler
		return
	}

	if z.ErstelltAm.After(b.jetzt.Add(-relayAbholFenster)) {
		z.Status = "offen"
		return
	}

	z.Status = "gedruckt"
	gedruckt := z.ErstelltAm.Add(druckQuittungNach)
	z.GedrucktAm = &gedruckt
}

// kassenbeleg baut den Kassenbeleg-Druckauftrag zu einer Zahlung oder einem Direktverkauf —
// wie KassenbelegDrucken im Produktivbetrieb, mit dem TSE-Abschnitt aus den Signaturspalten
// des Auftrags. Vorgänge ohne quittierte Signatur (offen, fehlgeschlagen, tse_nicht_konfiguriert) liefern
// keinen Beleg: Der Beleg-Abruf hätte „ausstehend" geantwortet und keinen Druckauftrag
// angelegt. Der Beleg entsteht kurz nach der Quittierung (bei nachsignierten Vorgängen also
// erst nach der Störung, auf erneute Anforderung des Gastes). Als Kassen-ID steht die
// Fake-Seriennummer auf dem Beleg — dieselbe wie in den QR-Code-Daten der Fake-TSE.
func (b *bondruckBauer) kassenbeleg(evt e.Event) (druckauftragZeile, bool, error) {
	signatur := b.signaturen[evt.ID]
	if signatur == nil {
		return druckauftragZeile{}, false, nil
	}

	var positionen []kasse.Position
	var gesamtbetragCents int
	var referenz string
	var ersteBestellung *time.Time

	switch kasse.EventType(evt.Type) {
	case kasse.EventTypeZahlungKassiertV1:
		data, err := parseEventData[kasse.ZahlungKassiertV1Data](evt)
		if err != nil {
			return druckauftragZeile{}, false, err
		}
		positionen = zuPositionen(data.Positionen)
		gesamtbetragCents = data.GesamtZahlungCents
		referenz = fmt.Sprintf("zahlung-kassiert:%d", evt.ID)
		if t, ok := b.ersteBestellung[evt.Subject]; ok {
			ersteBestellung = &t
		}

	case kasse.EventTypeDirektverkaufGetaetigtV1:
		data, err := parseEventData[kasse.DirektverkaufGetaetigtV1Data](evt)
		if err != nil {
			return druckauftragZeile{}, false, err
		}
		positionen = zuPositionen(data.Positionen)
		gesamtbetragCents = data.GesamtbetragCents
		referenz = fmt.Sprintf("direktverkauf-getaetigt:%d", evt.ID)

	default:
		return druckauftragZeile{}, false, fmt.Errorf("kein belegfähiger Event-Typ %q", evt.Type)
	}

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:              b.betreiber.Vereinsname,
		Strasse:                  b.betreiber.Strasse,
		Plz:                      b.betreiber.Plz,
		Ort:                      b.betreiber.Ort,
		KassenSeriennummer:       fakeKassenSeriennummer,
		Belegnummer:              fmt.Sprintf("%d", evt.ID),
		Zeitpunkt:                evt.Time,
		ErsteBestellungZeitpunkt: ersteBestellung,
		Positionen:               positionen,
		Steuermatrix:             steuer.Steuermatrix(zuSteuermatrixPositionen(positionen)),
		TSE:                      zuTSEAbschnitt(signatur),
		GesamtbetragCents:        gesamtbetragCents,
		Zahlungsart:              "bar",
	})

	z := druckauftragZeile{
		ZielIP:     b.stationen[string(druckstation.KategorieKassenbeleg)].IP,
		Payload:    base64.StdEncoding.EncodeToString(payload),
		BonArt:     "kassenbeleg",
		Referenz:   referenz,
		ErstelltAm: signatur.LogTimeEnd.Add(belegVerlangtNach),
	}
	b.setzeStatus(&z)
	return z, true, nil
}

// zuTSEAbschnitt wandelt die quittierte Signatur des Auftrags in den Beleg-Abschnitt, wie
// der Belegdruck im Produktivbetrieb.
func zuTSEAbschnitt(s *tse.Signatur) *escpos.TSEAbschnitt {
	return &escpos.TSEAbschnitt{
		TransaktionNr:   s.TransaktionNummer,
		Signaturzaehler: s.SignaturZaehler,
		TSESeriennummer: s.TSESeriennummer,
		ZeitpunktBeginn: s.LogTimeStart,
		ZeitpunktEnde:   s.LogTimeEnd,
		Signatur:        s.Signatur,
		QRCodeData:      s.QRCodeData,
	}
}

func parseEventData[T any](evt e.Event) (T, error) {
	var data T
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return data, fmt.Errorf("event-daten parsen: %w", err)
	}
	return data, nil
}

// zuPositionen wandelt die persistierte Positions-Darstellung zurück in Domain-Positionen.
func zuPositionen(eventPositionen []kasse.PositionEventData) []kasse.Position {
	positionen := make([]kasse.Position, len(eventPositionen))
	for i, p := range eventPositionen {
		positionen[i] = kasse.PositionFromEventData(p)
	}
	return positionen
}

func zuSteuermatrixPositionen(positionen []kasse.Position) []steuer.SteuermatrixPosition {
	matrixPositionen := make([]steuer.SteuermatrixPosition, 0, len(positionen))
	for _, p := range positionen {
		matrixPositionen = append(matrixPositionen, steuer.SteuermatrixPosition{
			Brutto:     p.Einzelpreis * p.Menge,
			Steuersatz: steuer.Steuersatz(p.Steuersatz),
		})
	}
	return matrixPositionen
}
