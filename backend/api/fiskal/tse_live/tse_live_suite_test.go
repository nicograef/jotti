//go:build integration

// Package tse_live signs every signature-bound business event through the real services and signing worker
// against the fiskaly TEST TSS; any non-TEST environment aborts the run.
// Run with `make test-tse-live`.
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
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	kassenfuehrungApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/db/dbtest"
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

// signaturWartefrist is generous because a real fiskaly round trip including a 429 backoff takes seconds.
const signaturWartefrist = 90 * time.Second

type liveTestUmgebung struct {
	db         *sql.DB
	tisch      tischgeschaeftApp.Command
	direkt     direktverkaufApp.Command
	kasse      kassenfuehrungApp.Command
	tssID      string
	userID     int
	tischID    int
	tischID2   int
	produktID  int
	varianteID int
}

// credentialsOderSkip skips unless JOTTI_TSE_LIVE=1 and the FISKALY_TEST_* credentials are set.
// The opt-in keeps regular integration runs hermetic even with FISKALY_TEST_* exported.
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

// pruefeTestUmgebungOderAbbruch aborts unless the credentials point at the TEST environment:
// the suite must never sign against a LIVE TSS.
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

// cleanLiveDB briefly disables the delete trigger that keeps kassenjournal append-only.
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

func setupLiveUmgebung(t *testing.T, credentials tse.Credentials) *liveTestUmgebung {
	t.Helper()

	db := dbtest.Open()
	cleanLiveDB(t, db)
	t.Cleanup(func() {
		cleanLiveDB(t, db)
		_ = db.Close()
	})

	ctx := context.Background()

	// The signing worker reads its credentials from this singleton configuration.
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
	).Scan(&u.produktID); err != nil {
		t.Fatalf("create produkt: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.5L', 350, 'active', now(), now()) RETURNING id",
		u.produktID,
	).Scan(&u.varianteID); err != nil {
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

// warteAufSignatur polls until the event's job is erledigt; a failed job aborts at once instead of at the timeout.
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

// pruefeSignatur checks a signed job's data and its processType, which the fiscal projection sets
// per DSFinV-K 2.4 Anhang I. See docs/compliance.md §3.3.
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

// eventIDByType returns the newest event of the type, since one subject may hold several (partial and full payment).
// The suite therefore reads it right after each business event.
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

// positionRefsAusSession refers to menge units of the table's first unpaid position.
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

// restBezahlen pays every unpaid position so the table reaches saldo_cents = 0, which the cash close requires.
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

// TestTSELiveSuite_GeschaeftsvorfaelleUndStammdaten signs each signature-bound business event live and checks
// its signature data and processType mapping, then the completeness of the persisted TSS master data.
func TestTSELiveSuite_GeschaeftsvorfaelleUndStammdaten(t *testing.T) {
	credentials := credentialsOderSkip(t)
	pruefeTestUmgebungOderAbbruch(t, credentials)

	u := setupLiveUmgebung(t, credentials)
	starteWorker(t, u.db)

	ctx := context.Background()
	db := u.db

	signiereUndPruefe := func(vorfall, eventType, subject, erwarteterProcessType string) {
		id := eventIDByType(t, db, eventType, subject)
		z := warteAufSignatur(t, db, id)
		pruefeSignatur(t, vorfall, z, erwarteterProcessType)
	}

	// (1) Opening a session with a float > 0 → Kassenbeleg-V1 (Bareinlage).
	ksNr, err := u.kasse.KassensitzungEroeffnen(ctx, u.userID, "test", "Live-Suite", 5000)
	if err != nil {
		t.Fatalf("KassensitzungEroeffnen: %v", err)
	}
	ksSubject := kasse.KassensitzungSubject(ksNr)
	signiereUndPruefe("Kassensitzung-Eröffnung (Bareinlage)", string(kasse.EventTypeKassensitzungEroeffnetV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// (2) Order → Bestellung-V1; 3 units leave quantities for partial payment, full payment and cancellation.
	bestellungID := uuid.NewString()
	inputs := []enrichment.PositionInput{{ProduktID: u.produktID, VarianteID: u.varianteID, Menge: 3}}
	if err := u.tisch.BestellungAufnehmen(ctx, u.userID, "test", bestellungID, u.tischID, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen: %v", err)
	}
	tischSubject := kasse.TischSessionSubject(ksNr, u.tischID)
	signiereUndPruefe("Bestellung", string(kasse.EventTypeBestellungAufgenommenV1), tischSubject, tse.ProcessTypeBestellungV1)

	// (3) Partial payment of 1 of 3 units → Kassenbeleg-V1.
	teilRefs := positionRefsAusSession(t, u, ksNr, u.tischID, 1)
	if err := u.tisch.ZahlungKassieren(ctx, u.userID, "test", u.tischID, teilRefs, ""); err != nil {
		t.Fatalf("Teilzahlung: %v", err)
	}
	teilZahlungID := eventIDByType(t, db, string(kasse.EventTypeZahlungKassiertV1), tischSubject)
	z := warteAufSignatur(t, db, teilZahlungID)
	pruefeSignatur(t, "Teilzahlung", z, tse.ProcessTypeKassenbelegV1)

	// (4) Full payment of the remaining 2 units → Kassenbeleg-V1, a second zahlung-kassiert:v1 on the same subject.
	vollRefs := positionRefsAusSession(t, u, ksNr, u.tischID, 2)
	if err := u.tisch.ZahlungKassieren(ctx, u.userID, "test", u.tischID, vollRefs, ""); err != nil {
		t.Fatalf("Vollzahlung: %v", err)
	}
	vollZahlungID := eventIDByType(t, db, string(kasse.EventTypeZahlungKassiertV1), tischSubject)
	if vollZahlungID == teilZahlungID {
		t.Errorf("Vollzahlung erzeugte kein neues zahlung-kassiert-Event")
	}
	z = warteAufSignatur(t, db, vollZahlungID)
	pruefeSignatur(t, "Vollzahlung", z, tse.ProcessTypeKassenbelegV1)

	// (5) Return of 1 paid unit → cash-relevant stornierung-erteilt:v1 (negative Kassenbeleg-V1).
	stornoRefs := []kasse.PositionRef{{PositionID: teilRefs[0].PositionID, Menge: 1}}
	if err := u.tisch.StornierungErteilen(ctx, u.userID, "test", u.tischID, stornoRefs, "Rücknahme"); err != nil {
		t.Fatalf("StornierungErteilen (Warenrücknahme): %v", err)
	}
	signiereUndPruefe("Warenrücknahme", string(kasse.EventTypeStornierungErteiltV1), tischSubject, tse.ProcessTypeKassenbelegV1)

	// (6) Cancelling an unpaid order on table 2 → cash-neutral bestellung-korrigiert:v1 (Bestellung-V1, negative quantities).
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

	// (7) Transfer from table 2 to table 1 → bestellung-umgebucht:v1 on both subjects (Bestellung-V1),
	// negative quantities on the source table.
	umbuchRefs := positionRefsAusSession(t, u, ksNr, u.tischID2, 1)
	if err := u.tisch.BestellungUmbuchen(ctx, u.userID, "test", u.tischID2, u.tischID, umbuchRefs, ""); err != nil {
		t.Fatalf("BestellungUmbuchen: %v", err)
	}
	signiereUndPruefe("Umbuchung Abgang", string(kasse.EventTypeBestellungUmgebuchtV1), tisch2Subject, tse.ProcessTypeBestellungV1)
	signiereUndPruefe("Umbuchung Zugang", string(kasse.EventTypeBestellungUmgebuchtV1), tischSubject, tse.ProcessTypeBestellungV1)

	// (8) Direct sale → Kassenbeleg-V1.
	verkaufID := uuid.NewString()
	verkaufInputs := []enrichment.PositionInput{{ProduktID: u.produktID, VarianteID: u.varianteID, Menge: 2}}
	if err := u.direkt.DirektverkaufTaetigen(ctx, u.userID, "test", verkaufID, verkaufInputs, ""); err != nil {
		t.Fatalf("DirektverkaufTaetigen: %v", err)
	}
	verkaufSubject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	signiereUndPruefe("Direktverkauf", string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufSubject, tse.ProcessTypeKassenbelegV1)

	// (9) Direct-sale cancellation → negative Kassenbeleg-V1.
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

	// (10) Cash deposit → Kassenbeleg-V1 (Eigenbeleg, 0 % field).
	geldtransitID := uuid.NewString()
	if err := u.kasse.GeldtransitBuchen(ctx, u.userID, "test", geldtransitID, "einlage", 1000, "Wechselgeld"); err != nil {
		t.Fatalf("GeldtransitBuchen: %v", err)
	}
	signiereUndPruefe("Geldtransit", string(kasse.EventTypeGeldtransitGebuchtV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// Transfer and correction leave unpaid rests on both tables, which would block the cash close.
	restBezahlen(t, u, ksNr, u.tischID)
	restBezahlen(t, u, ksNr, u.tischID2)

	// Unsigned jobs would block the close with *SignaturenAusstehendError.
	warteBisKeineOffenenAuftraege(t, db)

	// (11) The cash close books the Kassensturz (unsigned), the difference (Kassenbeleg-V1) and the Z-Bon
	// (SonstigerVorgang).
	sollBestand := aktuellerSollBestand(t, db, ksNr)
	istBestand := sollBestand - 137 // a 1.37 € shortfall forces a difference booking
	if _, err := u.kasse.KasseAbschliessen(ctx, u.userID, "test", istBestand); err != nil {
		t.Fatalf("KasseAbschliessen: %v", err)
	}

	// Cash difference → Kassenbeleg-V1 (Eigenbeleg).
	signiereUndPruefe("Kassendifferenz", string(kasse.EventTypeDifferenzSollIstGebuchtV1), ksSubject, tse.ProcessTypeKassenbelegV1)

	// Tagesabschluss (Z-Bon) → SonstigerVorgang.
	signiereUndPruefe("Tagesabschluss", string(kasse.EventTypeTagesabschlussErstelltV1), ksSubject, tse.ProcessTypeSonstigerVorgang)

	pruefeStammdatenVollstaendigkeit(t, u, credentials)
}

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

func aktuellerSollBestand(t *testing.T, db *sql.DB, ksNr int) int {
	t.Helper()
	bestand, err := kassenjournal_repo.NewRepository(db).GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Kassenbestand lesen: %v", err)
	}
	return bestand.SollBestandCents
}

// pruefeStammdatenVollstaendigkeit asserts that fiskaly delivers every TSS master data field of DSFinV-K tse.csv.
// The serial number is the TSS resource's serial_number, not tss_serial_number.
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

	// The DSFinV-K export reads tse_stammdaten, so the stored copy must be complete too.
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
