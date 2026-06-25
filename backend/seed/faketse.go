package seed

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// Feste Fake-TSE-Identität für das Demo-Szenario. Die Werte sind frei erfunden, folgen aber
// den Formaten einer echten fiskaly Cloud-TSE (Seriennummer = SHA-256-Hex, Schlüssel und
// Signaturen = Base64).
const (
	fakeTSESeriennummer     = "9c4f2d8a71e3b65042dca9f01b87e6d355a1c0fb29e84d7613f5a2b8c90e4761"
	fakeKassenSeriennummer  = "JOTTI-DEMO-KASSE-1"
	fakeSignaturAlgorithmus = "ecdsa-plain-SHA256"
	fakeLogTimeFormat       = "unixTime"
	fakePublicKey           = "BJottiDemoFakeTSEPublicKeySommerfestTSVMusterstadt2026AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
)

// nachsignierungAbgelehntFehler ist der Fehlertext der dauerhaft gescheiterten Aufträge:
// Anders als beim vorübergehenden Ausfall (Fenster-Grund) lehnt die TSE diese Transaktionen
// auch nach der Störung ab — nur solche Aufträge bleiben fehlgeschlagen bzw. werden verworfen.
const nachsignierungAbgelehntFehler = "Cloud-TSE lehnt die Transaktion dauerhaft ab (HTTP 400: ungültige process_data)"

// fehlschlagJederNte steuert die Dramaturgie aufgelöster Ausfallfenster: Jeder 16. Auftrag
// (ab dem vierten) scheitert dauerhaft; der erste dieser Fehlschläge wurde vom Admin
// verworfen, die übrigen bleiben fehlgeschlagen.
const fehlschlagJederNte = 16

// ausfallFenster ist ein TSE-Ausfallfenster mit absoluten Zeiten. aufgeloest steuert, ob die
// Nachsignier-Aufträge als vom Worker abgearbeitet gelten (abgeschlossene Sitzung) oder
// offen bleiben (offene Sitzung, Worker ohne TSE-Konfiguration inaktiv).
type ausfallFenster struct {
	von, bis   time.Time
	grund      string
	aufgeloest bool
}

// ausfallFensterAus übersetzt die TSE-Ausfälle des Drehbuchs in absolute Zeitfenster.
func ausfallFensterAus(s szenario, jetzt time.Time) []ausfallFenster {
	var fenster []ausfallFenster
	for i := range s.Sitzungen {
		sitzung := &s.Sitzungen[i]
		start := sitzung.startZeit(jetzt)
		for _, a := range sitzung.TSEAusfaelle {
			fenster = append(fenster, ausfallFenster{
				von:        start.Add(a.NachStart),
				bis:        start.Add(a.NachStart + a.Dauer),
				grund:      a.Grund,
				aufgeloest: sitzung.Abgeschlossen,
			})
		}
	}
	return fenster
}

// nachsignierAuftragZeile ist die zu persistierende Zeile der tse_nachsignier_auftraege-Tabelle.
type nachsignierAuftragZeile struct {
	TxID               string
	ProcessType        string
	ProcessData        string
	Status             string
	Versuche           int
	LetzterFehler      *string
	NaechsterVersuchAm time.Time
	ErstelltAm         time.Time
	ErledigtAm         *time.Time
}

// tseSignaturZeile ist die zu persistierende Zeile der tse_signaturen-Tabelle
// (nachgetragene Signatur eines erledigten Nachsignier-Auftrags).
type tseSignaturZeile struct {
	TxID              string
	TransaktionNummer int
	SignaturZaehler   int
	TSESeriennummer   string
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	Signatur          string
	QRCodeData        string
	ErstelltAm        time.Time
}

// tseSeitentabellen bündelt die vom Fake-Signierer erzeugten Zeilen der TSE-Seitentabellen.
type tseSeitentabellen struct {
	Auftraege  []nachsignierAuftragZeile
	Signaturen []tseSignaturZeile
}

// fiskalischeSignierung beschreibt, wie ein fiskalischer Event-Typ signiert wird:
// processType/processData nach DSFinV-K (über die produktiven Builder) plus die
// Embed-Funktion der Domain-Schicht.
type fiskalischeSignierung struct {
	processType string
	processData string
	embed       tseApp.EmbedTSE
}

