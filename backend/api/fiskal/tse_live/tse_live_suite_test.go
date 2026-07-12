//go:build integration

// Package tse_live ist die TSE-Live-Suite: Sie loest jeden signaturpflichtigen
// Geschaeftsvorfall ueber die echten Anwendungsdienste aus, laesst ihn vom
// echten Signatur-Worker gegen die fiskaly-TEST-TSS real signieren und prueft
// je Vorfall den abgeschlossenen Signaturauftrag, die Signaturdaten im
// Kassenjournal-Outbox-Eintrag und das processType-Mapping.
//
// Live-Guard nach dem Muster von repository/tse_repo/fiskaly_client_live_test.go:
// Ohne FISKALY_TEST_*-Credentials skippt die Suite; die Verbindung wird zu
// Beginn gegen tse.UmgebungTest geprueft und bricht bei jeder Nicht-TEST-
// Umgebung hart ab, damit nie gegen eine LIVE-TSS signiert wird.
//
//	make test-tse-live   # Wegwerf-Postgres + Migrationen + .env.fiskaly-test
package tse_live

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nicograef/jotti/backend/api/fiskal/signatur"
	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	kassenfuehrungApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/betreiber_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// signaturWartefrist begrenzt, wie lange auf die Quittierung eines Auftrags
// durch den Signatur-Worker gewartet wird. Grosszuegig gewaehlt: ein realer
// fiskaly-Roundtrip inkl. moeglichem 429-Backoff des Workers dauert Sekunden.
const signaturWartefrist = 90 * time.Second

// liveTestUmgebung buendelt die reale Umgebung eines Live-Laufs: DB, die
// verdrahteten Anwendungsdienste und die Stammdaten-IDs.
type liveTestUmgebung struct {
	db       *sql.DB
	tisch    tischgeschaeftApp.Command
	direkt   direktverkaufApp.Command
	kasse    kassenfuehrungApp.Command
	tssID    string
	userID   int
	tischID  int
	tischID2 int
	produkt  int
	variante int
}

// credentialsOderSkip verlangt das explizite Opt-in JOTTI_TSE_LIVE=1 und liest
// dann die fiskaly-TEST-Credentials aus der Umgebung; fehlt eines von beidem,
// wird die Suite geskippt (gleiche Guard-Schwelle wie im
// fiskaly_client_live_test.go). Das Opt-in hält normale Integrationsläufe
// (scripts/test-integration.sh) hermetisch, auch wenn FISKALY_TEST_*-Variablen
// in der Shell exportiert sind.
func credentialsOderSkip(t *testing.T) tse.Credentials {
	t.Helper()
	if os.Getenv("JOTTI_TSE_LIVE") != "1" {
		t.Skip("JOTTI_TSE_LIVE != 1 — TSE-Live-Suite übersprungen (Opt-in via make test-tse-live)")
	}
	credentials := tse.Credentials{
		ApiKey:    os.Getenv("FISKALY_TEST_API_KEY"),
		ApiSecret: os.Getenv("FISKALY_TEST_API_SECRET"),
		TssID:     os.Getenv("FISKALY_TEST_TSS_ID"),
		ClientID:  os.Getenv("FISKALY_TEST_CLIENT_ID"),
	}
	if credentials.Validate() != nil {
		t.Skip("FISKALY_TEST_* nicht gesetzt — TSE-Live-Suite übersprungen")
	}
	return credentials
}

func fiskalyBaseURL() string {
	baseURL := os.Getenv("FISKALY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://kassensichv-middleware.fiskaly.com"
	}
	return baseURL
}

// pruefeTestUmgebungOderAbbruch stellt sicher, dass die Credentials auf die
// TEST-Umgebung zeigen. Jeder andere Befund ist ein harter Abbruch: gegen eine
// LIVE-TSS darf die Suite nie signieren.
func pruefeTestUmgebungOderAbbruch(t *testing.T, credentials tse.Credentials) {
	t.Helper()
	client, err := tse_repo.NewFiskalyTSEClient(fiskalyBaseURL(), credentials, nil)
	if err != nil {
		t.Fatalf("fiskaly-Client bauen: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	status, err := client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("TestConnection gegen fiskaly fehlgeschlagen: %v", err)
	}
	if status.Umgebung != tse.UmgebungTest {
		t.Fatalf("Live-Guard: Suite nur gegen die TEST-Umgebung erlaubt, Credentials zeigen auf %s", status.Umgebung)
	}
	if status.ClientState != "REGISTERED" {
		t.Fatalf("Live-Guard: erwartet einen REGISTERED Client, bekam %q", status.ClientState)
	}
}

// cleanLiveDB raeumt alle im Lauf beschriebenen Tabellen ab. Das Kassenjournal
// ist append-only (Loesch-Trigger); fuer den Test-Reset wird der Trigger
// kurzzeitig ausgesetzt.
func cleanLiveDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM tse_stoerungen",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM tische",
		"DELETE FROM betreiber",
		"DELETE FROM users",
		"UPDATE tse_stammdaten SET seriennummer='', signatur_algorithmus='', public_key='', zertifikat='', log_time_format='', updated_at=now() WHERE id=1",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanLiveDB %q: %v", stmt, err)
		}
	}
}

