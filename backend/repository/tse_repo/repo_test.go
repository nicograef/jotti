//go:build integration

package tse_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// testUmgebung holds the test DB with Kassensitzung and user; each Signaturauftrag needs its own
// Kassenjournal event (event_id NOT NULL UNIQUE).
type testUmgebung struct {
	db      *sql.DB
	userID  int
	ksNr    int
	version int
}

func setupRepository(t *testing.T) (Repository, *testUmgebung, func(t *testing.T)) {
	t.Helper()
	database := dbtest.Open()

	reset := func(t *testing.T) {
		t.Helper()
		stmts := []string{
			"DELETE FROM tse_stoerungen",
			"DELETE FROM tse_signaturauftraege",
			"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassenjournal",
			"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassensitzungen",
			"DELETE FROM users",
		}
		for _, stmt := range stmts {
			if _, err := database.Exec(stmt); err != nil {
				t.Fatalf("reset %q: %v", stmt, err)
			}
		}
	}
	reset(t)

	umgebung := &testUmgebung{db: database}
	if err := database.QueryRow(
		"INSERT INTO users (name, username, role, status, created_at, updated_at) VALUES ('Test', 'tse-repo-test', 'admin', 'active', NOW(), NOW()) RETURNING id",
	).Scan(&umgebung.userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := database.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&umgebung.ksNr); err != nil {
		t.Fatalf("insert kassensitzung: %v", err)
	}

	return NewRepository(database), umgebung, func(t *testing.T) {
		reset(t)
		_ = database.Close()
	}
}

// insertAuftrag creates a Kassenjournal event of the default session with an open Signaturauftrag and returns (auftragID, eventID).
func (u *testUmgebung) insertAuftrag(t *testing.T, txID string) (int, int) {
	t.Helper()
	return u.insertAuftragFuerSitzung(t, txID, u.ksNr)
}

// insertAuftragFuerSitzung creates an event with an open Signaturauftrag for the given Kassensitzung.
func (u *testUmgebung) insertAuftragFuerSitzung(t *testing.T, txID string, ksNr int) (int, int) {
	t.Helper()
	u.version++
	var eventID int
	if err := u.db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'Test', 'zahlung-kassiert:v1', $2, $3, '{}', NOW(), $4) RETURNING id",
		u.userID, fmt.Sprintf("kassensitzung-%d/tisch-1", ksNr), u.version, ksNr,
	).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	var auftragID int
	if err := u.db.QueryRow(`
		INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
		VALUES ($1, $2, 'Kassenbeleg-V1', 'Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar', 'offen', NOW(), NOW())
		RETURNING id
	`, eventID, txID).Scan(&auftragID); err != nil {
		t.Fatalf("insert auftrag: %v", err)
	}
	return auftragID, eventID
}

// insertKassensitzung opens another Kassensitzung and returns its z_nr.
// idx_kassensitzungen_eine_aktiv allows one active session only, so close the previous one first.
func (u *testUmgebung) insertKassensitzung(t *testing.T) int {
	t.Helper()
	var nr int
	if err := u.db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&nr); err != nil {
		t.Fatalf("insert kassensitzung: %v", err)
	}
	return nr
}

// closeKassensitzung sets the session abgeschlossen, the status change the Tagesabschluss event causes.
func (u *testUmgebung) closeKassensitzung(t *testing.T, ksNr int) {
	t.Helper()
	if _, err := u.db.Exec(
		"UPDATE kassensitzungen SET status = 'abgeschlossen', updated_at = NOW() WHERE z_nr = $1", ksNr,
	); err != nil {
		t.Fatalf("close kassensitzung: %v", err)
	}
}

// markiereFehlgeschlagen fails an order MaxSignaturVersuche times, leaving it fehlgeschlagen with letzter_fehler set.
func markiereFehlgeschlagen(ctx context.Context, t *testing.T, store Repository, auftragID int, fehler string) {
	t.Helper()
	for i := range MaxSignaturVersuche {
		if err := store.TSESignaturauftragFehlversuch(ctx, auftragID, fehler); err != nil {
			t.Fatalf("Fehlversuch %d: %v", i, err)
		}
	}
}

func testSignatur(txNr int) tse.Signatur {
	return tse.Signatur{
		TransaktionNummer: txNr,
		SignaturZaehler:   txNr + 1,
		TSESeriennummer:   "TSE-SN-1",
		LogTimeStart:      time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
		LogTimeEnd:        time.Date(2026, 6, 11, 12, 0, 1, 0, time.UTC),
		Signatur:          "SIG-1",
		QRCodeData:        "V0;QR",
	}
}