// signiereEvents versieht genau die fiskalischen Events mit Fake-TSE-Daten: global streng
// monotone Transaktionsnummern und Signaturzähler, logTime-Paare aus den Event-Zeitstempeln,
// Signatur- und QR-Code-Daten im fiskaly-Format. Events in einem Ausfallfenster erhalten nur
// die txID; je Event entsteht ein Nachsignier-Auftrag. Aufgelöste Fenster werden beim ersten
// fiskalischen Event nach Fensterende nachsigniert (der Worker holt die Signaturen nach,
// bevor neue Transaktionsnummern vergeben werden), Fenster der offenen Sitzung lassen ihre
// Aufträge offen. Die Events werden in place um die TSE-Daten erweitert.
func signiereEvents(events []seedEvent, fenster []ausfallFenster) (tseSeitentabellen, error) {
	s := &fakeSignierer{fenster: fenster, pending: make([][]offeneNachsignierung, len(fenster))}

	for i := range events {
		evt := events[i].event
		spec, fiskalisch, err := fiskalischeSignierungFuer(evt)
		if err != nil {
			return tseSeitentabellen{}, fmt.Errorf("event %s v%d: %w", evt.Subject, evt.Version, err)
		}
		if !fiskalisch {
			continue
		}

		s.nachsigniereFaelligeFenster(evt.Time)

		s.txSeq++
		txID := tseTxID(s.txSeq)

		if f := s.fensterIndex(evt.Time); f >= 0 {
			unsigniert, err := spec.embed(evt, txID, nil)
			if err != nil {
				return tseSeitentabellen{}, fmt.Errorf("event %s v%d: tse-ausfall einbetten: %w", evt.Subject, evt.Version, err)
			}
			events[i].event = unsigniert
			s.vermerkeAusfall(f, txID, spec, evt.Time)
			continue
		}

		tseData := s.signiere(spec.processType, spec.processData, evt.Time, evt.Time.Add(time.Second), txID)
		signiert, err := spec.embed(evt, txID, &tseData)
		if err != nil {
			return tseSeitentabellen{}, fmt.Errorf("event %s v%d: tse-daten einbetten: %w", evt.Subject, evt.Version, err)
		}
		events[i].event = signiert
	}

	s.nachsigniereAlleFenster()
	return s.seiten, nil
}

// offeneNachsignierung ist ein Ausfall-Vorgang, der nach Fensterende nachsigniert wird.
type offeneNachsignierung struct {
	txID        string
	processType string
	processData string
	erstellt    time.Time
}

// fakeSignierer hält die globalen TSE-Zähler und sammelt die Seitentabellen-Zeilen.
type fakeSignierer struct {
	fenster []ausfallFenster
	pending [][]offeneNachsignierung

	txSeq             int // laufende Nummer für die txID-Vergabe (alle fiskalischen Events)
	txNummer          int // TSE-Transaktionsnummer (nur tatsächlich signierte Vorgänge)
	sigZaehler        int
	verworfenVergeben bool

	seiten tseSeitentabellen
}

// signiere vergibt die nächste Transaktionsnummer samt Signaturzähler und baut die TSE-Daten.
// Jede Transaktion verbraucht zwei Signaturen (Start + Finish) — der Zähler im Beleg ist der
// der Finish-Signatur, wie bei einer echten TSE.
func (s *fakeSignierer) signiere(processType, processData string, logStart, logEnd time.Time, txID string) kasse.TSEData {
	s.txNummer++
	s.sigZaehler += 2
	signatur := fakeSignatur(txID, s.txNummer)
	return kasse.TSEData{
		TransactionNumber: s.txNummer,
		SignatureCounter:  s.sigZaehler,
		SerialNumberTSE:   fakeTSESeriennummer,
		LogTimeStart:      logStart.UTC().Format(time.RFC3339),
		LogTimeEnd:        logEnd.UTC().Format(time.RFC3339),
		Signature:         signatur,
		ProcessType:       processType,
		QRCodeData:        qrCodeData(processType, processData, s.txNummer, s.sigZaehler, logStart, logEnd, signatur),
	}
}

