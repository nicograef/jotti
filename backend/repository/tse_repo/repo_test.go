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
	"github.com/nicograef/jotti/backend/domain/tse"
)

// testUmgebung haelt die Test-DB samt Kassensitzung und Benutzer: Jeder
// Signaturauftrag referenziert ein Kassenjournal-Event (event_id NOT NULL
// UNIQUE), daher braucht jeder Auftrag ein eigenes Event.
type testUmgebung struct {
	db      *sql.DB
	userID  int
	ksNr    int
	version int
}

func setupRepository(t *testing.T) (Repository, *testUmgebung, func(t *testing.T)) {
	t.Helper()
	database := dbpkg.OpenTestDatabase()

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
		database.Close()
	}
}

// insertAuftrag legt ein Kassenjournal-Event der Standard-Sitzung samt offenem
// Signaturauftrag an und liefert (auftragID, eventID).
func (u *testUmgebung) insertAuftrag(t *testing.T, txID string) (int, int) {
	t.Helper()
	return u.insertAuftragFuerSitzung(t, txID, u.ksNr)
}

// insertAuftragFuerSitzung legt Event und offenen Signaturauftrag fuer eine
// bestimmte Kassensitzung an — Grundlage der sitzungsbezogenen Queue-Sicht.
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

// insertKassensitzung legt eine weitere offene Kassensitzung an und liefert ihre
// z_nr. Wegen idx_kassensitzungen_eine_aktiv darf hoechstens eine Sitzung aktiv
// sein — die vorige muss vorher abgeschlossen werden.
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

// closeKassensitzung setzt die Sitzung auf abgeschlossen — derselbe
// Statuswechsel, den der Kassenabschluss ueber das Tagesabschluss-Event bewirkt.
func (u *testUmgebung) closeKassensitzung(t *testing.T, ksNr int) {
	t.Helper()
	if _, err := u.db.Exec(
		"UPDATE kassensitzungen SET status = 'abgeschlossen', updated_at = NOW() WHERE z_nr = $1", ksNr,
	); err != nil {
		t.Fatalf("close kassensitzung: %v", err)
	}
}

// markiereFehlgeschlagen laesst einen Auftrag ueber MaxSignaturVersuche
// Fehlversuche endgueltig fehlschlagen (Status fehlgeschlagen, letzter_fehler
// gesetzt).
func markiereFehlgeschlagen(t *testing.T, store Repository, ctx context.Context, auftragID int, fehler string) {
	t.Helper()
	for i := 0; i < MaxSignaturVersuche; i++ {
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

// Die Quittierung fuellt die Signaturspalten genau einmal (Status-Guard offen):
// Der Auftrag wird erledigt, der Beleg-Abruf liest die Signatur vom Auftrag,
// und eine zweite Quittierung aendert nichts mehr.
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
		t.Fatalf("Expected status erledigt, got %q", stand.Status)
	}
	if stand.Signatur == nil || stand.Signatur.TransaktionNummer != 41 || stand.Signatur.Signatur != "SIG-1" {
		t.Fatalf("Expected quittierte signatur at auftrag, got %+v", stand.Signatur)
	}

	// Erledigte Auftraege sind nicht mehr faellig.
	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected no due auftraege after quittierung, got %+v", offene)
	}

	// Zweite Quittierung ist ein No-Op (Signaturspalten genau einmal beschrieben).
	if err := store.QuittiereTSESignaturauftrag(ctx, auftragID, testSignatur(99)); err != nil {
		t.Fatalf("Expected no error from repeated quittierung, got %v", err)
	}
	stand, err = store.GetSignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if stand.Signatur.TransaktionNummer != 41 {
		t.Fatalf("Expected signature to stay at 41, got %d", stand.Signatur.TransaktionNummer)
	}
}

// Der Beleg-Abruf unterscheidet ueber GetSignaturauftragZuEvent: kein Auftrag
// (nicht signaturpflichtig) -> db.ErrNotFound; offener Auftrag -> Stand ohne
// Signatur.
func TestGetSignaturauftragZuEvent_Faelle(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	if _, err := store.GetSignaturauftragZuEvent(ctx, 999999); !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected db.ErrNotFound for event without auftrag, got %v", err)
	}

	_, eventID := umgebung.insertAuftrag(t, "tx-offen-stand")
	stand, err := store.GetSignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if stand.Status != tse.StatusOffen || stand.Signatur != nil {
		t.Fatalf("Expected offenen stand ohne signatur, got %+v", stand)
	}
	if stand.ErstelltAm.IsZero() {
		t.Fatal("Expected erstellt_am to be set")
	}
}