// Quittierung fills the signature columns exactly once (status guard offen): the order becomes erledigt,
// the Beleg reads the signature from it, and a second Quittierung changes nothing.
func TestQuittiereTSESignaturauftrag_EinzelUpdateMitStatusGuard(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	auftragID, eventID := umgebung.insertAuftrag(t, "tx-quittierung")

	if err := store.QuittiereTSESignaturauftrag(ctx, auftragID, testSignatur(41)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}

	stand, err := store.GetSignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if stand.Status != tse.StatusErledigt {
		t.Errorf("Expected status erledigt, got %q", stand.Status)
	}
	if stand.Signatur == nil || stand.Signatur.TransaktionNummer != 41 || stand.Signatur.Signatur != "SIG-1" {
		t.Fatalf("Expected quittierte signatur at auftrag, got %+v", stand.Signatur)
	}

	// Erledigt orders are not due.
	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Errorf("Expected no due auftraege after quittierung, got %+v", offene)
	}

	// A second Quittierung is a no-op (signature columns written exactly once).
	if err := store.QuittiereTSESignaturauftrag(ctx, auftragID, testSignatur(99)); err != nil {
		t.Errorf("Expected no error from repeated quittierung, got %v", err)
	}
	stand, err = store.GetSignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if stand.Signatur.TransaktionNummer != 41 {
		t.Errorf("Expected signature to stay at 41, got %d", stand.Signatur.TransaktionNummer)
	}
}

// GetSignaturauftragZuEvent: no order (not signaturpflichtig) -> db.ErrNotFound; open order -> state without signature.
func TestGetSignaturauftragZuEvent_Faelle(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	if _, err := store.GetSignaturauftragZuEvent(ctx, 999999); !errors.Is(err, dbpkg.ErrNotFound) {
		t.Errorf("Expected db.ErrNotFound for event without auftrag, got %v", err)
	}

	_, eventID := umgebung.insertAuftrag(t, "tx-offen-stand")
	stand, err := store.GetSignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if stand.Status != tse.StatusOffen || stand.Signatur != nil {
		t.Errorf("Expected offenen stand ohne signatur, got %+v", stand)
	}
	if stand.ErstelltAm.IsZero() {
		t.Error("Expected erstellt_am to be set")
	}
}

// A failure moves the next attempt into the future (backoff): the failing order leaves the worker batch
// while a newer order stays fetchable, so there is no head-of-line blocking.
func TestTSESignaturauftragFehlversuch_BackoffBlockiertNeuereNicht(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	fehlschlagendID, _ := umgebung.insertAuftrag(t, "tx-fehlschlagend")
	neuererID, _ := umgebung.insertAuftrag(t, "tx-neuer")

	if err := store.TSESignaturauftragFehlversuch(ctx, fehlschlagendID, "fiskaly timeout"); err != nil {
		t.Fatalf("Expected no fehlversuch error, got %v", err)
	}

	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != neuererID {
		t.Errorf("Expected only the newer auftrag to be due, got %+v", offene)
	}

	// The failing order recorded the failure and stays open.
	status, versuche, letzterFehler := auftragStatus(t, umgebung.db, fehlschlagendID)
	if status != "offen" || versuche != 1 || letzterFehler != "fiskaly timeout" {
		t.Errorf("Expected recorded fehlversuch, got status=%q versuche=%d fehler=%q", status, versuche, letzterFehler)
	}
}

