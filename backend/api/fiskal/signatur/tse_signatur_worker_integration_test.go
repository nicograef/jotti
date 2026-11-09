//go:build integration

package signatur

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/domain/tse/tsetest"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// While another session holds the advisory lock, the worker is refused without failing fast.
// After release, the next tick's retry acquires it.
func TestTSESignaturWorker_AdvisoryLock_ZweiteSessionHaeltLock(t *testing.T) {
	db := dbtest.Open()
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	// First "instance": its own pinned session holds the lock.
	halter, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Halter-Connection oeffnen: %v", err)
	}
	t.Cleanup(func() { _ = halter.Close() })

	var gehalten bool
	if err := halter.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&gehalten); err != nil {
		t.Fatalf("Halter-Lock erwerben: %v", err)
	}
	if !gehalten {
		t.Fatal("Halter-Session hat den Lock nicht bekommen — haelt ihn ein anderer Prozess?")
	}

	worker := &tseSignaturWorker{lockDB: db}
	defer worker.releaseLock()

	// Second instance waits: ensureLock returns false, the app keeps running.
	if worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock erhalten, obwohl eine zweite Session ihn haelt")
	}

	// The holder releases; the next tick's retry acquires the lock.
	if _, err := halter.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", tseSignaturWorkerLockKey); err != nil {
		t.Fatalf("Halter-Lock freigeben: %v", err)
	}

	if !worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock nach der Freigabe nicht erworben")
	}

	// The lock is session-scoped: a further session is refused.
	var frei bool
	if err := db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&frei); err != nil {
		t.Fatalf("Kontroll-Lock pruefen: %v", err)
	}
	if frei {
		t.Error("Lock war trotz haltendem Worker frei erwerbbar")
	}
}

// zaehlenderClient counts fiskaly calls to prove a run aborts or the outage state stays silent.
type zaehlenderClient struct {
	tsetest.FakeClient
	mu    sync.Mutex
	calls int
}

func (c *zaehlenderClient) zaehle() {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
}

func (c *zaehlenderClient) anzahlCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (c *zaehlenderClient) RetrieveTransaction(ctx context.Context, txID string) (tse.RetrieveResult, error) {
	c.zaehle()
	return c.FakeClient.RetrieveTransaction(ctx, txID)
}

func (c *zaehlenderClient) StartTransaction(ctx context.Context, txID string) (tse.StartResult, error) {
	c.zaehle()
	return c.FakeClient.StartTransaction(ctx, txID)
}

func (c *zaehlenderClient) FinishTransaction(ctx context.Context, txID string, processType string, processData string) (tse.FinishResult, error) {
	c.zaehle()
	return c.FakeClient.FinishTransaction(ctx, txID, processType, processData)
}

// txAbhaengigerClient fails only the txIDs it holds an error for, for poison-job tests.
type txAbhaengigerClient struct {
	ablehnungen map[string]error
}

func (c txAbhaengigerClient) RetrieveTransaction(_ context.Context, _ string) (tse.RetrieveResult, error) {
	return tse.RetrieveResult{}, tse.ErrTransactionNichtGefunden
}

func (c txAbhaengigerClient) StartTransaction(_ context.Context, txID string) (tse.StartResult, error) {
	if err, ok := c.ablehnungen[txID]; ok {
		return tse.StartResult{}, err
	}
	return tse.StartResult{TransactionNumber: 60, LogTime: time.Date(2026, 6, 10, 20, 0, 1, 0, time.UTC)}, nil
}

func (c txAbhaengigerClient) FinishTransaction(_ context.Context, txID string, _ string, _ string) (tse.FinishResult, error) {
	if err, ok := c.ablehnungen[txID]; ok {
		return tse.FinishResult{}, err
	}
	return tse.FinishResult{TransactionNumber: 60, SignatureCounter: 900, SerialNumberTSE: "TSE-SN", LogTimeEnd: time.Date(2026, 6, 10, 20, 0, 2, 0, time.UTC), Signature: "SIG-" + txID}, nil
}

// workerUmgebung is a configured TSE with one offen Kassensitzung; the worker talks to the real tse_repo.
type workerUmgebung struct {
	db      *sql.DB
	repo    tse_repo.Repository
	userID  int
	ksNr    int
	version int
}