// Ein Fehlversuch verschiebt den naechsten Versuch in die Zukunft (Backoff):
// Der fehlschlagende Auftrag verschwindet aus dem Worker-Batch, ein neuerer
// Auftrag bleibt abholbar — kein Head-of-Line-Blocking mehr.
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
		t.Fatalf("Expected only the newer auftrag to be due, got %+v", offene)
	}

	// Der fehlschlagende Auftrag hat den Fehlversuch verbucht und bleibt offen.
	status, versuche, letzterFehler := auftragStatus(t, umgebung.db, fehlschlagendID)
	if status != "offen" || versuche != 1 || letzterFehler != "fiskaly timeout" {
		t.Fatalf("Expected recorded fehlversuch, got status=%q versuche=%d fehler=%q", status, versuche, letzterFehler)
	}
}

// Die auftragsspezifische Fehlversuchs-Kurve ist eine Sekunden-Kurve (5, 15 s
// Backoff, dritter Fehlversuch endgueltig fehlgeschlagen): Sie endet deutlich
// unter der Rueckstands-Schwelle, und der fehlgeschlagene Auftrag verschwindet
// aus der Rueckstands-Messung — ein Gift-Auftrag oeffnet nie einen
// Rueckstands-Zeitraum und liefert bis zum endgueltigen Fehlschlag das
// Signaturstatus-Ergebnis ausstehend.
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

		// Der Gift-Auftrag ist waehrend der Kurve ausstehend, nie Ausfall.
		stand, err := store.GetSignaturauftragZuEvent(ctx, eventID)
		if err != nil {
			t.Fatalf("Stand nach Fehlversuch %d lesen: %v", versuch, err)
		}
		if ergebnis := tse.DetermineSignaturstatus(stand, nil); ergebnis.Status != tse.SignaturstatusAusstehend {
			t.Fatalf("Fehlversuch %d: erwartet ausstehend, got %q", versuch, ergebnis.Status)
		}

		backoff := backoffBis(t, umgebung.db, id)
		erwartet := erwarteteBackoffs[versuch-1]
		if backoff < erwartet-2*time.Second || backoff > erwartet+2*time.Second {
			t.Fatalf("Fehlversuch %d: Backoff %v, erwartet ~%v", versuch, backoff, erwartet)
		}
		gesamtBackoff += backoff
	}

	// Der MaxSignaturVersuche-te Fehlversuch macht den Auftrag endgueltig.
	if err := store.TSESignaturauftragFehlversuch(ctx, id, "fiskaly api error 400 (E_FAILED_SCHEMA_VALIDATION)"); err != nil {
		t.Fatalf("letzter Fehlversuch: %v", err)
	}
	if status, _, _ := auftragStatus(t, umgebung.db, id); status != tse.StatusFehlgeschlagen {
		t.Fatalf("Expected endgueltig fehlgeschlagenen Auftrag, got %q", status)
	}

	// Die gesamte Wartezeit der Kurve bleibt weit unter der
	// Rueckstands-Schwelle — Platz fuer Tick- und Verarbeitungs-Schlupf.
	if gesamtBackoff >= tse.RueckstandSchwelle/2 {
		t.Fatalf("Backoff-Kurve %v zu nah an der Rueckstands-Schwelle %v", gesamtBackoff, tse.RueckstandSchwelle)
	}

	// Fehlgeschlagene Auftraege zaehlen nicht als Rueckstand: Der Watchdog
	// misst nur offene Auftraege.
	aeltester, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Rueckstand messen: %v", err)
	}
	if aeltester != nil {
		t.Fatalf("Expected fehlgeschlagenen Auftrag nicht in der Rueckstands-Messung, got %v", aeltester)
	}
}

// backoffBis misst den von der Fehlversuchs-Query gesetzten Backoff
// (naechster_versuch_am − NOW()).
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

// Ohne TSE-Konfiguration markiert der Worker offene Auftraege endgueltig als
// tse_nicht_konfiguriert; erledigte bleiben unberuehrt.
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
		t.Fatalf("Expected 2 marked auftraege (die beiden offenen), got %d", markiert)
	}

	status := statusMap(t, umgebung.db)
	if status[offenID] != tse.StatusTSENichtKonfiguriert || status[zweiterID] != tse.StatusTSENichtKonfiguriert {
		t.Fatalf("Expected offene auftraege marked tse_nicht_konfiguriert, got %+v", status)
	}
	if status[erledigtID] != tse.StatusErledigt {
		t.Fatalf("Expected erledigten auftrag untouched, got %q", status[erledigtID])
	}

	// Markierte Auftraege sind nicht mehr faellig.
	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected no due auftraege after marking, got %+v", offene)
	}

	// Eine zweite Markierung ohne offene Auftraege markiert nichts.
	markiert, err = store.MarkOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		t.Fatalf("Expected no mark error, got %v", err)
	}
	if markiert != 0 {
		t.Fatalf("Expected 0 marked on second run, got %d", markiert)
	}
}