// The order-specific failure curve runs in seconds (5 s, 15 s backoff, third failure final), far below the Rückstand threshold.
// A poison order never opens a Rückstand Störungszeitraum, leaves the backlog measurement, and stays ausstehend until final.
func TestTSESignaturauftragFehlversuch_SekundenKurveEndetVorRueckstandsSchwelle(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	id, eventID := umgebung.insertAuftrag(t, "tx-gift")
	erwarteteBackoffs := []time.Duration{5 * time.Second, 15 * time.Second}

	var gesamtBackoff time.Duration
	for versuch := 1; versuch < MaxSignaturVersuche; versuch++ {
		if err := store.TSESignaturauftragFehlversuch(ctx, id, "fiskaly api error 400 (E_FAILED_SCHEMA_VALIDATION)"); err != nil {
			t.Fatalf("Fehlversuch %d: %v", versuch, err)
		}

		// The poison order is ausstehend during the curve, never Ausfall.
		stand, err := store.GetSignaturauftragZuEvent(ctx, eventID)
		if err != nil {
			t.Fatalf("Stand nach Fehlversuch %d lesen: %v", versuch, err)
		}
		if ergebnis := tse.DetermineSignaturstatus(stand, nil); ergebnis.Status != tse.SignaturstatusAusstehend {
			t.Errorf("Fehlversuch %d: erwartet ausstehend, got %q", versuch, ergebnis.Status)
		}

		backoff := backoffBis(t, umgebung.db, id)
		erwartet := erwarteteBackoffs[versuch-1]
		if backoff < erwartet-2*time.Second || backoff > erwartet+2*time.Second {
			t.Errorf("Fehlversuch %d: Backoff %v, erwartet ~%v", versuch, backoff, erwartet)
		}
		gesamtBackoff += backoff
	}

	// The MaxSignaturVersuche-th failure makes the order final.
	if err := store.TSESignaturauftragFehlversuch(ctx, id, "fiskaly api error 400 (E_FAILED_SCHEMA_VALIDATION)"); err != nil {
		t.Fatalf("letzter Fehlversuch: %v", err)
	}
	if status, _, _ := auftragStatus(t, umgebung.db, id); status != tse.StatusFehlgeschlagen {
		t.Errorf("Expected endgueltig fehlgeschlagenen Auftrag, got %q", status)
	}

	// The whole curve stays far below the Rückstand threshold, leaving room for tick and processing slack.
	if gesamtBackoff >= tse.RueckstandSchwelle/2 {
		t.Errorf("Backoff-Kurve %v zu nah an der Rueckstands-Schwelle %v", gesamtBackoff, tse.RueckstandSchwelle)
	}

	// Fehlgeschlagen orders are no backlog: the watchdog measures open orders only.
	aeltester, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Rueckstand messen: %v", err)
	}
	if aeltester != nil {
		t.Errorf("Expected fehlgeschlagenen Auftrag nicht in der Rueckstands-Messung, got %v", aeltester)
	}
}

// backoffBis measures the backoff the failure query set (naechster_versuch_am − NOW()).
func backoffBis(t *testing.T, db *sql.DB, auftragID int) time.Duration {
	t.Helper()
	var sekunden float64
	if err := db.QueryRow(
		"SELECT EXTRACT(EPOCH FROM (naechster_versuch_am - NOW())) FROM tse_signaturauftraege WHERE id = $1", auftragID,
	).Scan(&sekunden); err != nil {
		t.Fatalf("Backoff lesen: %v", err)
	}
	return time.Duration(sekunden * float64(time.Second))
}

// Without TSE configuration the worker marks open orders tse_nicht_konfiguriert for good; erledigt orders stay untouched.
func TestMarkOffeneAlsNichtKonfiguriert_MarkiertNurOffene(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	offenID, _ := umgebung.insertAuftrag(t, "tx-ohne-konfig-1")
	zweiterID, _ := umgebung.insertAuftrag(t, "tx-ohne-konfig-2")
	erledigtID, _ := umgebung.insertAuftrag(t, "tx-erledigt")
	if err := store.QuittiereTSESignaturauftrag(ctx, erledigtID, testSignatur(41)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}

	markiert, err := store.MarkOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		t.Fatalf("Expected no mark error, got %v", err)
	}
	if markiert != 2 {
		t.Errorf("Expected 2 marked auftraege (die beiden offenen), got %d", markiert)
	}

	status := statusMap(t, umgebung.db)
	if status[offenID] != tse.StatusTSENichtKonfiguriert || status[zweiterID] != tse.StatusTSENichtKonfiguriert {
		t.Errorf("Expected offene auftraege marked tse_nicht_konfiguriert, got %+v", status)
	}
	if status[erledigtID] != tse.StatusErledigt {
		t.Errorf("Expected erledigten auftrag untouched, got %q", status[erledigtID])
	}

	// Marked orders are not due.
	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Errorf("Expected no due auftraege after marking, got %+v", offene)
	}

	// Marking again without open orders marks nothing.
	markiert, err = store.MarkOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		t.Fatalf("Expected no mark error, got %v", err)
	}
	if markiert != 0 {
		t.Errorf("Expected 0 marked on second run, got %d", markiert)
	}
}

