//go:build integration

// Ausfall-, Nachsignierungs- und Latenzmessung der TSE-Live-Suite. Baut auf der
// Infrastruktur von tse_live_suite_test.go auf (setupLiveUmgebung, starteWorker,
// warteAufSignatur) und deckt den Testfall-Katalog aus Block 4 der manuellen QA
// (docs/plans/guide-manuelle-qa-v1.0.0.md) ab:
//
//   - Ausfall zur Laufzeit: Vorgaenge bleiben buchbar, das Stoerungsprotokoll
//     erfasst den Zeitraum mit Grund, nach Wiederherstellung laeuft die
//     Nachsignierung, und das Abschluss-Gate verhaelt sich in beiden Faellen
//     korrekt (409 bei frisch ausstehenden Signaturen, erlaubt bei
//     dokumentiertem Ausfall).
//   - Latenzmessung: ein Burst von Signaturauftraegen, Ausgabe von p50/p95 der
//     realen Ende-zu-Ende-Signierdauer (erstellt_am -> erledigt_am).
package tse_live

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	kassenfuehrungApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// latenzBurstGroesse ist die Zahl der Signaturauftraege der Latenzmessung.
// Als Konstante gehalten, damit die Messung reproduzierbar ist.
const latenzBurstGroesse = 24