// auftragStatus liest Status, Versuche und letzten Fehler eines Auftrags direkt
// aus der Tabelle — die Admin-Lese-Query gibt es nicht mehr, die Tests pruefen
// den Auftragszustand per SQL.
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

// statusMap liest Status je Auftrag-ID direkt aus der Tabelle.
func statusMap(t *testing.T, db *sql.DB) map[int]string {
	t.Helper()
	rows, err := db.Query("SELECT id, status FROM tse_signaturauftraege")
	if err != nil {
		t.Fatalf("statusMap query: %v", err)
	}
	defer rows.Close()
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

// Der Queue-Zustand zaehlt offene und fehlgeschlagene Auftraege, misst den
// Rueckstand (Alter des aeltesten offenen) und die Leistung ueber das
// 15-Minuten-Fenster (Signaturen pro Minute, Signierdauer p95). Ohne Auftraege
// sind alle Werte 0.
func TestGetTSESignaturQueueZustand(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	leer, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if leer.OffeneAuftraege != 0 || leer.RueckstandSekunden != 0 || leer.SignaturenProMinute != 0 || leer.SignierdauerP95Sekunden != 0 {
		t.Fatalf("Expected empty queue zustand, got %+v", leer)
	}

	// Ein offener und ein fehlgeschlagener Auftrag; ein erledigter im Fenster.
	umgebung.insertAuftrag(t, "tx-offen")
	fehlID, _ := umgebung.insertAuftrag(t, "tx-fehl")
	markiereFehlgeschlagen(t, store, ctx, fehlID, "fiskaly down")
	erledigtID, _ := umgebung.insertAuftrag(t, "tx-erledigt")
	if err := store.QuittiereTSESignaturauftrag(ctx, erledigtID, testSignatur(60)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}

	zustand, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.OffeneAuftraege != 1 {
		t.Fatalf("Expected 1 offenen auftrag, got %d", zustand.OffeneAuftraege)
	}
	if zustand.FehlgeschlageneAuftraege != 1 {
		t.Fatalf("Expected 1 fehlgeschlagenen auftrag, got %d", zustand.FehlgeschlageneAuftraege)
	}
	if zustand.LetzterFehler != "fiskaly down" {
		t.Fatalf("Expected letzter fehler of active session, got %q", zustand.LetzterFehler)
	}
	if zustand.SignaturenProMinute <= 0 {
		t.Fatalf("Expected positive signaturen pro minute, got %v", zustand.SignaturenProMinute)
	}
}

// Die fehlgeschlagen-Zahl und der letzte Fehlertext sind sitzungsbezogen: Sie
// zaehlen nur Auftraege der aktiven Kassensitzung. Mit dem Kassenabschluss
// (Sitzung -> abgeschlossen) verschwindet die Warnung ohne weiteres Zutun, und
// ein Vorfall aus einer bereits abgeschlossenen Sitzung zaehlt nicht mehr — die
// neue aktive Sitzung weist nur ihren eigenen juengsten Fehler aus.
func TestGetTSESignaturQueueZustand_FehlgeschlagenSitzungsbezogen(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	// Mit aktiver Sitzung: der fehlgeschlagene Auftrag zaehlt und traegt seinen Fehlertext.
	fehlID, _ := umgebung.insertAuftrag(t, "tx-fehl-aktiv")
	markiereFehlgeschlagen(t, store, ctx, fehlID, "fiskaly 503")

	zustand, err := store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 1 || zustand.LetzterFehler != "fiskaly 503" {
		t.Fatalf("Expected 1 fehlgeschlagenen auftrag der aktiven Sitzung mit Fehlertext, got %+v", zustand)
	}

	// Nach dem Kassenabschluss (Sitzung abgeschlossen) verschwindet die Warnung.
	umgebung.closeKassensitzung(t, umgebung.ksNr)

	zustand, err = store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 0 || zustand.LetzterFehler != "" {
		t.Fatalf("Expected no fehlgeschlagen-Warnung ohne aktive Sitzung, got %+v", zustand)
	}

	// Eine neue aktive Sitzung mit eigenem Fehler weist nur ihren eigenen Fehler
	// aus; der Vorfall der abgeschlossenen Sitzung zaehlt nicht mehr.
	neueNr := umgebung.insertKassensitzung(t)
	neuFehlID, _ := umgebung.insertAuftragFuerSitzung(t, "tx-fehl-neu", neueNr)
	markiereFehlgeschlagen(t, store, ctx, neuFehlID, "fiskaly timeout")

	zustand, err = store.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		t.Fatalf("Expected no queue error, got %v", err)
	}
	if zustand.FehlgeschlageneAuftraege != 1 || zustand.LetzterFehler != "fiskaly timeout" {
		t.Fatalf("Expected only the new session's failure, got %+v", zustand)
	}
}