// auftragStatus reads status, attempts and last error of an order via SQL, since no admin read query exists.
func auftragStatus(t *testing.T, db *sql.DB, id int) (status string, versuche int, letzterFehler string) {
	t.Helper()
	var fehler sql.NullString
	if err := db.QueryRow(
		"SELECT status, versuche, letzter_fehler FROM tse_signaturauftraege WHERE id = $1", id,
	).Scan(&status, &versuche, &fehler); err != nil {
		t.Fatalf("Auftrag %d lesen: %v", id, err)
	}
	return status, versuche, fehler.String
}

// statusMap reads the status per order ID directly from the table.
func statusMap(t *testing.T, db *sql.DB) map[int]string {
	t.Helper()
	rows, err := db.Query("SELECT id, status FROM tse_signaturauftraege")
	if err != nil {
		t.Fatalf("statusMap query: %v", err)
	}
	defer func() { _ = rows.Close() }()
	status := map[int]string{}
	for rows.Next() {
		var id int
		var s string
		if err := rows.Scan(&id, &s); err != nil {
			t.Fatalf("statusMap scan: %v", err)
		}
		status[id] = s
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("statusMap rows: %v", err)
	}
	return status
}

// The queue state counts open and fehlgeschlagen orders and measures the backlog (age of the oldest open order)
// and the 15-minute window (signatures per minute, signing p95). Without orders all values are 0.
func TestGetTSESignaturQueueZustand(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	leer, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if leer.OffeneAuftraege != 0 || leer.RueckstandSekunden != 0 || leer.SignaturenProMinute != 0 || leer.SignierdauerP95Sekunden != 0 {
		t.Errorf("Expected empty queue zustand, got %+v", leer)
	}

	// One open and one fehlgeschlagen order; one erledigt within the window.
	umgebung.insertAuftrag(t, "tx-offen")
	fehlID, _ := umgebung.insertAuftrag(t, "tx-fehl")
	markiereFehlgeschlagen(ctx, t, store, fehlID, "fiskaly down")
	erledigtID, _ := umgebung.insertAuftrag(t, "tx-erledigt")
	if err := store.QuittiereTSESignaturauftrag(ctx, erledigtID, testSignatur(60)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}

	zustand, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.OffeneAuftraege != 1 {
		t.Errorf("Expected 1 offenen auftrag, got %d", zustand.OffeneAuftraege)
	}
	if zustand.FehlgeschlageneAuftraege != 1 {
		t.Errorf("Expected 1 fehlgeschlagenen auftrag, got %d", zustand.FehlgeschlageneAuftraege)
	}
	if zustand.LetzterFehler != "fiskaly down" {
		t.Errorf("Expected letzter fehler of active session, got %q", zustand.LetzterFehler)
	}
	if zustand.SignaturenProMinute <= 0 {
		t.Errorf("Expected positive signaturen pro minute, got %v", zustand.SignaturenProMinute)
	}
}

// The fehlgeschlagen count and last error cover the active Kassensitzung only: the Kassenabschluss clears the warning,
// and a new session reports only its own latest error.
func TestGetTSESignaturQueueZustand_FehlgeschlagenSitzungsbezogen(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	// With an active session the fehlgeschlagen order counts and carries its error text.
	fehlID, _ := umgebung.insertAuftrag(t, "tx-fehl-aktiv")
	markiereFehlgeschlagen(ctx, t, store, fehlID, "fiskaly 503")

	zustand, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 1 || zustand.LetzterFehler != "fiskaly 503" {
		t.Errorf("Expected 1 fehlgeschlagenen auftrag der aktiven Sitzung mit Fehlertext, got %+v", zustand)
	}

	// After the Kassenabschluss (session abgeschlossen) the warning disappears.
	umgebung.closeKassensitzung(t, umgebung.ksNr)

	zustand, err = store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 0 || zustand.LetzterFehler != "" {
		t.Errorf("Expected no fehlgeschlagen-Warnung ohne aktive Sitzung, got %+v", zustand)
	}

	// A new active session reports only its own error; the closed session's incident no longer counts.
	neueNr := umgebung.insertKassensitzung(t)
	neuFehlID, _ := umgebung.insertAuftragFuerSitzung(t, "tx-fehl-neu", neueNr)
	markiereFehlgeschlagen(ctx, t, store, neuFehlID, "fiskaly timeout")

	zustand, err = store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 1 || zustand.LetzterFehler != "fiskaly timeout" {
		t.Errorf("Expected only the new session's failure, got %+v", zustand)
	}
}