func setupWorkerUmgebung(t *testing.T) *workerUmgebung {
	t.Helper()
	database := dbtest.Open()
	reset := func() {
		for _, stmt := range []string{
			"DELETE FROM tse_stoerungen",
			"DELETE FROM tse_signaturauftraege",
			"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassenjournal",
			"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
			"DELETE FROM kassensitzungen",
			"DELETE FROM users",
			"UPDATE tse_konfiguration SET api_key = '', api_secret = '', tss_id = '', client_id = '' WHERE id = 1",
		} {
			if _, err := database.Exec(stmt); err != nil {
				t.Fatalf("reset %q: %v", stmt, err)
			}
		}
	}
	reset()
	t.Cleanup(func() {
		reset()
		_ = database.Close()
	})

	u := &workerUmgebung{db: database, repo: tse_repo.NewRepository(database)}
	// Saved before the Kassensitzung opens, as the TSE setup requires.
	if err := u.repo.SaveEinrichtung(context.Background(), configuredTSE()); err != nil {
		t.Fatalf("save konfiguration: %v", err)
	}
	if err := database.QueryRow(
		"INSERT INTO users (name, username, role, status, created_at, updated_at) VALUES ('Test', 'worker-test', 'admin', 'active', NOW(), NOW()) RETURNING id",
	).Scan(&u.userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := database.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&u.ksNr); err != nil {
		t.Fatalf("insert kassensitzung: %v", err)
	}
	return u
}

// auftrag enqueues an offen Signaturauftrag for its own Kassenjournal event.
func (u *workerUmgebung) auftrag(t *testing.T, txID string) int {
	t.Helper()
	u.version++
	var eventID int
	if err := u.db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'Test', 'zahlung-kassiert:v1', $2, $3, '{}', NOW(), $4) RETURNING id",
		u.userID, fmt.Sprintf("kassensitzung-%d/tisch-1", u.ksNr), u.version, u.ksNr,
	).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	var id int
	if err := u.db.QueryRow(`
		INSERT INTO tse_signaturauftraege (event_id, tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
		VALUES ($1, $2, 'Kassenbeleg-V1', 'Beleg^1.00', 'offen', NOW(), NOW())
		RETURNING id
	`, eventID, txID).Scan(&id); err != nil {
		t.Fatalf("insert auftrag: %v", err)
	}
	return id
}

type auftragZeile struct {
	status   string
	versuche int
	signatur string
}

func (u *workerUmgebung) zeile(t *testing.T, id int) auftragZeile {
	t.Helper()
	var z auftragZeile
	var signatur sql.NullString
	if err := u.db.QueryRow("SELECT status, versuche, signatur FROM tse_signaturauftraege WHERE id = $1", id).Scan(&z.status, &z.versuche, &signatur); err != nil {
		t.Fatalf("read auftrag %d: %v", id, err)
	}
	z.signatur = signatur.String
	return z
}

// faelligMachen skips the job backoff the SQL sets after a failed attempt.
func (u *workerUmgebung) faelligMachen(t *testing.T) {
	t.Helper()
	if _, err := u.db.Exec("UPDATE tse_signaturauftraege SET naechster_versuch_am = NOW() WHERE status = 'offen'"); err != nil {
		t.Fatalf("make due: %v", err)
	}
}

func (u *workerUmgebung) aktiveStoerung(t *testing.T) *tse.Stoerung {
	t.Helper()
	stoerung, err := u.repo.GetAktiveTSEStoerung(context.Background())
	if err != nil {
		t.Fatalf("read aktive stoerung: %v", err)
	}
	return stoerung
}