// GetAlleTSEStoerungen liefert das Stoerungsprotokoll neueste zuerst; der aktive
// Zeitraum traegt kein Ende, der geschlossene eines.
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
	// Neueste zuerst: der aktive Rueckstands-Zeitraum ohne Ende.
	if stoerungen[0].GrundArt != tse.StoerungGrundRueckstand || stoerungen[0].Ende != nil {
		t.Fatalf("Expected active rueckstand first, got %+v", stoerungen[0])
	}
	if stoerungen[1].GrundArt != tse.StoerungGrundTSEFehler || stoerungen[1].Ende == nil {
		t.Fatalf("Expected closed tse_fehler second, got %+v", stoerungen[1])
	}
}

// Hoechstens ein Stoerungszeitraum ist aktiv: Oeffnen bei aktivem Zeitraum ist
// ein No-Op — auch fuer eine andere Grund-Art. Nach dem Schliessen kann ein
// neuer Zeitraum entstehen.
func TestTSEStoerung_OeffnenIdempotentHoechstensEineAktiv(t *testing.T) {
	store, _, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	// Ohne Stoerung: kein aktiver Zeitraum.
	aktive, err := store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive != nil {
		t.Fatalf("Expected no active stoerung, got %+v", aktive)
	}

	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand ueber der Schwelle"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Erneutes Oeffnen (auch anderer Grund-Art) ist ein No-Op.
	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
		t.Fatalf("Expected no-op oeffnen without error, got %v", err)
	}

	aktive, err = store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive == nil || aktive.GrundArt != tse.StoerungGrundRueckstand {
		t.Fatalf("Expected single active rueckstand stoerung, got %+v", aktive)
	}
	if aktive.Fehlertext != "Rueckstand ueber der Schwelle" || aktive.Beginn.IsZero() {
		t.Fatalf("Expected fehlertext and beginn of first oeffnen, got %+v", aktive)
	}

	var anzahl int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM tse_stoerungen").Scan(&anzahl); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if anzahl != 1 {
		t.Fatalf("Expected exactly 1 stoerung row, got %d", anzahl)
	}
}

// Jeder Schreiber schliesst nur Zeitraeume seiner Grund-Art: Das Schliessen
// einer fremden Grund-Art ist ein No-Op, das eigene beendet den Zeitraum und
// macht Platz fuer einen neuen.
func TestTSEStoerung_SchliessenNurEigeneGrundArt(t *testing.T) {
	store, _, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Fremde Grund-Art schliesst nicht.
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
		t.Fatalf("Expected no-op schliessen without error, got %v", err)
	}
	aktive, err := store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive == nil {
		t.Fatal("Expected stoerung to stay active after foreign schliessen")
	}

	// Eigene Grund-Art schliesst; erneutes Schliessen ist ein No-Op.
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
		t.Fatalf("Expected no schliessen error, got %v", err)
	}
	if err := store.CloseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
		t.Fatalf("Expected idempotent schliessen without error, got %v", err)
	}
	aktive, err = store.GetAktiveTSEStoerung(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aktive != nil {
		t.Fatalf("Expected no active stoerung after schliessen, got %+v", aktive)
	}

	// Der geschlossene Zeitraum bleibt erhalten (kein Loeschpfad); ein neuer
	// Zeitraum kann jetzt entstehen.
	if err := store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
		t.Fatalf("Expected no oeffnen error after schliessen, got %v", err)
	}
	var anzahl int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM tse_stoerungen").Scan(&anzahl); err != nil {
		t.Fatalf("count stoerungen: %v", err)
	}
	if anzahl != 2 {
		t.Fatalf("Expected 2 stoerung rows (geschlossen + aktiv), got %d", anzahl)
	}
}

// GetAeltesterOffenerTSESignaturauftrag liefert den Erstellungszeitpunkt des
// aeltesten offenen Auftrags; erledigte Auftraege zaehlen nicht, ohne offene
// Auftraege kommt nil.
func TestGetAeltesterOffenerTSESignaturauftrag(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	aeltester, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if aeltester != nil {
		t.Fatalf("Expected nil without open auftraege, got %v", aeltester)
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

	// Der aelteste Auftrag wird quittiert — der juengere bestimmt jetzt das Alter.
	if err := store.QuittiereTSESignaturauftrag(ctx, ersterID, testSignatur(41)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}
	juengster, err := store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if juengster == nil || juengster.Before(*aeltester) {
		t.Fatalf("Expected timestamp of remaining open auftrag, got %v", juengster)
	}
}
