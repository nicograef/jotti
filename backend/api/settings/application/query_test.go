//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type stubSettingsRepo struct {
	konfiguration settings.TSEKonfiguration
	identitaet    settings.Kassenidentitaet
}

func (s stubSettingsRepo) GetKassenidentitaet(context.Context) (settings.Kassenidentitaet, error) {
	return s.identitaet, nil
}

func (s stubSettingsRepo) GetBetreiber(context.Context) (settings.Betreiber, error) {
	return settings.Betreiber{}, errors.New("not implemented")
}

func (s stubSettingsRepo) GetTSEKonfiguration(context.Context) (settings.TSEKonfiguration, error) {
	return s.konfiguration, nil
}

type stubTSEStatusRepo struct{ offen int }

func (s stubTSEStatusRepo) CountOffeneTSENachsignierAuftraege(context.Context) (int, error) {
	return s.offen, nil
}

func konfiguriert() settings.TSEKonfiguration {
	return settings.TSEKonfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}
}

func testerLiefert(status tse.VerbindungStatus) NewTSEConnectionTester {
	return func(tse.Credentials) (tse.ConnectionTester, error) {
		return tse.FakeClient{ConnectionResponse: status}, nil
	}
}

func setupClientLiefert(client tse.SetupClient) NewTSESetupClient {
	return func(tse.SetupCredentials) (tse.SetupClient, error) {
		return client, nil
	}
}

func gueltigeZugangsdaten() tse.SetupCredentials {
	return tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}
}

// TestPruefeTSESetup_ErkenntPassendenClient sichert den Kern des Befunds: die
// Umgebung und die vorhandenen TSS werden gemeldet, und ein Client, dessen
// serial_number der Kassen-Seriennummer entspricht, wird als passend erkannt.
func TestPruefeTSESetup_ErkenntPassendenClient(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		SettingsRepo: stubSettingsRepo{
			identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer},
		},
		NewTSESetupClient: setupClientLiefert(&tse.FakeSetupClient{
			UmgebungResponse: tse.UmgebungTest,
			TSSResponse: []tse.TSSInfo{
				{ID: "tss-1", State: "INITIALIZED"},
				{ID: "tss-2", State: "CREATED"},
			},
			ClientsByTSS: map[string][]tse.ClientInfo{
				"tss-1": {
					{ID: "client-fremd", SerialNumber: "andere-serial", State: "REGISTERED"},
					{ID: "client-passt", SerialNumber: seriennummer.String(), State: "REGISTERED"},
				},
			},
		}),
	}

	befund, err := q.PruefeTSESetup(context.Background(), gueltigeZugangsdaten())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if befund.Umgebung != "TEST" {
		t.Fatalf("expected TEST environment, got %q", befund.Umgebung)
	}
	if len(befund.VorhandeneTSS) != 2 {
		t.Fatalf("expected two TSS in befund, got %d", len(befund.VorhandeneTSS))
	}
	tss1 := befund.VorhandeneTSS[0]
	if tss1.State != "INITIALIZED" {
		t.Fatalf("expected TSS state INITIALIZED, got %q", tss1.State)
	}
	if tss1.PassenderClient == nil || tss1.PassenderClient.ID != "client-passt" {
		t.Fatalf("expected matching client client-passt, got %+v", tss1.PassenderClient)
	}
	if befund.VorhandeneTSS[1].PassenderClient != nil {
		t.Fatalf("expected no matching client for tss-2, got %+v", befund.VorhandeneTSS[1].PassenderClient)
	}
}

// TestPruefeTSESetup_FalscheZugangsdaten sichert, dass ein Auth-Fehler des
// Setup-Clients zu ErrTSESetupZugangsdaten wird — der Code für die
// verständliche deutsche Fehlermeldung im Wizard.
func TestPruefeTSESetup_FalscheZugangsdaten(t *testing.T) {
	q := Query{
		SettingsRepo: stubSettingsRepo{},
		NewTSESetupClient: setupClientLiefert(&tse.FakeSetupClient{
			TSSErr: tse.ErrSetupAuthFehlgeschlagen,
		}),
	}

	_, err := q.PruefeTSESetup(context.Background(), gueltigeZugangsdaten())
	if !errors.Is(err, ErrTSESetupZugangsdaten) {
		t.Fatalf("expected ErrTSESetupZugangsdaten, got %v", err)
	}
}

