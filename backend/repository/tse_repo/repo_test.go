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

// insertAuftrag legt ein Kassenjournal-Event samt offenem Signaturauftrag an
// und liefert (auftragID, eventID).
func (u *testUmgebung) insertAuftrag(t *testing.T, txID string) (int, int) {
	t.Helper()
	u.version++
	var eventID int
	if err := u.db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'Test', 'zahlung-kassiert:v1', $2, $3, '{}', NOW(), $4) RETURNING id",
		u.userID, fmt.Sprintf("kassensitzung-%d/tisch-1", u.ksNr), u.version, u.ksNr,
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

	auftraege, err := store.GetTSESignaturauftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	// Neueste zuerst: tx-neuer vor tx-fehlschlagend.
	if len(auftraege) != 2 || auftraege[1].ID != fehlschlagendID {
		t.Fatalf("Expected both auftraege newest first, got %+v", auftraege)
	}
	if auftraege[1].Status != "offen" || auftraege[1].Versuche != 1 || auftraege[1].LetzterFehler != "fiskaly timeout" {
		t.Fatalf("Expected recorded fehlversuch, got %+v", auftraege[1])
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
		if ergebnis := tse.BestimmeSignaturstatus(stand, nil); ergebnis.Status != tse.SignaturstatusAusstehend {
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
	auftraege, err := store.GetTSESignaturauftraege(ctx)
	if err != nil {
		t.Fatalf("Auftraege lesen: %v", err)
	}
	if len(auftraege) != 1 || auftraege[0].Status != tse.StatusFehlgeschlagen {
		t.Fatalf("Expected endgueltig fehlgeschlagenen Auftrag, got %+v", auftraege)
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

// Nach MaxSignaturVersuche Fehlversuchen wechselt der Auftrag auf
// fehlgeschlagen; Zuruecksetzen reiht ihn wieder ein, Verwerfen beendet ihn.
func TestTSESignaturauftrag_FehlgeschlagenZuruecksetzenVerwerfen(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	id, _ := umgebung.insertAuftrag(t, "tx-dauerhaft")

	for i := 0; i < MaxSignaturVersuche; i++ {
		if err := store.TSESignaturauftragFehlversuch(ctx, id, "fiskaly down"); err != nil {
			t.Fatalf("Expected no fehlversuch error, got %v", err)
		}
	}

	auftraege, err := store.GetTSESignaturauftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(auftraege) != 1 || auftraege[0].Status != "fehlgeschlagen" || auftraege[0].Versuche != MaxSignaturVersuche {
		t.Fatalf("Expected fehlgeschlagenen auftrag after max versuche, got %+v", auftraege)
	}

	// Ein weiterer Fehlversuch aendert nichts mehr (Status-Guard offen).
	if err := store.TSESignaturauftragFehlversuch(ctx, id, "noch ein fehler"); err != nil {
		t.Fatalf("Expected no fehlversuch error, got %v", err)
	}
	auftraege, _ = store.GetTSESignaturauftraege(ctx)
	if auftraege[0].Versuche != MaxSignaturVersuche {
		t.Fatalf("Expected versuche to stay at %d, got %d", MaxSignaturVersuche, auftraege[0].Versuche)
	}

	// Zuruecksetzen: fehlgeschlagen -> offen, sofort wieder faellig.
	if err := store.TSESignaturauftragZuruecksetzen(ctx, id); err != nil {
		t.Fatalf("Expected no zuruecksetzen error, got %v", err)
	}
	offene, err := store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != id {
		t.Fatalf("Expected reset auftrag to be due again, got %+v", offene)
	}

	// Erneut fehlschlagen lassen und verwerfen.
	for i := 0; i < MaxSignaturVersuche; i++ {
		if err := store.TSESignaturauftragFehlversuch(ctx, id, "fiskaly down"); err != nil {
			t.Fatalf("Expected no fehlversuch error, got %v", err)
		}
	}
	if err := store.TSESignaturauftragVerwerfen(ctx, id); err != nil {
		t.Fatalf("Expected no verwerfen error, got %v", err)
	}
	auftraege, _ = store.GetTSESignaturauftraege(ctx)
	if len(auftraege) != 1 || auftraege[0].Status != "verworfen" {
		t.Fatalf("Expected verworfenen auftrag, got %+v", auftraege)
	}
}

// Ohne TSE-Konfiguration markiert der Worker offene Auftraege endgueltig als
// tse_nicht_konfiguriert; erledigte bleiben unberuehrt. Das Admin-Zuruecksetzen
// reiht einen so markierten Auftrag wieder ein, damit der Worker ihn nach einer
// spaeten Einrichtung nachsigniert.
func TestMarkiereOffeneAlsNichtKonfiguriert_MarkiertOffeneUndBleibtZuruecksetzbar(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	offenID, _ := umgebung.insertAuftrag(t, "tx-ohne-konfig-1")
	zweiterID, _ := umgebung.insertAuftrag(t, "tx-ohne-konfig-2")
	erledigtID, _ := umgebung.insertAuftrag(t, "tx-erledigt")
	if err := store.QuittiereTSESignaturauftrag(ctx, erledigtID, testSignatur(41)); err != nil {
		t.Fatalf("Expected no quittierung error, got %v", err)
	}

	markiert, err := store.MarkiereOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		t.Fatalf("Expected no mark error, got %v", err)
	}
	if markiert != 2 {
		t.Fatalf("Expected 2 marked auftraege (die beiden offenen), got %d", markiert)
	}

	auftraege, err := store.GetTSESignaturauftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	status := map[int]string{}
	for _, a := range auftraege {
		status[a.ID] = a.Status
	}
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
	markiert, err = store.MarkiereOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		t.Fatalf("Expected no mark error, got %v", err)
	}
	if markiert != 0 {
		t.Fatalf("Expected 0 marked on second run, got %d", markiert)
	}

	// Admin-Zuruecksetzen reiht den markierten Auftrag wieder ein.
	if err := store.TSESignaturauftragZuruecksetzen(ctx, offenID); err != nil {
		t.Fatalf("Expected no zuruecksetzen error, got %v", err)
	}
	offene, err = store.GetOffeneTSESignaturauftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != offenID {
		t.Fatalf("Expected reset auftrag to be due again, got %+v", offene)
	}
}

// Zuruecksetzen und Verwerfen wirken nur auf fehlgeschlagene Auftraege.
func TestTSESignaturauftrag_StatusGuards(t *testing.T) {
	store, umgebung, teardown := setupRepository(t)
	defer teardown(t)
	ctx := context.Background()

	id, _ := umgebung.insertAuftrag(t, "tx-offen")

	if err := store.TSESignaturauftragVerwerfen(ctx, id); err != nil {
		t.Fatalf("Expected no verwerfen error, got %v", err)
	}
	if err := store.TSESignaturauftragZuruecksetzen(ctx, id); err != nil {
		t.Fatalf("Expected no zuruecksetzen error, got %v", err)
	}

	auftraege, err := store.GetTSESignaturauftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(auftraege) != 1 || auftraege[0].Status != "offen" {
		t.Fatalf("Expected offenen auftrag to stay offen, got %+v", auftraege)
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

	if err := store.OeffneTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand ueber der Schwelle"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Erneutes Oeffnen (auch anderer Grund-Art) ist ein No-Op.
	if err := store.OeffneTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
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

	if err := store.OeffneTSEStoerung(ctx, tse.StoerungGrundRueckstand, "Rueckstand"); err != nil {
		t.Fatalf("Expected no oeffnen error, got %v", err)
	}

	// Fremde Grund-Art schliesst nicht.
	if err := store.SchliesseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
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
	if err := store.SchliesseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
		t.Fatalf("Expected no schliessen error, got %v", err)
	}
	if err := store.SchliesseTSEStoerung(ctx, tse.StoerungGrundRueckstand); err != nil {
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
	if err := store.OeffneTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "HTTP 503"); err != nil {
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