// GetAlleTSEStoerungen returns the Störungsprotokoll newest first; the active Störungszeitraum has no end, the closed one does.
func TestGetAlleTSEStoerungen(t *testing.T) {
	store, _, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
		t.Fatalf("Expected no schliessen error, got %v", err)
	}
	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	stoerungen, err := store.GetAlleTSEStoerungen(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(stoerungen) != 2 {
		t.Fatalf("Expected 2 stoerungen, got %d", len(stoerungen))
	}
	// Newest first: the active Rückstand Störungszeitraum without end.
	if stoerungen[0].GrundArt != tse.StoerungGrundRueckstand || stoerungen[0].Ende != nil {
		t.Errorf("Expected active rueckstand first, got %+v", stoerungen[0])
	}
	if stoerungen[1].GrundArt != tse.StoerungGrundTSEFehler || stoerungen[1].Ende == nil {
		t.Errorf("Expected closed tse_fehler second, got %+v", stoerungen[1])
	}
}

// At most one Störungszeitraum is active: opening while one is active is a no-op, even for another Grund-Art.
// After closing, a new one can open.
func TestTSEStoerung_OeffnenIdempotentHoechstensEineAktiv(t *testing.T) {
	store, _, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	// No Störung: no active Störungszeitraum.
	aktive, err := store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive != nil {
		t.Errorf("Expected no active stoerung, got %+v", aktive)
	}

	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand ueber der Schwelle"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Opening again (even for another Grund-Art) is a no-op.
	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
		t.Errorf("Expected no-op oeffnen without error, got %v", err)
	}

	aktive, err = store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive == nil || aktive.GrundArt != tse.StoerungGrundRueckstand {
		t.Fatalf("Expected single active rueckstand stoerung, got %+v", aktive)
	}
	if aktive.Fehlertext != "Rueckstand ueber der Schwelle" || aktive.Beginn.IsZero() {
		t.Errorf("Expected fehlertext and beginn of first oeffnen, got %+v", aktive)
	}

	var anzahl int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM tse_stoerungen").Scan(&anzahl); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if anzahl != 1 {
		t.Errorf("Expected exactly 1 stoerung row, got %d", anzahl)
	}
}

// Each writer closes only its own Grund-Art: closing another Grund-Art is a no-op,
// closing its own ends the Störungszeitraum and makes room for a new one.
func TestTSEStoerung_SchliessenNurEigeneGrundArt(t *testing.T) {
	store, _, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Another Grund-Art does not close it.
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
		t.Errorf("Expected no-op schliessen without error, got %v", err)
	}
	aktive, err := store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive == nil {
		t.Error("Expected stoerung to stay active after foreign schliessen")
	}

	// The own Grund-Art closes it; closing again is a no-op.
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
		t.Fatalf("Expected no schliessen error, got %v", err)
	}
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
		t.Errorf("Expected idempotent schliessen without error, got %v", err)
	}
	aktive, err = store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive != nil {
		t.Errorf("Expected no active stoerung after schliessen, got %+v", aktive)
	}

	// The closed Störungszeitraum stays (no delete path); a new one can open now.
	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
		t.Fatalf("Expected no oeffnen error after schliessen, got %v", err)
	}
	var anzahl int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM tse_stoerungen").Scan(&anzahl); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if anzahl != 2 {
		t.Errorf("Expected 2 stoerung rows (geschlossen + aktiv), got %d", anzahl)
	}
}

// GetAeltesterOffenerTSESignaturauftrag returns the oldest open order's creation time;
// erledigt orders do not count, and no open order yields nil.
func TestGetAeltesterOffenerTSESignaturauftrag(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	aeltester, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aeltester != nil {
		t.Errorf("Expected nil without open auftraege, got %v", aeltester)
	}

	ersterID, _ := umgebung.insertAuftrag(t, "tx-aelter")
	umgebung.insertAuftrag(t, "tx-juenger")

	aeltester, err = store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aeltester == nil {
		t.Fatal("Expected erstellungszeitpunkt of oldest open auftrag, got nil")
	}

	// The oldest order is quittiert, so the younger one now sets the age.
	if err := store.QuittiereTSESignaturauftrag(ctx, ersterID, testSignatur(41)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}
	juengster, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if juengster == nil || juengster.Before(*aeltester) {
		t.Errorf("Expected timestamp of remaining open auftrag, got %v", juengster)
	}
}