// TestPruefeTSESetup_LeeresKonto sichert, dass ein leeres Konto einen gültigen
// Befund ohne TSS liefert (keine nil-Slice, kein Fehler).
func TestPruefeTSESetup_LeeresKonto(t *testing.T) {
	q := Query{
		SettingsRepo: stubSettingsRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}},
		NewTSESetupClient: setupClientLiefert(&tse.FakeSetupClient{
			UmgebungResponse: tse.UmgebungLive,
		}),
	}

	befund, err := q.PruefeTSESetup(context.Background(), gueltigeZugangsdaten())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if befund.Umgebung != "LIVE" {
		t.Fatalf("expected LIVE environment, got %q", befund.Umgebung)
	}
	if len(befund.VorhandeneTSS) != 0 {
		t.Fatalf("expected no TSS for an empty account, got %d", len(befund.VorhandeneTSS))
	}
}

// TestGetTSEStatus_NutztLeichtenUmgebungsPfad sichert, dass der Status die
// Umgebung ueber den leichten Pfad (tester.Umgebung) bezieht und nicht den
// vollen Verbindungstest (TSS-/Client-Abruf) ausloest: Der Fake laesst
// TestConnection bewusst fehlschlagen, der Status muss trotzdem die Umgebung
// und die offenen Nachsignierungen liefern.
func TestGetTSEStatus_NutztLeichtenUmgebungsPfad(t *testing.T) {
	q := Query{
		SettingsRepo:  stubSettingsRepo{konfiguration: konfiguriert()},
		TSEStatusRepo: stubTSEStatusRepo{offen: 3},
		NewTSEConnectionTester: func(tse.Credentials) (tse.ConnectionTester, error) {
			return tse.FakeClient{
				UmgebungResponse: tse.UmgebungLive,
				ConnectionErr:    errors.New("voller Verbindungstest darf hier nicht laufen"),
			}, nil
		},
	}

	status, err := q.GetTSEStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.IstKonfiguriert {
		t.Fatal("expected IstKonfiguriert to be true")
	}
	if status.Umgebung != "LIVE" {
		t.Fatalf("expected LIVE environment, got %q", status.Umgebung)
	}
	if status.OffeneNachsignierungen != 3 {
		t.Fatalf("expected 3 open retry jobs, got %d", status.OffeneNachsignierungen)
	}
}

// TestTestTSEVerbindung_SeriennummerAbweichung sichert ab, dass eine Client-
// serial_number, die nicht der Kassen-Seriennummer entspricht, als Abweichung
// gemeldet wird (SeriennummerKorrekt = false).
func TestTestTSEVerbindung_SeriennummerAbweichung(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		SettingsRepo: stubSettingsRepo{
			konfiguration: konfiguriert(),
			identitaet:    settings.Kassenidentitaet{Seriennummer: seriennummer},
		},
		NewTSEConnectionTester: testerLiefert(tse.VerbindungStatus{
			Umgebung:           tse.UmgebungTest,
			TSSState:           "INITIALIZED",
			ClientState:        "REGISTERED",
			ClientSerialNumber: "eine-andere-seriennummer",
		}),
	}

	status, err := q.TestTSEVerbindung(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.SeriennummerKorrekt {
		t.Fatal("expected SeriennummerKorrekt to be false for a deviating serial number")
	}
}

// TestTestTSEVerbindung_SeriennummerStimmtUeberein sichert den positiven Fall:
// stimmt die Client-serial_number mit der Kassen-Seriennummer überein, ist der
// Abgleich erfolgreich.
func TestTestTSEVerbindung_SeriennummerStimmtUeberein(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		SettingsRepo: stubSettingsRepo{
			konfiguration: konfiguriert(),
			identitaet:    settings.Kassenidentitaet{Seriennummer: seriennummer},
		},
		NewTSEConnectionTester: testerLiefert(tse.VerbindungStatus{
			Umgebung:           tse.UmgebungTest,
			TSSState:           "INITIALIZED",
			ClientState:        "REGISTERED",
			ClientSerialNumber: seriennummer.String(),
		}),
	}

	status, err := q.TestTSEVerbindung(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.SeriennummerKorrekt {
		t.Fatal("expected SeriennummerKorrekt to be true for a matching serial number")
	}
}