func (s *fakeSignierer) fensterIndex(t time.Time) int {
	for i, f := range s.fenster {
		if !t.Before(f.von) && t.Before(f.bis) {
			return i
		}
	}
	return -1
}

// vermerkeAusfall registriert den Ausfall eines Events: In aufgelösten Fenstern wandert der
// Vorgang in die Warteliste der späteren Nachsignierung, in offenen Fenstern entsteht sofort
// ein offener Auftrag — so wie ihn der Produktivpfad bei der Ausfall-Persistenz anlegt.
func (s *fakeSignierer) vermerkeAusfall(fensterIdx int, txID string, spec fiskalischeSignierung, zeit time.Time) {
	if s.fenster[fensterIdx].aufgeloest {
		s.pending[fensterIdx] = append(s.pending[fensterIdx], offeneNachsignierung{
			txID:        txID,
			processType: spec.processType,
			processData: spec.processData,
			erstellt:    zeit,
		})
		return
	}
	s.seiten.Auftraege = append(s.seiten.Auftraege, nachsignierAuftragZeile{
		TxID:               txID,
		ProcessType:        spec.processType,
		ProcessData:        spec.processData,
		Status:             "offen",
		NaechsterVersuchAm: zeit,
		ErstelltAm:         zeit,
	})
}

func (s *fakeSignierer) nachsigniereFaelligeFenster(jetzt time.Time) {
	for i, f := range s.fenster {
		if len(s.pending[i]) > 0 && !jetzt.Before(f.bis) {
			s.nachsigniereFenster(i)
		}
	}
}

func (s *fakeSignierer) nachsigniereAlleFenster() {
	for i := range s.fenster {
		if len(s.pending[i]) > 0 {
			s.nachsigniereFenster(i)
		}
	}
}

// nachsigniereFenster spielt den Nachsignier-Worker nach dem Fensterende nach: Die Aufträge
// werden in Batches im Sekundenabstand quittiert; je erledigtem Auftrag entsteht die
// nachgetragene Signatur-Zeile. Versuche und letzter Fehler ergeben sich aus den
// Fehlversuchen, die mit dem Worker-Backoff noch ins Ausfallfenster fielen.
func (s *fakeSignierer) nachsigniereFenster(fensterIdx int) {
	f := s.fenster[fensterIdx]
	for i, p := range s.pending[fensterIdx] {
		if i%fehlschlagJederNte == 3 {
			s.seiten.Auftraege = append(s.seiten.Auftraege, s.dauerhaftGescheitert(p))
			continue
		}

		quittiert := f.bis.Add(5*time.Second + time.Duration(i)*250*time.Millisecond)
		versuche := fehlversucheBis(p.erstellt, f.bis)
		var fehler *string
		if versuche > 0 {
			grund := f.grund
			fehler = &grund
		}

		tseData := s.signiere(p.processType, p.processData, quittiert, quittiert.Add(time.Second), p.txID)

		s.seiten.Auftraege = append(s.seiten.Auftraege, nachsignierAuftragZeile{
			TxID:               p.txID,
			ProcessType:        p.processType,
			ProcessData:        p.processData,
			Status:             "erledigt",
			Versuche:           versuche,
			LetzterFehler:      fehler,
			NaechsterVersuchAm: quittiert,
			ErstelltAm:         p.erstellt,
			ErledigtAm:         &quittiert,
		})
		s.seiten.Signaturen = append(s.seiten.Signaturen, tseSignaturZeile{
			TxID:              p.txID,
			TransaktionNummer: tseData.TransactionNumber,
			SignaturZaehler:   tseData.SignatureCounter,
			TSESeriennummer:   tseData.SerialNumberTSE,
			LogTimeStart:      quittiert,
			LogTimeEnd:        quittiert.Add(time.Second),
			Signatur:          tseData.Signature,
			QRCodeData:        tseData.QRCodeData,
			ErstelltAm:        quittiert.Add(time.Second),
		})
	}
	s.pending[fensterIdx] = nil
}

