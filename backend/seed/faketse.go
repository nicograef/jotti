package seed

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// signierungAbgelehntFehler ist der Fehlertext der dauerhaft gescheiterten Aufträge:
// Anders als beim vorübergehenden Ausfall (Fenster-Grund) lehnt die TSE diese Transaktionen
// auch nach der Störung ab — nur solche Aufträge bleiben fehlgeschlagen.
const signierungAbgelehntFehler = "Cloud-TSE lehnt die Transaktion dauerhaft ab (HTTP 400: ungültige process_data)"

// fehlschlagJederNte steuert die Dramaturgie aufgelöster Ausfallfenster: Jeder 16. Auftrag
// (ab dem vierten) scheitert dauerhaft und bleibt fehlgeschlagen.
const fehlschlagJederNte = 16

// nachsignierVerzoegerung ist der Abstand zwischen Fensterende und der ersten erfolgreichen
// Nachsignierung — zugleich das Ende des geseedeten Störungszeitraums, denn im echten
// Betrieb schließt die erste erfolgreiche Signatur die Störung.
const nachsignierVerzoegerung = 5 * time.Second

// stoerungFehlertext ist der Fehlertext der geseedeten tse_fehler-Störungszeiträume.
const stoerungFehlertext = "Cloud-TSE nicht erreichbar (HTTP 503)"

// ausfallFenster ist ein TSE-Ausfallfenster mit absoluten Zeiten. aufgeloest steuert, ob die
// Signaturaufträge als vom Worker nachsigniert gelten (abgeschlossene Sitzung) oder offen
// bleiben (offene Sitzung).
type ausfallFenster struct {
	von, bis   time.Time
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
				aufgeloest: sitzung.Abgeschlossen,
			})
		}
	}
	return fenster
}

// stoerungZeile ist die zu persistierende Zeile der tse_stoerungen-Tabelle
// (Störungsprotokoll).
type stoerungZeile struct {
	Beginn     time.Time
	Ende       time.Time
	GrundArt   string
	Fehlertext string
}

// stoerungszeitraeumeAus übersetzt die aufgelösten Ausfallfenster in geschlossene
// tse_fehler-Störungszeiträume — was Worker und Störungsprotokoll im echten Betrieb
// dokumentiert hätten; so passt die Ausfalldokumentations-Ansicht zur Demo mit ihren
// nachsignierten Belegen. Das offene Fenster der laufenden Sitzung bleibt außen vor:
// Es materialisiert nach dem App-Start live über Worker und Watchdog.
func stoerungszeitraeumeAus(fenster []ausfallFenster) []stoerungZeile {
	var zeilen []stoerungZeile
	for _, f := range fenster {
		if !f.aufgeloest {
			continue
		}
		zeilen = append(zeilen, stoerungZeile{
			Beginn:     f.von,
			Ende:       f.bis.Add(nachsignierVerzoegerung),
			GrundArt:   tse.StoerungGrundTSEFehler,
			Fehlertext: stoerungFehlertext,
		})
	}
	return zeilen
}

// signaturauftragZeile ist die zu persistierende Zeile der tse_signaturauftraege-Tabelle:
// genau ein Auftrag je fiskalischem Event, die Signatur direkt am Auftrag (NULL bis zur
// Quittierung).
type signaturauftragZeile struct {
	EventID            int
	TxID               string
	ProcessType        string
	ProcessData        string
	Status             string
	Versuche           int
	LetzterFehler      *string
	NaechsterVersuchAm time.Time
	ErstelltAm         time.Time
	ErledigtAm         *time.Time
	Signatur           *tse.Signatur
}