// A TSE-wide error (503) aborts the run at the first job without counting attempts and logs a tse_fehler
// period. No fiskaly call happens during the backoff; the half-open probe that fails grows it. The first
// success re-signs the whole backlog and closes the period.
func TestTSESignaturWorker_StoerungUndNachsignierung(t *testing.T) {
	u := setupWorkerUmgebung(t)
	ersterID := u.auftrag(t, "tx-stoerung-1")
	zweiterID := u.auftrag(t, "tx-stoerung-2")
	client := &zaehlenderClient{FakeClient: tsetest.FakeClient{RetrieveErr: errors.New("fiskaly api error 503")}}
	jetzt := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)
	worker := &tseSignaturWorker{
		settingsRepo: u.repo,
		store:        u.repo,
		newTSEClient: func(tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          func() time.Time { return jetzt },
	}
	ctx := context.Background()

	// Run 1: abort after the first job, both stay offen with 0 attempts, the outage is logged.
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected a TSE-wide run error")
	}
	if client.anzahlCalls() != 1 {
		t.Errorf("expected abort after 1 fiskaly call, got %d", client.anzahlCalls())
	}
	for _, id := range []int{ersterID, zweiterID} {
		if z := u.zeile(t, id); z.status != tse.StatusOffen || z.versuche != 0 {
			t.Errorf("auftrag %d = %+v, want offen with 0 versuche", id, z)
		}
	}
	if s := u.aktiveStoerung(t); s == nil || s.GrundArt != tse.StoerungGrundTSEFehler {
		t.Fatalf("expected an active tse_fehler stoerung, got %+v", s)
	}
	if !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(5 * time.Second)) {
		t.Errorf("expected next try after 5s backoff, got %v", worker.stoerungNaechsterVersuch.Sub(jetzt))
	}

	// During the backoff no fiskaly call.
	jetzt = jetzt.Add(2 * time.Second)
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected a gated run without error, got %v", err)
	}
	if client.anzahlCalls() != 1 {
		t.Errorf("expected no fiskaly call during the backoff, got %d calls", client.anzahlCalls())
	}

	// The failing probe makes one call and grows the backoff to 10s.
	jetzt = jetzt.Add(4 * time.Second)
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected the probe to fail TSE-wide")
	}
	if client.anzahlCalls() != 2 {
		t.Errorf("expected exactly one probe call, got %d calls", client.anzahlCalls())
	}
	if worker.stoerungSerie != 2 || !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(10*time.Second)) {
		t.Errorf("expected serie 2 and 10s backoff, got serie %d backoff %v", worker.stoerungSerie, worker.stoerungNaechsterVersuch.Sub(jetzt))
	}

	// fiskaly is back: the backlog signs and the period closes.
	client.FakeClient = tsetest.FakeClient{
		RetrieveErr:    tse.ErrTransactionNichtGefunden,
		StartResponse:  tse.StartResult{TransactionNumber: 70, LogTime: jetzt},
		FinishResponse: tse.FinishResult{TransactionNumber: 70, SignatureCounter: 900, SerialNumberTSE: "TSE-SN", LogTimeEnd: jetzt, Signature: "SIG-NACH"},
	}
	jetzt = jetzt.Add(11 * time.Second)
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected the recovery run without error, got %v", err)
	}
	for _, id := range []int{ersterID, zweiterID} {
		if z := u.zeile(t, id); z.status != tse.StatusErledigt || z.signatur != "SIG-NACH" {
			t.Errorf("auftrag %d = %+v, want erledigt with SIG-NACH", id, z)
		}
	}
	if s := u.aktiveStoerung(t); s != nil {
		t.Errorf("expected the stoerung closed, got %+v", s)
	}
	zeitraeume, err := u.repo.GetAlleTSEStoerungen(ctx)
	if err != nil {
		t.Fatalf("read stoerungen: %v", err)
	}
	if len(zeitraeume) != 1 || zeitraeume[0].GrundArt != tse.StoerungGrundTSEFehler || zeitraeume[0].Ende == nil {
		t.Errorf("expected one closed tse_fehler period, got %+v", zeitraeume)
	}
	if worker.stoerungSerie != 0 || !worker.stoerungNaechsterVersuch.IsZero() {
		t.Errorf("expected a reset outage state, got serie %d next %v", worker.stoerungSerie, worker.stoerungNaechsterVersuch)
	}
}

// A job-specific error counts an attempt and skips the job, so the next one still signs in the same run;
// the MaxSignaturVersuche-th failure makes it fehlgeschlagen for good. No outage is logged.
func TestTSESignaturWorker_AuftragsFehlerBisFehlgeschlagen(t *testing.T) {
	u := setupWorkerUmgebung(t)
	giftID := u.auftrag(t, "tx-gift")
	okID := u.auftrag(t, "tx-ok")
	client := txAbhaengigerClient{ablehnungen: map[string]error{
		"tx-gift": tse.AuftragsFehler{Err: errors.New("fiskaly api error 400 (E_FAILED_SCHEMA_VALIDATION)")},
	}}
	worker := &tseSignaturWorker{
		settingsRepo: u.repo,
		store:        u.repo,
		newTSEClient: func(tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          time.Now,
	}
	ctx := context.Background()

	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("run 1: expected no run error, got %v", err)
	}
	if z := u.zeile(t, giftID); z.status != tse.StatusOffen || z.versuche != 1 {
		t.Errorf("run 1: poison job = %+v, want offen with 1 versuch", z)
	}
	if z := u.zeile(t, okID); z.status != tse.StatusErledigt || z.signatur != "SIG-tx-ok" {
		t.Errorf("run 1: next job = %+v, want erledigt with SIG-tx-ok", z)
	}

	for lauf := 2; lauf <= tse_repo.MaxSignaturVersuche; lauf++ {
		u.faelligMachen(t)
		if err := worker.processOnce(ctx); err != nil {
			t.Fatalf("run %d: expected no run error, got %v", lauf, err)
		}
	}
	if z := u.zeile(t, giftID); z.status != tse.StatusFehlgeschlagen || z.versuche != 3 {
		t.Errorf("poison job = %+v, want fehlgeschlagen with 3 versuche", z)
	}
	if s := u.aktiveStoerung(t); s != nil {
		t.Errorf("expected no stoerung for job-specific errors, got %+v", s)
	}
}
