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

// Fixed, invented fake-TSE identity for the demo scenario in the formats of a real fiskaly Cloud-TSE.
// Serial number is SHA-256 hex; key, certificate and signatures are Base64.
const (
	fakeTSESeriennummer     = "9c4f2d8a71e3b65042dca9f01b87e6d355a1c0fb29e84d7613f5a2b8c90e4761"
	fakeKassenSeriennummer  = "JOTTI-DEMO-KASSE-1"
	fakeSignaturAlgorithmus = "ecdsa-plain-SHA256"
	fakeLogTimeFormat       = "unixTime"
	fakePublicKey           = "BJottiDemoFakeTSEPublicKeySommerfestTSVMusterstadt2026AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	// fakeZertifikat stays under 1000 characters to fit TSE_ZERTIFIKAT_I.
	// It completes the tse.csv master data so the DSFinV-K content check sees them as complete.
	fakeZertifikat = "MIIBdummyJottiDemoFakeTSECertificateBase64AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
)

// fakeTSEStammdaten are written like a real TSE setup for tse.csv in the DSFinV-K export.
// Without them the tse_stammdaten singleton keeps its empty migration default and tse.csv lacks mandatory fields.
func fakeTSEStammdaten() tse.Stammdaten {
	return tse.Stammdaten{
		Seriennummer:        fakeTSESeriennummer,
		SignaturAlgorithmus: fakeSignaturAlgorithmus,
		PublicKey:           fakePublicKey,
		Zertifikat:          fakeZertifikat,
		LogTimeFormat:       fakeLogTimeFormat,
	}
}

// signierungAbgelehntFehler is the error of permanently failed jobs: unlike a window outage, the TSE rejects them after it too.
const signierungAbgelehntFehler = "Cloud-TSE lehnt die Transaktion dauerhaft ab (HTTP 400: ungültige process_data)"

// fehlschlagJederNte steuert die Dramaturgie aufgelöster Ausfallfenster: Jeder 16. Auftrag
// (ab dem vierten) scheitert dauerhaft und bleibt fehlgeschlagen.
const fehlschlagJederNte = 16

// nachsignierVerzoegerung separates the window end from the first successful re-signing.
// It also ends the seeded Störungszeitraum, since in production the first successful signature closes it.
const nachsignierVerzoegerung = 5 * time.Second

// stoerungFehlertext ist der Fehlertext der geseedeten tse_fehler-Störungszeiträume.
const stoerungFehlertext = "Cloud-TSE nicht erreichbar (HTTP 503)"

// ausfallFenster is a TSE outage window with absolute times.
// aufgeloest marks its jobs as re-signed by the worker (closed session) instead of left open (open session).
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

// stoerungszeitraeumeAus turns the resolved outage windows into the closed tse_fehler periods production would have logged.
// The open window of the running session is left out: worker and watchdog create it live after app start.
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

// buildSignaturauftraege replays outbox and signature worker: each fiscal event gets exactly one job, normally acknowledged promptly.
// Resolved outage windows are re-signed after the window end; see docs/handbuch.md §3.13.
func buildSignaturauftraege(events []seedEvent, fenster []ausfallFenster) ([]signaturauftragZeile, error) {
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

// signiere assigns the next transaction number and signature counter and builds the signature.
// Each transaction consumes two signatures (start and finish); the receipt shows the finish counter, as with a real TSE.
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

// vermerkeAusfall queues an event of a resolved window for later re-signing.
// In an open window the job stays open and unsigned, as the outbox enqueues it.
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

// nachsigniereFenster replays the worker after the window end, acknowledging jobs in sub-second steps.
// Jobs carry no failed attempts: during the outage the worker aborts TSE-wide without counting on jobs.
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

// dauerhaftGescheitert builds a job the TSE rejects job-specifically during catch-up.
// It fails for good on the third attempt after 5 s and 15 s backoff.
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
		// The last attempt lands about 20 s after the window end; the query then sets the unused next attempt 45 s later.
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

// qrCodeData baut den KassenSichV-üblichen V0-String (DSFinV-K Anhang I), wie ihn fiskaly
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