// baueSignaturauftraege spielt Outbox und Signatur-Worker für das Drehbuch nach: Jedes
// fiskalische Event (Entscheidung über die produktive fiskalische Projektion) erhält genau
// eine Auftragszeile mit seiner Event-ID. Im Normalfall gilt der Auftrag als prompt
// quittiert (logTime-Paar aus dem Event-Zeitstempel, erledigt kurz danach). Events in
// aufgelösten Ausfallfenstern werden beim ersten fiskalischen Event nach Fensterende
// nachsigniert — verspätete Signatur ohne Auftrags-Fehlversuche, denn TSE-weite Fehler
// zählen nie auf den Auftrag; einzelne scheitern in der Aufholphase auftragsspezifisch
// und dauerhaft. Events im offenen Fenster der laufenden Sitzung bleiben offen.
// Transaktionsnummern und Signaturzähler sind global streng monoton in
// Quittier-Reihenfolge.
func baueSignaturauftraege(events []seedEvent, fenster []ausfallFenster) ([]signaturauftragZeile, error) {
	s := &fakeSignierer{fenster: fenster, pending: make([][]offeneNachsignierung, len(fenster))}

	for i := range events {
		evt := events[i].event
		vorgang, fiskalisch, err := kasse.FiskalischeProjektion(evt)
		if err != nil {
			return nil, fmt.Errorf("event %s v%d: %w", evt.Subject, evt.Version, err)
		}
		if !fiskalisch {
			continue
		}

		s.nachsigniereFaelligeFenster(evt.Time)

		s.txSeq++
		txID := tseTxID(s.txSeq)

		if f := s.fensterIndex(evt.Time); f >= 0 {
			s.vermerkeAusfall(f, evt.ID, txID, vorgang, evt.Time)
			continue
		}

		signatur := s.signiere(vorgang.ProcessType, vorgang.ProcessData, evt.Time, evt.Time.Add(time.Second), txID)
		erledigt := evt.Time.Add(2 * time.Second)
		s.zeilen = append(s.zeilen, signaturauftragZeile{
			EventID:            evt.ID,
			TxID:               txID,
			ProcessType:        vorgang.ProcessType,
			ProcessData:        vorgang.ProcessData,
			Status:             tse.StatusErledigt,
			NaechsterVersuchAm: evt.Time,
			ErstelltAm:         evt.Time,
			ErledigtAm:         &erledigt,
			Signatur:           &signatur,
		})
	}

	s.nachsigniereAlleFenster()
	return s.zeilen, nil
}

// signaturenNachEventID liefert die quittierten Signaturen je Event-ID — der Leseweg des
// Belegdrucks (Event → Auftrag → Signaturspalten) als Map.
func signaturenNachEventID(auftraege []signaturauftragZeile) map[int]*tse.Signatur {
	signaturen := make(map[int]*tse.Signatur, len(auftraege))
	for i := range auftraege {
		if auftraege[i].Signatur != nil {
			signaturen[auftraege[i].EventID] = auftraege[i].Signatur
		}
	}
	return signaturen
}

// offeneNachsignierung ist ein Ausfall-Vorgang, der nach Fensterende nachsigniert wird.
type offeneNachsignierung struct {
	eventID     int
	txID        string
	processType string
	processData string
	erstellt    time.Time
}

// fakeSignierer hält die globalen TSE-Zähler und sammelt die Auftragszeilen.
type fakeSignierer struct {
	fenster []ausfallFenster
	pending [][]offeneNachsignierung

	txSeq      int // laufende Nummer für die txID-Vergabe (alle fiskalischen Events)
	txNummer   int // TSE-Transaktionsnummer (nur tatsächlich signierte Vorgänge)
	sigZaehler int

	zeilen []signaturauftragZeile
}