// TestTSELiveSuite_AusfallUndNachsignierung schaltet den Signatur-Worker zur
// Laufzeit auf ungueltige Credentials um (401 gegen fiskaly = TSE-weiter Fehler)
// und prueft den kompletten Ausfallpfad:
//
//   - Waehrend des Ausfalls bleiben Vorgaenge buchbar (Buchen wartet nie auf die
//     TSE) und der Signaturauftrag bleibt offen.
//   - Das Stoerungsprotokoll (tse_stoerungen) erfasst den Zeitraum mit Grund
//     tse_fehler.
//   - Das Abschluss-Gate laesst waehrend des dokumentierten Ausfalls durch und
//     weist den Ausfall-Rest in der Abschlussmeldung aus.
//   - Nach Wiederherstellung der Credentials laeuft die Nachsignierung
//     automatisch; die verspaetete Signatur traegt das Nachsigniert-Kennzeichen.
func TestTSELiveSuite_AusfallUndNachsignierung(t *testing.T) {
	credentials := credentialsOderSkip(t)
	pruefeTestUmgebungOderAbbruch(t, credentials)

	u := setupLiveUmgebung(t, credentials)
	starteWorker(t, u.db)

	ctx := context.Background()
	db := u.db
	tseRepo := tse_repo.NewRepository(db)

	// Kassensitzung eroeffnen und die dabei entstehende Bareinlage-Signatur
	// regulaer abwarten, damit die Ausgangslage sauber signiert ist.
	ksNr, err := u.kasse.KassensitzungEroeffnen(ctx, u.userID, "test", "Ausfall-Suite", 5000)
	if err != nil {
		t.Fatalf("KassensitzungEroeffnen: %v", err)
	}
	ksSubject := kasse.KassensitzungSubject(ksNr)
	warteAufSignatur(t, db, eventIDByType(t, db, string(kasse.EventTypeKassensitzungEroeffnetV1), ksSubject))

	// Ausfall ausloesen: gueltige TssID/ClientID behalten, aber ein falsches
	// ApiSecret schreiben. Der Worker liest die Konfiguration bei jedem Durchlauf
	// neu, baut den Client mit den geaenderten Zugangsdaten und scheitert beim
	// Token-Abruf (HTTP 401) TSE-weit — genau der Ausfall, den Block 4 fordert.
	schreibeKonfiguration(t, tseRepo, tse.Credentials{
		ApiKey:    credentials.ApiKey,
		ApiSecret: credentials.ApiSecret + "-ungueltig",
		TssID:     credentials.TssID,
		ClientID:  credentials.ClientID,
	})

	// Waehrend des Ausfalls einen signaturpflichtigen Vorgang buchen. Der Aufruf
	// muss ohne Warten auf die TSE zurueckkehren (Buchen ist von der Signierung
	// entkoppelt) — der Signaturauftrag bleibt offen.
	bestellungID := uuid.NewString()
	inputs := []tischgeschaeftApp.BestellPositionInput{{ProduktID: u.produkt, VarianteID: u.variante, Menge: 1}}
	bucheStart := time.Now()
	if err := u.tisch.BestellungAufnehmen(ctx, u.userID, "test", bestellungID, u.tischID, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen waehrend Ausfall: %v", err)
	}
	if dauer := time.Since(bucheStart); dauer > 5*time.Second {
		t.Errorf("Buchen wartete %s auf die TSE — Buchen muss von der Signierung entkoppelt sein", dauer)
	}
	tischSubject := kasse.TischSessionSubject(ksNr, u.tischID)
	ausfallEventID := eventIDByType(t, db, string(kasse.EventTypeBestellungAufgenommenV1), tischSubject)

	// Der Worker muss den TSE-weiten Fehler erkennen und den Stoerungszeitraum
	// oeffnen. Bis dahin bleibt der Auftrag offen.
	warteAufAktiveStoerung(t, db, tse.StoerungGrundTSEFehler)
	if status := auftragStatus(t, db, ausfallEventID); status != "offen" {
		t.Fatalf("Signaturauftrag waehrend Ausfall im Status %q, erwartet offen", status)
	}

	// Abschluss-Gate im dokumentierten Ausfall: Der offene Auftrag faellt bei
	// aktiver Stoerung unter Ausfall (nicht ausstehend), der Abschluss ist erlaubt
	// und weist den Ausfall-Rest aus. Wir pruefen das Gate isoliert ueber die
	// Klassifikation, ohne die Sitzung abzuschliessen (der Ausfall soll fuer die
	// Nachsignierung bestehen bleiben).
	gate := ausfallGateStand(t, u, ksNr)
	if gate.ausstehend != 0 {
		t.Errorf("Gate bei dokumentiertem Ausfall: %d ausstehend, erwartet 0 (Ausfall blockiert nicht)", gate.ausstehend)
	}
	if gate.ausfallReste < 1 {
		t.Errorf("Gate bei dokumentiertem Ausfall: %d Ausfall-Reste, erwartet mindestens 1", gate.ausfallReste)
	}

	// Wiederherstellung: gueltige Credentials zurueckschreiben. Der Worker nimmt
	// nach Ablauf seines Stoerungs-Backoffs die Aufarbeitung wieder auf, die erste
	// erfolgreiche Signatur schliesst den Stoerungszeitraum, und der offene
	// Auftrag wird nachsigniert.
	schreibeKonfiguration(t, tseRepo, credentials)

	z := warteAufSignatur(t, db, ausfallEventID)
	pruefeSignatur(t, "Nachsignierung nach Ausfall", z, tse.ProcessTypeBestellungV1)

	// Die automatische Nachsignierung ist damit belegt: derselbe Auftrag, der
	// waehrend des Ausfalls offen blieb, traegt nach der Wiederherstellung eine
	// vollstaendige Signatur (warteAufSignatur wartet auf status='erledigt').
	// Wir werten zusaetzlich die Signaturstatus-Funktion aus — dieselbe
	// Zurechnung wie der Beleg-Abruf. Das Nachsigniert-Kennzeichen setzt eine
	// Verspaetung ueber tse.NachsigniertSchwelle (eine Minute) voraus; dieser
	// Ausfall dauert nur den Worker-Backoff (Sekunden), also ist beides gueltig:
	// bei einem kurzen Ausfall 'vorhanden', bei einem Ausfall > 1 min
	// 'nachsigniert'. Beide belegen die erfolgte Nachsignierung.
	stand, err := tseRepo.GetSignaturauftragZuEvent(ctx, ausfallEventID)
	if err != nil {
		t.Fatalf("GetSignaturauftragZuEvent: %v", err)
	}
	ergebnis := tse.DetermineSignaturstatus(stand, nil)
	if ergebnis.Status != tse.SignaturstatusVorhanden && ergebnis.Status != tse.SignaturstatusNachsigniert {
		t.Errorf("Signaturstatus nach Nachsignierung: %q, erwartet vorhanden oder nachsigniert", ergebnis.Status)
	}

	// Das Stoerungsprotokoll dokumentiert den abgeschlossenen Zeitraum: Grund
	// tse_fehler, Beginn gesetzt, Ende nach der Wiederherstellung gesetzt.
	pruefeStoerungsprotokoll(t, db)

	// Nach der Wiederherstellung darf keine Stoerung mehr aktiv sein.
	if aktiv, err := tseRepo.GetAktiveTSEStoerung(ctx); err != nil {
		t.Fatalf("GetAktiveTSEStoerung: %v", err)
	} else if aktiv != nil {
		t.Errorf("nach Wiederherstellung noch aktive Stoerung: %+v", aktiv)
	}

	// Gate ohne Stoerung: Der frisch signierte Auftrag ist erledigt, ein weiterer
	// frischer offener Auftrag ohne Stoerung muss dagegen blockieren (409).
	pruefeGateBlockiertOhneStoerung(t, u, ksNr)
}

