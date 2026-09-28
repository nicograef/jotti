//go:build integration

// Outage, re-signing and latency tests of the TSE live suite, on the setup of tse_live_suite_test.go.
package tse_live

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	kassenfuehrungApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// latenzBurstGroesse is the job count of each latency scenario, fixed for reproducible measurements.
const latenzBurstGroesse = 24

// TestTSELiveSuite_AusfallUndNachsignierung breaks the worker's credentials at runtime (fiskaly 401, a TSE-wide error):
// booking must not wait, the outage is logged and passes the close gate, and restored credentials re-sign the open job.
// See docs/compliance.md §3.8.
func TestTSELiveSuite_AusfallUndNachsignierung(t *testing.T) {
	credentials := credentialsOderSkip(t)
	pruefeTestUmgebungOderAbbruch(t, credentials)

	u := setupLiveUmgebung(t, credentials)
	starteWorker(t, u.db)

	ctx := context.Background()
	db := u.db
	tseRepo := tse_repo.NewRepository(db)

	// Await the opening float's signature so the outage starts from a fully signed state.
	ksNr, err := u.kasse.KassensitzungEroeffnen(ctx, u.userID, "test", "Ausfall-Suite", 5000)
	if err != nil {
		t.Fatalf("KassensitzungEroeffnen: %v", err)
	}
	ksSubject := kasse.KassensitzungSubject(ksNr)
	warteAufSignatur(t, db, eventIDByType(t, db, string(kasse.EventTypeKassensitzungEroeffnetV1), ksSubject))

	// Keep TssID/ClientID but corrupt the ApiSecret: the worker rereads the configuration
	// each pass and fails the token fetch with HTTP 401.
	schreibeKonfiguration(t, tseRepo, tse.Credentials{
		ApiKey:    credentials.ApiKey,
		ApiSecret: credentials.ApiSecret + "-ungueltig",
		TssID:     credentials.TssID,
		ClientID:  credentials.ClientID,
	})

	// Booking during the outage must return without waiting for the TSE; the job stays open.
	bestellungID := uuid.NewString()
	inputs := []enrichment.PositionInput{{ProduktID: u.produktID, VarianteID: u.varianteID, Menge: 1}}
	bucheStart := time.Now()
	if err := u.tisch.BestellungAufnehmen(ctx, u.userID, "test", bestellungID, u.tischID, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen waehrend Ausfall: %v", err)
	}
	if dauer := time.Since(bucheStart); dauer > 5*time.Second {
		t.Errorf("Buchen wartete %s auf die TSE — Buchen muss von der Signierung entkoppelt sein", dauer)
	}
	tischSubject := kasse.TischSessionSubject(ksNr, u.tischID)
	ausfallEventID := eventIDByType(t, db, string(kasse.EventTypeBestellungAufgenommenV1), tischSubject)

	// The worker must detect the TSE-wide error and open an outage period.
	warteAufAktiveStoerung(t, db, tse.StoerungGrundTSEFehler)
	if status := auftragStatus(t, db, ausfallEventID); status != "offen" {
		t.Fatalf("Signaturauftrag waehrend Ausfall im Status %q, erwartet offen", status)
	}

	// During a logged outage the open job counts as Ausfall, not ausstehend, so the close gate passes.
	// Only the classification is checked, so the session stays open for re-signing.
	gate := ausfallGateStand(t, u, ksNr)
	if gate.ausstehend != 0 {
		t.Errorf("Gate bei dokumentiertem Ausfall: %d ausstehend, erwartet 0 (Ausfall blockiert nicht)", gate.ausstehend)
	}
	if gate.ausfallReste < 1 {
		t.Errorf("Gate bei dokumentiertem Ausfall: %d Ausfall-Reste, erwartet mindestens 1", gate.ausfallReste)
	}

	// After its outage backoff the worker resumes: the first success closes the outage period
	// and the open job is re-signed.
	schreibeKonfiguration(t, tseRepo, credentials)

	z := warteAufSignatur(t, db, ausfallEventID)
	pruefeSignatur(t, "Nachsignierung nach Ausfall", z, tse.ProcessTypeBestellungV1)

	// The Nachsigniert flag needs a delay above tse.NachsigniertSchwelle (one minute), longer than
	// this outage's backoff, so either status proves the re-signing.
	stand, err := tseRepo.GetSignaturauftragZuEvent(ctx, ausfallEventID)
	if err != nil {
		t.Fatalf("GetSignaturauftragZuEvent: %v", err)
	}
	ergebnis := tse.DetermineSignaturstatus(stand, nil)
	if ergebnis.Status != tse.SignaturstatusVorhanden && ergebnis.Status != tse.SignaturstatusNachsigniert {
		t.Errorf("Signaturstatus nach Nachsignierung: %q, erwartet vorhanden oder nachsigniert", ergebnis.Status)
	}

	pruefeStoerungsprotokoll(t, db)

	if aktiv, err := tseRepo.GetAktiveTSEStoerung(ctx); err != nil {
		t.Fatalf("GetAktiveTSEStoerung: %v", err)
	} else if aktiv != nil {
		t.Errorf("nach Wiederherstellung noch aktive Stoerung: %+v", aktiv)
	}

	// Without an outage a fresh open job must block the close (409).
	pruefeGateBlockiertOhneStoerung(t, u, ksNr)
}