// signiere vergibt die nächste Transaktionsnummer samt Signaturzähler und baut die Signatur.
// Jede Transaktion verbraucht zwei Signaturen (Start + Finish) — der Zähler im Beleg ist der
// der Finish-Signatur, wie bei einer echten TSE.
func (s *fakeSignierer) signiere(processType, processData string, logStart, logEnd time.Time, txID string) tse.Signatur {
	s.txNummer++
	s.sigZaehler += 2
	signatur := fakeSignatur(txID, s.txNummer)
	return tse.Signatur{
		TransaktionNummer: s.txNummer,
		SignaturZaehler:   s.sigZaehler,
		TSESeriennummer:   fakeTSESeriennummer,
		LogTimeStart:      logStart.UTC(),
		LogTimeEnd:        logEnd.UTC(),
		Signatur:          signatur,
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
// Vorgang in die Warteliste der späteren Nachsignierung, in offenen Fenstern bleibt der
// Auftrag offen und ohne Signatur — so, wie ihn die Outbox beim Einreihen anlegt.
func (s *fakeSignierer) vermerkeAusfall(fensterIdx, eventID int, txID string, vorgang kasse.FiskalischerVorgang, zeit time.Time) {
	if s.fenster[fensterIdx].aufgeloest {
		s.pending[fensterIdx] = append(s.pending[fensterIdx], offeneNachsignierung{
			eventID:     eventID,
			txID:        txID,
			processType: vorgang.ProcessType,
			processData: vorgang.ProcessData,
			erstellt:    zeit,
		})
		return
	}
	s.zeilen = append(s.zeilen, signaturauftragZeile{
		EventID:            eventID,
		TxID:               txID,
		ProcessType:        vorgang.ProcessType,
		ProcessData:        vorgang.ProcessData,
		Status:             tse.StatusOffen,
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

// nachsigniereFenster spielt den Signatur-Worker nach dem Fensterende nach: Die Aufträge
// werden in Batches im Sekundenabstand quittiert; die verspätete Signatur steht direkt am
// Auftrag. Fehlversuche tragen die Aufträge nicht — während der Störung bricht der Worker
// jeden Durchlauf TSE-weit ab, ohne auf die Aufträge zu zählen.
func (s *fakeSignierer) nachsigniereFenster(fensterIdx int) {
	f := s.fenster[fensterIdx]
	for i, p := range s.pending[fensterIdx] {
		if i%fehlschlagJederNte == 3 {
			s.zeilen = append(s.zeilen, s.dauerhaftGescheitert(p, f.bis))
			continue
		}

		quittiert := f.bis.Add(nachsignierVerzoegerung + time.Duration(i)*250*time.Millisecond)

		signatur := s.signiere(p.processType, p.processData, quittiert, quittiert.Add(time.Second), p.txID)
		erledigt := quittiert.Add(2 * time.Second)
		s.zeilen = append(s.zeilen, signaturauftragZeile{
			EventID:            p.eventID,
			TxID:               p.txID,
			ProcessType:        p.processType,
			ProcessData:        p.processData,
			Status:             tse.StatusErledigt,
			NaechsterVersuchAm: quittiert,
			ErstelltAm:         p.erstellt,
			ErledigtAm:         &erledigt,
			Signatur:           &signatur,
		})
	}
	s.pending[fensterIdx] = nil
}

// dauerhaftGescheitert baut den fehlgeschlagenen Auftrag: Die TSE lehnt die Transaktion
// in der Aufholphase auftragsspezifisch ab, der Auftrag durchläuft die Sekunden-Kurve
// (5, 15 s Backoff) und schlägt mit dem dritten Fehlversuch endgültig fehl.
func (s *fakeSignierer) dauerhaftGescheitert(p offeneNachsignierung, fensterEnde time.Time) signaturauftragZeile {
	fehler := signierungAbgelehntFehler
	return signaturauftragZeile{
		EventID:       p.eventID,
		TxID:          p.txID,
		ProcessType:   p.processType,
		ProcessData:   p.processData,
		Status:        tse.StatusFehlgeschlagen,
		Versuche:      tse_repo.MaxSignaturVersuche,
		LetzterFehler: &fehler,
		// Der letzte Fehlversuch fällt rund 20 s nach das Fensterende (5 + 15 s
		// Backoff); die Query schreibt dabei den (nie mehr genutzten) nächsten
		// Versuch weitere 45 s später.
		NaechsterVersuchAm: fensterEnde.Add(65 * time.Second),
		ErstelltAm:         p.erstellt,
	}
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