// dauerhaftGescheitert baut den fehlgeschlagenen bzw. (genau einmal) verworfenen Auftrag.
func (s *fakeSignierer) dauerhaftGescheitert(p offeneNachsignierung) nachsignierAuftragZeile {
	status := "fehlgeschlagen"
	if !s.verworfenVergeben {
		status = "verworfen"
		s.verworfenVergeben = true
	}
	fehler := nachsignierungAbgelehntFehler
	return nachsignierAuftragZeile{
		TxID:          p.txID,
		ProcessType:   p.processType,
		ProcessData:   p.processData,
		Status:        status,
		Versuche:      tse_repo.MaxNachsignierVersuche,
		LetzterFehler: &fehler,
		// Nach den maximalen Fehlversuchen mit gedeckeltem Backoff läge der (nie mehr
		// genutzte) nächste Versuch gut drei Stunden nach dem Auftrag.
		NaechsterVersuchAm: p.erstellt.Add(3*time.Hour + 30*time.Minute),
		ErstelltAm:         p.erstellt,
	}
}

// fehlversucheBis zählt, wie viele Nachsignier-Versuche mit dem Worker-Backoff (1, 2, 4, …
// Minuten, gedeckelt auf 30) noch vor dem Fensterende lagen — sie scheiterten alle; der
// erste Versuch nach der Störung gelang. Gedeckelt unter dem Produktiv-Maximum, sonst wäre
// der Auftrag fehlgeschlagen statt erledigt.
func fehlversucheBis(erstellt, fensterEnde time.Time) int {
	versuche := 0
	naechster := erstellt
	backoff := time.Minute
	for {
		naechster = naechster.Add(backoff)
		if !naechster.Before(fensterEnde) || versuche >= tse_repo.MaxNachsignierVersuche-1 {
			return versuche
		}
		versuche++
		backoff = min(backoff*2, 30*time.Minute)
	}
}