// burstDeckelP95 caps the burst p95 to catch a signing-rate regression; the serial worker drains the burst one job
// at a time. Measured values: docs/handbuch.md §3.13.
const burstDeckelP95 = 12 * time.Second

// TestTSELiveSuite_SignaturLatenz logs p50/p95 of the end-to-end signing time for single jobs and for a burst.
// Single jobs must hold the p95 < 5 s target of docs/verfahrensdokumentation.md §4.
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

	// Each job is awaited before the next, so no backlog inflates the round trip.
	regelDauern := make([]time.Duration, 0, latenzBurstGroesse)
	for range latenzBurstGroesse {
		id := bucheDirektverkauf(t, u, ksNr)
		warteAufSignatur(t, db, id)
		regelDauern = append(regelDauern, signierDauer(t, db, id))
	}
	regelP50 := perzentil(regelDauern, 0.50)
	regelP95 := perzentil(regelDauern, 0.95)
	t.Logf("TSE-Signaturlatenz Regelbetrieb (einzeln, n=%d): p50=%dms p95=%dms",
		len(regelDauern), regelP50.Milliseconds(), regelP95.Milliseconds())

	if regelP95 > 5*time.Second {
		t.Errorf("Regelbetrieb-p95 %s verletzt die Zusage < 5 s (Verfahrensdokumentation)", regelP95)
	}

	// The burst tail measures the queue depth of the serial worker.
	burstStart := time.Now()
	burstIDs := make([]int, 0, latenzBurstGroesse)
	for range latenzBurstGroesse {
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

	if burstP95 > burstDeckelP95 {
		t.Errorf("Burst-p95 %s ueberschreitet den Deckel %s — Signierrate-Regression?", burstP95, burstDeckelP95)
	}
}

func bucheDirektverkauf(t *testing.T, u *liveTestUmgebung, ksNr int) int {
	t.Helper()
	verkaufID := uuid.NewString()
	verkaufInputs := []enrichment.PositionInput{{ProduktID: u.produktID, VarianteID: u.varianteID, Menge: 1}}
	if err := u.direkt.DirektverkaufTaetigen(context.Background(), u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	return eventIDByType(t, u.db, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject)
}

// schreibeKonfiguration takes effect without a restart because the worker rereads the configuration each pass.
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

// warteAufAktiveStoerung polls until an outage period of grundArt is active.
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

func auftragStatus(t *testing.T, db *sql.DB, eventID int) string {
	t.Helper()
	var status string
	if err := db.QueryRow("SELECT status FROM tse_signaturauftraege WHERE event_id = $1", eventID).Scan(&status); err != nil {
		t.Fatalf("Auftragsstatus zu Event %d lesen: %v", eventID, err)
	}
	return status
}

type gateStand struct {
	ausstehend   int
	ausfallReste int
}

// ausfallGateStand classifies the open jobs with the close gate's logic without running the close.
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

// pruefeGateBlockiertOhneStoerung asserts that a fresh open job without an outage blocks the real close
// with *SignaturenAusstehendError (409).
func pruefeGateBlockiertOhneStoerung(t *testing.T, u *liveTestUmgebung, ksNr int) {
	t.Helper()
	ctx := context.Background()

	// The close must run before the worker signs the fresh job.
	verkaufID := uuid.NewString()
	verkaufInputs := []enrichment.PositionInput{{ProduktID: u.produktID, VarianteID: u.varianteID, Menge: 1}}
	if err := u.direkt.DirektverkaufTaetigen(ctx, u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen fuer Gate-Blockade: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	verkaufEventID := eventIDByType(t, u.db, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject)

	_, err := u.kasse.KasseAbschliessen(ctx, u.userID, "test", 5000)
	var ausstehend *kassenfuehrungApp.SignaturenAusstehendError
	if !errors.As(err, &ausstehend) {
		// A worker that won the race hides the block without a gate fault.
		// A missing block while the job is still open is a real bug.
		if status := auftragStatus(t, u.db, verkaufEventID); status == "offen" {
			t.Fatalf("Abschluss trotz offenem Auftrag ohne Stoerung nicht blockiert: %v", err)
		}
		t.Logf("Gate-Blockade nicht beobachtet: Auftrag bereits signiert, bevor der Abschluss lief")
		return
	}
	if ausstehend.Anzahl < 1 {
		t.Errorf("SignaturenAusstehendError.Anzahl = %d, erwartet mindestens 1", ausstehend.Anzahl)
	}

	// Await the job so the session is closable again.
	warteAufSignatur(t, u.db, verkaufEventID)
}

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
	t.Error("Stoerungsprotokoll enthaelt keinen tse_fehler-Zeitraum")
}

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

func signierDauern(t *testing.T, db *sql.DB, eventIDs []int) []time.Duration {
	t.Helper()
	dauern := make([]time.Duration, 0, len(eventIDs))
	for _, id := range eventIDs {
		dauern = append(dauern, signierDauer(t, db, id))
	}
	return dauern
}

// perzentil uses the nearest-rank method without interpolation, p in 0..1.
func perzentil(dauern []time.Duration, p float64) time.Duration {
	if len(dauern) == 0 {
		return 0
	}
	sortiert := make([]time.Duration, len(dauern))
	copy(sortiert, dauern)
	slices.Sort(sortiert)

	rang := int(p * float64(len(sortiert)))
	if rang >= len(sortiert) {
		rang = len(sortiert) - 1
	}
	return sortiert[rang]
}