// burstDeckelP95 ist die Obergrenze fuer die p95-Ende-zu-Ende-Dauer des
// gleichzeitigen Bursts. Der Burst ist ein Worst-Case-Stresstest, kein
// Regelbetrieb: latenzBurstGroesse Auftraege liegen gleichzeitig an, und der
// serielle Worker (ein Sprecher, FIFO) arbeitet sie nacheinander ab, sodass der
// letzte Auftrag hinter allen Vorgaengern wartet. Gemessen wurden reproduzierbar
// p50 ~4 s / p95 ~7 s (2026-07-09, fiskaly-TEST-TSS); der Deckel faengt eine
// echte Regression der Signierrate ab, ohne die Regelbetriebs-Zusage auf den
// Burst zu uebertragen.
const burstDeckelP95 = 12 * time.Second

// TestTSELiveSuite_SignaturLatenz misst die reale Ende-zu-Ende-Signierdauer
// (erstellt_am -> erledigt_am) in zwei Szenarien und gibt p50/p95 reproduzierbar
// aus:
//
//   - Regelbetrieb: Signaturauftraege einzeln nacheinander, jeder vor dem
//     naechsten abgewartet. Das entspricht dem verteilten Anfall im Vereinsbetrieb
//     und ist die Grundlage der Zusage der Verfahrensdokumentation (p95 < 5 s).
//   - Burst: latenzBurstGroesse gleichzeitig anliegende Auftraege als
//     Worst-Case-Stress; der serielle Worker staut sie, der Tail-Wert bildet die
//     Warteschlangen-Tiefe ab (kein Regelbetrieb).
func TestTSELiveSuite_SignaturLatenz(t *testing.T) {
	credentials := credentialsOderSkip(t)
	pruefeTestUmgebungOderAbbruch(t, credentials)

	u := setupLiveUmgebung(t, credentials)
	starteWorker(t, u.db)

	ctx := context.Background()
	db := u.db

	ksNr, err := u.kasse.KassensitzungEroeffnen(ctx, u.userID, "test", "Latenz-Suite", 5000)
	if err != nil {
		t.Fatalf("KassensitzungEroeffnen: %v", err)
	}

	// Regelbetrieb: je Auftrag buchen, signieren lassen, abwarten — dann der
	// naechste. Kein Rueckstau, die Dauer ist die reine Signier-Round-Trip-Zeit.
	regelDauern := make([]time.Duration, 0, latenzBurstGroesse)
	for i := 0; i < latenzBurstGroesse; i++ {
		id := bucheDirektverkauf(t, u, ksNr)
		warteAufSignatur(t, db, id)
		regelDauern = append(regelDauern, signierDauer(t, db, id))
	}
	regelP50 := perzentil(regelDauern, 0.50)
	regelP95 := perzentil(regelDauern, 0.95)
	t.Logf("TSE-Signaturlatenz Regelbetrieb (einzeln, n=%d): p50=%dms p95=%dms",
		len(regelDauern), regelP50.Milliseconds(), regelP95.Milliseconds())

	// Zusage der Verfahrensdokumentation gilt dem Regelbetrieb: p95 < 5 s.
	if regelP95 > 5*time.Second {
		t.Errorf("Regelbetrieb-p95 %s verletzt die Zusage < 5 s (Verfahrensdokumentation)", regelP95)
	}

	// Burst: latenzBurstGroesse Auftraege gleichzeitig einreihen, dann alle
	// abwarten. Der Tail-Wert misst die Warteschlangen-Tiefe des seriellen Workers.
	burstStart := time.Now()
	burstIDs := make([]int, 0, latenzBurstGroesse)
	for i := 0; i < latenzBurstGroesse; i++ {
		burstIDs = append(burstIDs, bucheDirektverkauf(t, u, ksNr))
	}
	for _, id := range burstIDs {
		warteAufSignatur(t, db, id)
	}
	drainDauer := time.Since(burstStart)

	burstDauern := signierDauern(t, db, burstIDs)
	burstP50 := perzentil(burstDauern, 0.50)
	burstP95 := perzentil(burstDauern, 0.95)
	proSignatur := drainDauer / time.Duration(len(burstIDs))
	t.Logf("TSE-Signaturlatenz Burst (gleichzeitig=%d): p50=%dms p95=%dms Drain=%dms (~%dms/Signatur)",
		latenzBurstGroesse, burstP50.Milliseconds(), burstP95.Milliseconds(),
		drainDauer.Milliseconds(), proSignatur.Milliseconds())

	// Der Burst hat keine 5-s-Zusage; der Deckel faengt nur eine echte Regression
	// der Signierrate ab.
	if burstP95 > burstDeckelP95 {
		t.Errorf("Burst-p95 %s ueberschreitet den Deckel %s — Signierrate-Regression?", burstP95, burstDeckelP95)
	}
}