// setupLiveUmgebung faehrt die volle reale Umgebung hoch: DB reinigen, die
// echten TEST-Credentials in tse_konfiguration schreiben (damit der Worker sie
// liest), Stammdaten anlegen und die Anwendungsdienste verdrahten.
func setupLiveUmgebung(t *testing.T, credentials tse.Credentials) *liveTestUmgebung {
	t.Helper()

	db := dbpkg.OpenTestDatabase()
	cleanLiveDB(t, db)
	t.Cleanup(func() {
		cleanLiveDB(t, db)
		_ = db.Close()
	})

	ctx := context.Background()

	// Echte TEST-Credentials in die Singleton-Konfiguration schreiben: der
	// Signatur-Worker liest sie von dort und spricht damit die TEST-TSS an.
	konf, err := tse.NewKonfiguration(credentials.ApiKey, credentials.ApiSecret, credentials.TssID, credentials.ClientID)
	if err != nil {
		t.Fatalf("Konfiguration bauen: %v", err)
	}
	if err := tse_repo.NewRepository(db).SaveEinrichtung(ctx, konf); err != nil {
		t.Fatalf("TSE-Konfiguration speichern: %v", err)
	}

	u := &liveTestUmgebung{db: db, tssID: credentials.TssID}

	if err := db.QueryRow(
		"INSERT INTO betreiber (vereinsname, strasse, plz, ort, updated_at) VALUES ('Testverein e.V.', 'Teststr. 1', '10115', 'Berlin', now()) RETURNING id",
	).Scan(new(int)); err != nil {
		t.Fatalf("create betreiber: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&u.userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 1', 'active', now(), now()) RETURNING id",
	).Scan(&u.tischID); err != nil {
		t.Fatalf("create tisch 1: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 2', 'active', now(), now()) RETURNING id",
	).Scan(&u.tischID2); err != nil {
		t.Fatalf("create tisch 2: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ('Bier', 'getraenk', 'regel', 'active', now(), now()) RETURNING id",
	).Scan(&u.produkt); err != nil {
		t.Fatalf("create produkt: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.5L', 350, 'active', now(), now()) RETURNING id",
		u.produkt,
	).Scan(&u.variante); err != nil {
		t.Fatalf("create variante: %v", err)
	}

	kjRepo := kassenjournal_repo.NewRepository(db)
	ksRepo := kassensitzungen_repo.NewRepository(db)
	prodRepo := produkt_repo.NewRepository(db)
	tischRepo := tisch_repo.NewRepository(db)

	druckRepo := druckstation_repo.NewRepository(db)
	u.tisch = tischgeschaeftApp.Command{
		TischRepo:           tischRepo,
		EventRepo:           kjRepo,
		ProduktRepo:         prodRepo,
		KassensitzungenRepo: ksRepo,
		DruckstationRepo:    druckRepo,
	}
	u.direkt = direktverkaufApp.Command{
		EventRepo:           kjRepo,
		ProduktRepo:         prodRepo,
		KassensitzungenRepo: ksRepo,
		DruckstationRepo:    druckRepo,
	}
	u.kasse = kassenfuehrungApp.Command{
		KassenjournalRepo:   kjRepo,
		KassensitzungenRepo: ksRepo,
		BetreiberRepo:       betreiber_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
	}
	return u
}

// starteWorker startet den echten Signatur-Worker in einer Goroutine und stoppt
// ihn über t.Cleanup am Testende. Der Worker liest die TSE-Konfiguration aus der
// DB, spricht die echte TEST-TSS an und quittiert jede Signatur direkt am Auftrag.
func starteWorker(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	worker := signatur.NewTSESignaturWorker(fiskalyBaseURL(), db)
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		worker.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-fertig
	})
}

// signaturZeile ist der quittierte Zustand eines Signaturauftrags samt seinem
// Kassenjournal-Event: alles, was ein Vorfall zur Pruefung braucht.
type signaturZeile struct {
	status          string
	processType     string
	processData     string
	transaktionNr   sql.NullInt64
	signaturZaehler sql.NullInt64
	tseSeriennummer sql.NullString
	logTimeStart    sql.NullTime
	logTimeEnd      sql.NullTime
	signatur        sql.NullString
	qrCodeData      sql.NullString
	eventType       string
}

// warteAufSignatur pollt den Auftrag des Events, bis er 'erledigt' ist, und
// gibt seine Signaturdaten zurueck. Ein fehlgeschlagener Auftrag bricht sofort
// ab (kein Warten bis zum Timeout).
func warteAufSignatur(t *testing.T, db *sql.DB, eventID int) signaturZeile {
	t.Helper()
	deadline := time.Now().Add(signaturWartefrist)
	for {
		var z signaturZeile
		err := db.QueryRow(`
			SELECT a.status, a.process_type, a.process_data,
			       a.transaktion_nummer, a.signatur_zaehler, a.tse_seriennummer,
			       a.log_time_start, a.log_time_end, a.signatur, a.qr_code_data,
			       k.type
			FROM tse_signaturauftraege a
			JOIN kassenjournal k ON k.id = a.event_id
			WHERE a.event_id = $1`, eventID).Scan(
			&z.status, &z.processType, &z.processData,
			&z.transaktionNr, &z.signaturZaehler, &z.tseSeriennummer,
			&z.logTimeStart, &z.logTimeEnd, &z.signatur, &z.qrCodeData,
			&z.eventType,
		)
		if err != nil {
			t.Fatalf("Signaturauftrag zu Event %d lesen: %v", eventID, err)
		}
		switch z.status {
		case "erledigt":
			return z
		case "fehlgeschlagen", "tse_nicht_konfiguriert":
			t.Fatalf("Signaturauftrag zu Event %d endete als %q (letzter Fehler in DB)", eventID, z.status)
		}
		if time.Now().After(deadline) {
			t.Fatalf("Signaturauftrag zu Event %d nicht binnen %s erledigt (Status %q)", eventID, signaturWartefrist, z.status)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// pruefeSignatur prueft die Vollstaendigkeit der Signaturdaten eines erledigten
// Auftrags und den erwarteten processType. processType/processData werden von
// der fiskalischen Projektion (domain/kasse/fiskalische_projektion.go) nach
// DSFinV-K 2.4 Anhang I gesetzt; die Suite prueft den quittierten Snapshot.
func pruefeSignatur(t *testing.T, vorfall string, z signaturZeile, erwarteterProcessType string) {
	t.Helper()
	if z.processType != erwarteterProcessType {
		t.Errorf("%s: processType %q, erwartet %q (DSFinV-K Anhang I)", vorfall, z.processType, erwarteterProcessType)
	}
	if !z.signatur.Valid || z.signatur.String == "" {
		t.Errorf("%s: keine Signatur im Kassenjournal-Auftrag", vorfall)
	}
	if !z.qrCodeData.Valid || z.qrCodeData.String == "" {
		t.Errorf("%s: keine QR-Code-Daten im Auftrag", vorfall)
	}
	if !z.transaktionNr.Valid || z.transaktionNr.Int64 == 0 {
		t.Errorf("%s: keine TSE-Transaktionsnummer", vorfall)
	}
	if !z.signaturZaehler.Valid {
		t.Errorf("%s: kein Signaturzähler", vorfall)
	}
	if !z.tseSeriennummer.Valid || z.tseSeriennummer.String == "" {
		t.Errorf("%s: keine TSE-Seriennummer", vorfall)
	}
	if !z.logTimeStart.Valid || z.logTimeStart.Time.IsZero() {
		t.Errorf("%s: kein log_time_start", vorfall)
	}
	if !z.logTimeEnd.Valid || z.logTimeEnd.Time.IsZero() {
		t.Errorf("%s: kein log_time_end", vorfall)
	}
}

// eventIDByType liefert die kassenjournal-ID des (einzigen erwarteten) Events
// eines Typs zum Subject. Die Suite erzeugt je Vorfall genau ein Event dieses
// Typs, sodass die Zuordnung eindeutig ist.
func eventIDByType(t *testing.T, db *sql.DB, eventType, subject string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		"SELECT id FROM kassenjournal WHERE type = $1 AND subject = $2 ORDER BY id DESC LIMIT 1",
		eventType, subject,
	).Scan(&id); err != nil {
		t.Fatalf("Event %q zu %q suchen: %v", eventType, subject, err)
	}
	return id
}

// positionRefsAusSession liest die aktuell unbezahlten Positionen eines Tischs
// und baut PositionRefs ueber genau menge Stueck der ersten Position.
func positionRefsAusSession(t *testing.T, u *liveTestUmgebung, ksNr, tischID, menge int) []kasse.PositionRef {
	t.Helper()
	session, err := kassenjournal_repo.NewRepository(u.db).ReadTischSession(context.Background(), kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("Tisch-Session lesen: %v", err)
	}
	if len(session.UnbezahltePositionen) == 0 {
		t.Fatalf("keine unbezahlten Positionen auf Tisch %d", tischID)
	}
	return []kasse.PositionRef{{PositionID: session.UnbezahltePositionen[0].PositionID, Menge: menge}}
}

// restBezahlen kassiert alle noch unbezahlten Positionen eines Tischs, damit
// die Tisch-Session den für den Kassenabschluss nötigen saldo_cents = 0
// erreicht. No-Op, wenn nichts offen ist.
func restBezahlen(t *testing.T, u *liveTestUmgebung, ksNr, tischID int) {
	t.Helper()
	session, err := kassenjournal_repo.NewRepository(u.db).ReadTischSession(context.Background(), kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("Tisch-Session %d lesen: %v", tischID, err)
	}
	if len(session.UnbezahltePositionen) == 0 {
		return
	}
	refs := make([]kasse.PositionRef, 0, len(session.UnbezahltePositionen))
	for _, pos := range session.UnbezahltePositionen {
		refs = append(refs, kasse.PositionRef{PositionID: pos.PositionID, Menge: pos.Menge})
	}
	if err := u.tisch.ZahlungKassieren(context.Background(), u.userID, "test", tischID, refs, "Restzahlung"); err != nil {
		t.Fatalf("Restzahlung Tisch %d: %v", tischID, err)
	}
}

// TestTSELiveSuite_GeschaeftsvorfaelleUndStammdaten loest jeden
// signaturpflichtigen Geschaeftsvorfall ueber die Anwendungsdienste aus, laesst
// ihn real signieren und prueft Signatur, Kassenjournal-Outbox und
// processType-Mapping. Am Ende wird die Vollstaendigkeit der persistierten
// TSE-Stammdaten explizit assertet.
func TestTSELiveSuite_GeschaeftsvorfaelleUndStammdaten(t *testing.T) {
	credentials := credentialsOderSkip(t)
	pruefeTestUmgebungOderAbbruch(t, credentials)

	u := setupLiveUmgebung(t, credentials)
	starteWorker(t, u.db)

	ctx := context.Background()
	db := u.db

	// signiereUndPruefe wartet auf die Signatur des letzten Events dieses Typs
	// und prueft sie gegen den erwarteten processType.
	signiereUndPruefe := func(vorfall, eventType, subject, erwarteterProcessType string) {
		id := eventIDByType(t, db, eventType, subject)
		z := warteAufSignatur(t, db, id)
		pruefeSignatur(t, vorfall, z, erwarteterProcessType)
	}

	// (1) Kassensitzung eröffnen mit Anfangsbestand > 0 → Kassenbeleg-V1 (Bareinlage).
	ksNr, err := u.kasse.KassensitzungEroeffnen(ctx, u.userID, "test", "Live-Suite", 5000)
	if err != nil {
		t.Fatalf("KassensitzungEroeffnen: %v", err)
	}
	ksSubject := kasse.KassensitzungSubject(ksNr)
	signiereUndPruefe("Kassensitzung-Eröffnung (Bareinlage)", string(kasse.EventTypeKassensitzungEroeffnetV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// (2) Bestellung → Bestellung-V1. 3 Stück, damit Teil-/Vollzahlung und Storno Mengen haben.
	bestellungID := uuid.NewString()
	inputs := []tischgeschaeftApp.BestellPositionInput{{ProduktID: u.produkt, VarianteID: u.variante, Menge: 3}}
	if err := u.tisch.BestellungAufnehmen(ctx, u.userID, "test", bestellungID, u.tischID, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen: %v", err)
	}
	tischSubject := kasse.TischSessionSubject(ksNr, u.tischID)
	signiereUndPruefe("Bestellung", string(kasse.EventTypeBestellungAufgenommenV1), tischSubject, tse.ProcessTypeBestellungV1)

	// (3) Teilzahlung: 1 von 3 Stück → Kassenbeleg-V1.
	teilRefs := positionRefsAusSession(t, u, ksNr, u.tischID, 1)
	if err := u.tisch.ZahlungKassieren(ctx, u.userID, "test", u.tischID, teilRefs, ""); err != nil {
		t.Fatalf("Teilzahlung: %v", err)
	}
	teilZahlungID := eventIDByType(t, db, string(kasse.EventTypeZahlungKassiertV1), tischSubject)
	z := warteAufSignatur(t, db, teilZahlungID)
	pruefeSignatur(t, "Teilzahlung", z, tse.ProcessTypeKassenbelegV1)

	// (4) Vollzahlung: die restlichen 2 Stück → Kassenbeleg-V1. Ein weiteres
	// zahlung-kassiert:v1-Event auf demselben Subject (höhere ID).
	vollRefs := positionRefsAusSession(t, u, ksNr, u.tischID, 2)
	if err := u.tisch.ZahlungKassieren(ctx, u.userID, "test", u.tischID, vollRefs, ""); err != nil {
		t.Fatalf("Vollzahlung: %v", err)
	}
	vollZahlungID := eventIDByType(t, db, string(kasse.EventTypeZahlungKassiertV1), tischSubject)
	if vollZahlungID == teilZahlungID {
		t.Fatalf("Vollzahlung erzeugte kein neues zahlung-kassiert-Event")
	}
	z = warteAufSignatur(t, db, vollZahlungID)
	pruefeSignatur(t, "Vollzahlung", z, tse.ProcessTypeKassenbelegV1)

	// (5) Warenrücknahme: Storno von 1 bezahlten Stück → kassenwirksame
	// stornierung-erteilt:v1 (Kassenbeleg-V1, negativ).
	stornoRefs := []kasse.PositionRef{{PositionID: teilRefs[0].PositionID, Menge: 1}}
	if err := u.tisch.StornierungErteilen(ctx, u.userID, "test", u.tischID, stornoRefs, "Rücknahme"); err != nil {
		t.Fatalf("StornierungErteilen (Warenrücknahme): %v", err)
	}
	signiereUndPruefe("Warenrücknahme", string(kasse.EventTypeStornierungErteiltV1), tischSubject, tse.ProcessTypeKassenbelegV1)

	// (6) Geldneutrale Korrektur: Bestellung auf Tisch 2, unbezahlt stornieren →
	// bestellung-korrigiert:v1 (Bestellung-V1, negative Mengen).
	bestellung2ID := uuid.NewString()
	if err := u.tisch.BestellungAufnehmen(ctx, u.userID, "test", bestellung2ID, u.tischID2, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen Tisch 2: %v", err)
	}
	tisch2Subject := kasse.TischSessionSubject(ksNr, u.tischID2)
	korrekturRefs := positionRefsAusSession(t, u, ksNr, u.tischID2, 1)
	if err := u.tisch.StornierungErteilen(ctx, u.userID, "test", u.tischID2, korrekturRefs, "Korrektur"); err != nil {
		t.Fatalf("StornierungErteilen (Korrektur): %v", err)
	}
	signiereUndPruefe("Geldneutrale Korrektur", string(kasse.EventTypeBestellungKorrigiertV1), tisch2Subject, tse.ProcessTypeBestellungV1)

	// (7) Umbuchung: verbleibende unbezahlte Positionen von Tisch 2 auf Tisch 1 →
	// bestellung-umgebucht:v1 auf beiden Seiten (Bestellung-V1). Geprüft: Abgang
	// vom Quelltisch (negative Mengen) und Zugang auf dem Zieltisch.
	umbuchRefs := positionRefsAusSession(t, u, ksNr, u.tischID2, 1)
	if err := u.tisch.BestellungUmbuchen(ctx, u.userID, "test", u.tischID2, u.tischID, umbuchRefs, ""); err != nil {
		t.Fatalf("BestellungUmbuchen: %v", err)
	}
	signiereUndPruefe("Umbuchung Abgang", string(kasse.EventTypeBestellungUmgebuchtV1), tisch2Subject, tse.ProcessTypeBestellungV1)
	signiereUndPruefe("Umbuchung Zugang", string(kasse.EventTypeBestellungUmgebuchtV1), tischSubject, tse.ProcessTypeBestellungV1)

	// (8) Direktverkauf → Kassenbeleg-V1.
	verkaufID := uuid.NewString()
	verkaufInputs := []direktverkaufApp.VerkaufPositionInput{{ProduktID: u.produkt, VarianteID: u.variante, Menge: 2}}
	if err := u.direkt.DirektverkaufTaetigen(ctx, u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	signiereUndPruefe("Direktverkauf", string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject, tse.ProcessTypeKassenbelegV1)

	// (9) Direktverkauf-Storno → Kassenbeleg-V1 (negativ).
	dvSession, err := kassenjournal_repo.NewRepository(db).ReadEventsBySubject(ctx, verkaufSubject)
	if err != nil || len(dvSession) == 0 {
		t.Fatalf("Direktverkauf-Events lesen: %v", err)
	}
	nichtStorniert, err := kasse.ComputeNichtStornierteVerkaufPositionen(dvSession)
	if err != nil {
		t.Fatalf("nicht-stornierte Positionen: %v", err)
	}
	dvStornoRefs := []kasse.PositionRef{{PositionID: nichtStorniert[0].PositionID, Menge: 1}}
	if err := u.direkt.DirektverkaufStornieren(ctx, u.userID, "test", verkaufID, dvStornoRefs, "Rücknahme"); err != nil {
		t.Fatalf("DirektverkaufStornieren: %v", err)
	}
	signiereUndPruefe("Direktverkauf-Storno", string(kasse.EventTypeDirektverkaufStorniertV1), verkaufSubject, tse.ProcessTypeKassenbelegV1)

	// (10) Geldtransit (Einlage) → Kassenbeleg-V1 (Eigenbeleg, 0-%-Feld).
	geldtransitID := uuid.NewString()
	if err := u.kasse.GeldtransitBuchen(ctx, u.userID, "test", geldtransitID, "einlage", 1000, "Wechselgeld"); err != nil {
		t.Fatalf("GeldtransitBuchen: %v", err)
	}
	signiereUndPruefe("Geldtransit", string(kasse.EventTypeGeldtransitGebuchtV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// Tisch-Saldo-Sperre des Kassenabschlusses: alle Tisch-Sessions müssen
	// saldo_cents = 0 haben. Nach Umbuchung/Korrektur tragen beide Tische noch
	// unbezahlte Reste — diese abkassieren (weitere Kassenbeleg-V1-Signaturen,
	// vom Worker mitgesigniert), damit der Abschluss laufen kann.
	restBezahlen(t, u, ksNr, u.tischID)
	restBezahlen(t, u, ksNr, u.tischID2)

	// Vor dem Kassenabschluss müssen alle offenen Aufträge signiert sein: das
	// Signatur-Gate blockiert sonst mit *SignaturenAusstehendError. Alle
	// obigen Vorfälle wurden bereits einzeln bis 'erledigt' abgewartet;
	// zur Sicherheit prüfen, dass kein Auftrag mehr offen ist.
	warteBisKeineOffenenAuftraege(t, db)

	// (11) Kassenabschluss in einem Schritt: Kassensturz (nicht signaturpflichtig),
	// Differenzbuchung (Kassendifferenz, Kassenbeleg-V1) und Tagesabschluss
	// (SonstigerVorgang, Z-Bon). Ist-Bestand bewusst abweichend, damit eine
	// Differenz gebucht wird.
	sollBestand := aktuellerSollBestand(t, db, ksNr)
	istBestand := sollBestand - 137 // 1,37 € Fehlbetrag erzwingt eine Differenzbuchung
	if _, err := u.kasse.KasseAbschliessen(ctx, u.userID, "test", istBestand); err != nil {
		t.Fatalf("KasseAbschliessen: %v", err)
	}

	// Kassendifferenz (Differenzbuchung) → Kassenbeleg-V1 (Eigenbeleg).
	signiereUndPruefe("Kassendifferenz", string(kasse.EventTypeDifferenzSollIstGebuchtV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// Tagesabschluss (Z-Bon) → SonstigerVorgang.
	signiereUndPruefe("Tagesabschluss", string(kasse.EventTypeTagesabschlussErstelltV1), ksSubject, tse.ProcessTypeSonstigerVorgang)

	// Stammdaten-Vollständigkeit: die fiskalischen TSS-Stammdaten (DSFinV-K
	// tse.csv) müssen von der TSS-Ressource lesbar sein. serial_number liegt auf
	// der TSS-Ressource selbst (nicht tss_serial_number) — Lektion aus einem
	// früheren Bug. Wir lesen sie über den Setup-Client und persistieren sie.
	pruefeStammdatenVollstaendigkeit(t, u, credentials)
}

// warteBisKeineOffenenAuftraege stellt sicher, dass der Worker die Queue leer
// gearbeitet hat, bevor der Kassenabschluss läuft (das Signatur-Gate blockiert
// auf frische offene Aufträge).
func warteBisKeineOffenenAuftraege(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(signaturWartefrist)
	for {
		var offen int
		if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege WHERE status = 'offen'").Scan(&offen); err != nil {
			t.Fatalf("offene Aufträge zählen: %v", err)
		}
		if offen == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("noch %d offene Signaturaufträge nach %s", offen, signaturWartefrist)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// aktuellerSollBestand liest den Soll-Kassenbestand der Sitzung.
func aktuellerSollBestand(t *testing.T, db *sql.DB, ksNr int) int {
	t.Helper()
	bestand, err := kassenjournal_repo.NewRepository(db).GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Kassenbestand lesen: %v", err)
	}
	return bestand
}

// pruefeStammdatenVollstaendigkeit liest die TSS-Stammdaten real von fiskaly,
// persistiert sie und prüft explizit, dass alle DSFinV-K-Felder gefüllt sind:
// Signaturalgorithmus, Public Key, Zertifikat, Log-Time-Format und die
// Seriennummer (serial_number der TSS-Ressource).
func pruefeStammdatenVollstaendigkeit(t *testing.T, u *liveTestUmgebung, credentials tse.Credentials) {
	t.Helper()
	ctx := context.Background()

	setupClient, err := tse_repo.NewFiskalyTSESetupClient(fiskalyBaseURL(), tse.SetupCredentials{
		ApiKey:    credentials.ApiKey,
		ApiSecret: credentials.ApiSecret,
	}, nil)
	if err != nil {
		t.Fatalf("Setup-Client bauen: %v", err)
	}

	tssStammdaten, err := setupClient.RetrieveTSSStammdaten(ctx, u.tssID)
	if err != nil {
		t.Fatalf("RetrieveTSSStammdaten: %v", err)
	}
	if tssStammdaten.SignaturAlgorithmus == "" {
		t.Error("Stammdaten: leerer Signaturalgorithmus")
	}
	if tssStammdaten.PublicKey == "" {
		t.Error("Stammdaten: leerer Public Key")
	}
	if tssStammdaten.Zertifikat == "" {
		t.Error("Stammdaten: leeres Zertifikat")
	}
	if tssStammdaten.LogTimeFormat == "" {
		t.Error("Stammdaten: leeres Log-Time-Format")
	}
	if tssStammdaten.Seriennummer == "" {
		t.Error("Stammdaten: leere Seriennummer (serial_number der TSS-Ressource)")
	}

	// Persistieren und aus der DB zurücklesen: der DSFinV-K-Export liest die
	// Stammdaten aus tse_stammdaten, also muss die Persistenz vollständig sein.
	stammdaten := tse.NewStammdaten(
		tssStammdaten.Seriennummer,
		tssStammdaten.SignaturAlgorithmus,
		tssStammdaten.PublicKey,
		tssStammdaten.Zertifikat,
		tssStammdaten.LogTimeFormat,
	)
	repo := tse_repo.NewRepository(u.db)
	if err := repo.UpsertTSEStammdaten(ctx, stammdaten); err != nil {
		t.Fatalf("UpsertTSEStammdaten: %v", err)
	}
	gespeichert, err := repo.GetTSEStammdaten(ctx)
	if err != nil {
		t.Fatalf("GetTSEStammdaten: %v", err)
	}
	if gespeichert.Seriennummer == "" || gespeichert.SignaturAlgorithmus == "" ||
		gespeichert.PublicKey == "" || gespeichert.Zertifikat == "" || gespeichert.LogTimeFormat == "" {
		t.Errorf("persistierte Stammdaten unvollständig: %+v", gespeichert)
	}
}