// fiskalischeSignierungFuer baut für fiskalische Event-Typen processType und processData aus
// den Event-Daten (über die produktiven DSFinV-K-Builder) und liefert die passende
// Embed-Funktion. Nicht-fiskalische Typen (Eröffnung, Ausgabe, Kassensturz) liefern ok = false.
func fiskalischeSignierungFuer(evt e.Event) (fiskalischeSignierung, bool, error) {
	switch kasse.EventType(evt.Type) {
	case kasse.EventTypeBestellungAufgenommenV1:
		data, err := parseEventData[kasse.BestellungAufgenommenV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildBestellungProcessData(zuPositionen(data.Positionen))
		return signierung(tse.ProcessTypeBestellungV1, pd, kasse.EmbedTSEInBestellungAufgenommen, err)

	case kasse.EventTypeZahlungKassiertV1:
		data, err := parseEventData[kasse.ZahlungKassiertV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildKassenbelegProcessData(zuPositionen(data.Positionen), data.GesamtZahlungCents)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInZahlungKassiert, err)

	case kasse.EventTypeStornierungErteiltV1:
		data, err := parseEventData[kasse.StornierungErteiltV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildKassenbelegProcessDataWithFaktor(zuPositionen(data.Positionen), -data.GesamtStornierungCents, -1)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInStornierungErteilt, err)

	case kasse.EventTypeBestellungKorrigiertV1:
		data, err := parseEventData[kasse.BestellungKorrigiertV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildBestellungProcessData(zuPositionen(data.Positionen))
		return signierung(tse.ProcessTypeBestellungV1, pd, kasse.EmbedTSEInBestellungKorrigiert, err)

	case kasse.EventTypeBestellungUmgebuchtV1:
		data, err := parseEventData[kasse.BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildBestellungProcessData(zuPositionen(data.Positionen))
		return signierung(tse.ProcessTypeBestellungV1, pd, kasse.EmbedTSEInBestellungUmgebucht, err)

	case kasse.EventTypeDirektverkaufGetaetigtV1:
		data, err := parseEventData[kasse.DirektverkaufGetaetigtV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildKassenbelegProcessData(zuPositionen(data.Positionen), data.GesamtbetragCents)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInDirektverkaufGetaetigt, err)

	case kasse.EventTypeDirektverkaufStorniertV1:
		data, err := parseEventData[kasse.DirektverkaufStorniertV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildKassenbelegProcessDataWithFaktor(zuPositionen(data.Positionen), -data.GesamtStornierungCents, -1)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInDirektverkaufStorniert, err)

	case kasse.EventTypeGeldtransitGebuchtV1:
		data, err := parseEventData[kasse.GeldtransitGebuchtV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd, err := tseApp.BuildGeldtransitProcessData(data.Richtung, data.BetragCents)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInGeldtransitGebucht, err)

	case kasse.EventTypeDifferenzSollIstGebuchtV1:
		data, err := parseEventData[kasse.DifferenzSollIstGebuchtV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd := tseApp.BuildEigenbelegProcessData(data.BetragCents)
		return signierung(tse.ProcessTypeKassenbelegV1, pd, kasse.EmbedTSEInDifferenzSollIstGebucht, nil)

	case kasse.EventTypeTagesabschlussErstelltV1:
		data, err := parseEventData[kasse.TagesabschlussErstelltV1Data](evt)
		if err != nil {
			return fiskalischeSignierung{}, false, err
		}
		pd := tseApp.BuildTagesabschlussProcessData(data.ZNr, data.ZeitraumVon, data.ZeitraumBis)
		return signierung(tse.ProcessTypeSonstigerVorgang, pd, kasse.EmbedTSEInTagesabschlussErstellt, nil)

	case kasse.EventTypeKassensitzungEroeffnetV1, kasse.EventTypeAusgabeBestaetigtV1, kasse.EventTypeKassensturzDurchgefuehrtV1:
		return fiskalischeSignierung{}, false, nil

	default:
		return fiskalischeSignierung{}, false, fmt.Errorf("unbekannter Event-Typ %q", evt.Type)
	}
}

// signierung bündelt das Builder-Ergebnis; ein Builder-Fehler wird durchgereicht.
func signierung(processType, processData string, embed tseApp.EmbedTSE, err error) (fiskalischeSignierung, bool, error) {
	if err != nil {
		return fiskalischeSignierung{}, false, fmt.Errorf("process_data bauen: %w", err)
	}
	return fiskalischeSignierung{processType: processType, processData: processData, embed: embed}, true, nil
}

func parseEventData[T any](evt e.Event) (T, error) {
	var data T
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return data, fmt.Errorf("event-daten parsen: %w", err)
	}
	return data, nil
}

// zuPositionen wandelt die persistierte Positions-Darstellung zurück in Domain-Positionen,
// wie sie die processData-Builder erwarten.
func zuPositionen(eventPositionen []kasse.PositionEventData) []kasse.Position {
	positionen := make([]kasse.Position, len(eventPositionen))
	for i, p := range eventPositionen {
		positionen[i] = kasse.PositionFromEventData(p)
	}
	return positionen
}

// tseTxID liefert die feste TSE-Transaktions-ID nach erkennbarem Schema: Marker-Gruppe
// „aaaa“, letzte Gruppe = laufende Nummer des fiskalischen Vorgangs.
func tseTxID(nr int) string {
	return fmt.Sprintf("00000000-aaaa-4000-8000-%012d", nr)
}

// fakeSignatur erzeugt eine deterministische, Base64-kodierte Pseudo-Signatur.
func fakeSignatur(txID string, txNummer int) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "jotti-demo-tse:%s:%d", txID, txNummer))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// qrCodeData baut den KassenSichV-üblichen V0-String (BSI TR-03153-A), wie ihn fiskaly
// liefert; der Belegdruck rendert das Feld unverändert.
func qrCodeData(processType, processData string, txNummer, sigZaehler int, logStart, logEnd time.Time, signatur string) string {
	return strings.Join([]string{
		"V0",
		fakeKassenSeriennummer,
		processType,
		processData,
		strconv.Itoa(txNummer),
		strconv.Itoa(sigZaehler),
		logStart.UTC().Format(time.RFC3339),
		logEnd.UTC().Format(time.RFC3339),
		fakeSignaturAlgorithmus,
		fakeLogTimeFormat,
		signatur,
		fakePublicKey,
	}, ";")
}