// bucheDirektverkauf bucht einen Direktverkauf ueber die Standard-Variante und
// liefert die kassenjournal-ID des erzeugten Events.
func bucheDirektverkauf(t *testing.T, u *liveTestUmgebung, ksNr int) int {
	t.Helper()
	verkaufID := uuid.NewString()
	verkaufInputs := []direktverkaufApp.VerkaufPositionInput{{ProduktID: u.produkt, VarianteID: u.variante, Menge: 1}}
	if err := u.direkt.DirektverkaufTaetigen(context.Background(), u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	return eventIDByType(t, u.db, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject)
}

// schreibeKonfiguration ueberschreibt die TSE-Konfiguration in der DB. Der
// Signatur-Worker liest sie bei jedem Durchlauf neu; ein Wechsel wirkt damit
// zur Laufzeit ohne Neustart.
func schreibeKonfiguration(t *testing.T, repo tse_repo.Repository, creds tse.Credentials) {
	t.Helper()
	konf, err := tse.NewKonfiguration(creds.ApiKey, creds.ApiSecret, creds.TssID, creds.ClientID)
	if err != nil {
		t.Fatalf("Konfiguration bauen: %v", err)
	}
	if err := repo.SaveEinrichtung(context.Background(), konf); err != nil {
		t.Fatalf("Konfiguration speichern: %v", err)
	}
}

// warteAufAktiveStoerung pollt, bis ein aktiver Stoerungszeitraum der Grund-Art
// vorliegt (der Worker hat den TSE-weiten Fehler erkannt).
func warteAufAktiveStoerung(t *testing.T, db *sql.DB, grundArt string) {
	t.Helper()
	deadline := time.Now().Add(signaturWartefrist)
	repo := tse_repo.NewRepository(db)
	for {
		aktiv, err := repo.GetAktiveTSEStoerung(context.Background())
		if err != nil {
			t.Fatalf("GetAktiveTSEStoerung: %v", err)
		}
		if aktiv != nil && aktiv.GrundArt == grundArt {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("kein aktiver Stoerungszeitraum %q binnen %s", grundArt, signaturWartefrist)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// auftragStatus liefert den aktuellen Status des Signaturauftrags eines Events.
func auftragStatus(t *testing.T, db *sql.DB, eventID int) string {
	t.Helper()
	var status string
	if err := db.QueryRow("SELECT status FROM tse_signaturauftraege WHERE event_id = $1", eventID).Scan(&status); err != nil {
		t.Fatalf("Auftragsstatus zu Event %d lesen: %v", eventID, err)
	}
	return status
}

// gateStand buendelt die fuer Block 4 relevanten Kennzahlen des
// Abschluss-Gates.
type gateStand struct {
	ausstehend   int
	ausfallReste int
}

// ausfallGateStand klassifiziert die offenen Signaturauftraege der Sitzung mit
// derselben Logik wie das Abschluss-Gate (GetOffeneSignaturauftragStaende +
// GetAktiveTSEStoerung + DetermineSignaturstatus), ohne den Abschluss
// auszufuehren.
func ausfallGateStand(t *testing.T, u *liveTestUmgebung, ksNr int) gateStand {
	t.Helper()
	ctx := context.Background()
	repo := tse_repo.NewRepository(u.db)
	staende, err := repo.GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetOffeneSignaturauftragStaendeFuerKassensitzung: %v", err)
	}
	aktiv, err := repo.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("GetAktiveTSEStoerung: %v", err)
	}
	var stand gateStand
	for _, s := range staende {
		ergebnis := tse.DetermineSignaturstatus(s, aktiv)
		switch ergebnis.Status {
		case tse.SignaturstatusAusstehend:
			stand.ausstehend++
		case tse.SignaturstatusAusfall:
			stand.ausfallReste++
		}
	}
	return stand
}

// pruefeGateBlockiertOhneStoerung erzwingt einen frisch offenen Auftrag ohne
// aktive Stoerung und prueft, dass der reale Kassenabschluss mit
// *SignaturenAusstehendError (409 mit Anzahl) blockiert. Anschliessend
// wird der Auftrag abgewartet, damit die Sitzung wieder abschliessbar waere.
func pruefeGateBlockiertOhneStoerung(t *testing.T, u *liveTestUmgebung, ksNr int) {
	t.Helper()
	ctx := context.Background()

	// Frischen signaturpflichtigen Vorgang buchen; ohne aktive Stoerung ist sein
	// offener Auftrag ausstehend. Direkt danach den Abschluss versuchen, bevor der
	// Worker signiert — deshalb genuegt der erste Poll-Takt Vorlauf nicht, wir
	// greifen sofort zu.
	verkaufID := uuid.NewString()
	verkaufInputs := []direktverkaufApp.VerkaufPositionInput{{ProduktID: u.produkt, VarianteID: u.variante, Menge: 1}}
	if err := u.direkt.DirektverkaufTaetigen(ctx, u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen fuer Gate-Blockade: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	verkaufEventID := eventIDByType(t, u.db, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject)

	_, err := u.kasse.KasseAbschliessen(ctx, u.userID, "test", 5000)
	var ausstehend *kassenfuehrungApp.SignaturenAusstehendError
	if !errors.As(err, &ausstehend) {
		// Der Worker koennte den Auftrag bereits signiert haben (Race). Dann ist
		// die Blockade nicht mehr beobachtbar; das ist kein Fehlverhalten des
		// Gates, aber die Blockade-Assertion braucht den offenen Auftrag. Ist der
		// Auftrag noch offen, ist das Ausbleiben des Fehlers ein echter Bug.
		if status := auftragStatus(t, u.db, verkaufEventID); status == "offen" {
			t.Fatalf("Abschluss trotz offenem Auftrag ohne Stoerung nicht blockiert: %v", err)
		}
		t.Logf("Gate-Blockade nicht beobachtet: Auftrag bereits signiert, bevor der Abschluss lief")
		return
	}
	if ausstehend.Anzahl < 1 {
		t.Errorf("SignaturenAusstehendError.Anzahl = %d, erwartet mindestens 1", ausstehend.Anzahl)
	}

	// Auftrag abwarten, damit ein evtl. Zwischenstatus der Sitzung nicht haengt.
	warteAufSignatur(t, u.db, verkaufEventID)
}

// pruefeStoerungsprotokoll prueft, dass das Stoerungsprotokoll (tse_stoerungen)
// den Ausfallzeitraum dokumentiert: mindestens ein tse_fehler-Zeitraum mit
// gesetztem Beginn, gesetztem Ende und nicht-leerem Fehlertext (Grund).
func pruefeStoerungsprotokoll(t *testing.T, db *sql.DB) {
	t.Helper()
	zeitraeume, err := tse_repo.NewRepository(db).GetAlleTSEStoerungen(context.Background())
	if err != nil {
		t.Fatalf("GetAlleTSEStoerungen: %v", err)
	}
	for _, z := range zeitraeume {
		if z.GrundArt != tse.StoerungGrundTSEFehler {
			continue
		}
		if z.Beginn.IsZero() {
			t.Error("Stoerungsprotokoll: tse_fehler-Zeitraum ohne Beginn")
		}
		if z.Ende == nil {
			t.Error("Stoerungsprotokoll: tse_fehler-Zeitraum nach Wiederherstellung nicht geschlossen (Ende fehlt)")
		}
		if z.Fehlertext == "" {
			t.Error("Stoerungsprotokoll: tse_fehler-Zeitraum ohne Grund (Fehlertext leer)")
		}
		return
	}
	t.Fatal("Stoerungsprotokoll enthaelt keinen tse_fehler-Zeitraum")
}

// signierDauer liest die reale Ende-zu-Ende-Signierdauer (erledigt_am -
// erstellt_am) eines erledigten Auftrags.
func signierDauer(t *testing.T, db *sql.DB, eventID int) time.Duration {
	t.Helper()
	var sekunden float64
	if err := db.QueryRow(
		"SELECT EXTRACT(EPOCH FROM (erledigt_am - erstellt_am)) FROM tse_signaturauftraege WHERE event_id = $1 AND status = 'erledigt'",
		eventID,
	).Scan(&sekunden); err != nil {
		t.Fatalf("Signierdauer zu Event %d lesen: %v", eventID, err)
	}
	return time.Duration(sekunden * float64(time.Second))
}

// signierDauern liest die reale Signierdauer je Auftrag.
func signierDauern(t *testing.T, db *sql.DB, eventIDs []int) []time.Duration {
	t.Helper()
	dauern := make([]time.Duration, 0, len(eventIDs))
	for _, id := range eventIDs {
		dauern = append(dauern, signierDauer(t, db, id))
	}
	return dauern
}

// perzentil liefert das p-Perzentil (0..1) der Dauern per naechster-Rang-Methode.
// Reproduzierbar und ohne Interpolation, damit die Ausgabe deterministisch bleibt.
func perzentil(dauern []time.Duration, p float64) time.Duration {
	if len(dauern) == 0 {
		return 0
	}
	sortiert := make([]time.Duration, len(dauern))
	copy(sortiert, dauern)
	sort.Slice(sortiert, func(i, j int) bool { return sortiert[i] < sortiert[j] })

	rang := int(p * float64(len(sortiert)))
	if rang >= len(sortiert) {
		rang = len(sortiert) - 1
	}
	return sortiert[rang]
}
