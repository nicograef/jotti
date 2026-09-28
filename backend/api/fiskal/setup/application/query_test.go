package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/domain/tse/tsetest"
)

type stubTSERepo struct {
	konfiguration tse.Konfiguration
	identitaet    tse.Kassenidentitaet
}

func (s stubTSERepo) GetKassenidentitaet(context.Context) (tse.Kassenidentitaet, error) {
	return s.identitaet, nil
}

func (s stubTSERepo) GetTSEKonfiguration(context.Context) (tse.Konfiguration, error) {
	return s.konfiguration, nil
}

func konfiguriert() tse.Konfiguration {
	return tse.Konfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}
}

func testerLiefert(status tse.VerbindungStatus) NewTSEConnectionTester {
	return func(tse.Credentials) (tse.ConnectionTester, error) {
		return tsetest.FakeClient{ConnectionResponse: status}, nil
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

// A client whose serial_number equals the Kassen-Seriennummer counts as matching.
func TestCheckTSESetup_ErkenntPassendenClient(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		TSERepo: stubTSERepo{
			identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer},
		},
		NewTSESetupClient: setupClientLiefert(&tsetest.FakeSetupClient{
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

	befund, err := q.CheckTSESetup(context.Background(), gueltigeZugangsdaten())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if befund.Umgebung != "TEST" {
		t.Errorf("expected TEST environment, got %q", befund.Umgebung)
	}
	if len(befund.VorhandeneTSS) != 2 {
		t.Fatalf("expected two TSS in befund, got %d", len(befund.VorhandeneTSS))
	}
	tss1 := befund.VorhandeneTSS[0]
	if tss1.State != "INITIALIZED" {
		t.Errorf("expected TSS state INITIALIZED, got %q", tss1.State)
	}
	if tss1.PassenderClient == nil || tss1.PassenderClient.ID != "client-passt" {
		t.Errorf("expected matching client client-passt, got %+v", tss1.PassenderClient)
	}
	if befund.VorhandeneTSS[1].PassenderClient != nil {
		t.Errorf("expected no matching client for tss-2, got %+v", befund.VorhandeneTSS[1].PassenderClient)
	}
}

// ErrTSESetupZugangsdaten drives the wizard's readable error message.
func TestCheckTSESetup_FalscheZugangsdaten(t *testing.T) {
	q := Query{
		TSERepo: stubTSERepo{},
		NewTSESetupClient: setupClientLiefert(&tsetest.FakeSetupClient{
			TSSErr: tse.ErrSetupAuthFehlgeschlagen,
		}),
	}

	_, err := q.CheckTSESetup(context.Background(), gueltigeZugangsdaten())
	if !errors.Is(err, ErrTSESetupZugangsdaten) {
		t.Errorf("expected ErrTSESetupZugangsdaten, got %v", err)
	}
}

// An empty account yields an empty, non-nil TSS slice and no error.
func TestCheckTSESetup_LeeresKonto(t *testing.T) {
	q := Query{
		TSERepo: stubTSERepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}},
		NewTSESetupClient: setupClientLiefert(&tsetest.FakeSetupClient{
			UmgebungResponse: tse.UmgebungLive,
		}),
	}

	befund, err := q.CheckTSESetup(context.Background(), gueltigeZugangsdaten())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if befund.Umgebung != "LIVE" {
		t.Errorf("expected LIVE environment, got %q", befund.Umgebung)
	}
	if len(befund.VorhandeneTSS) != 0 {
		t.Errorf("expected no TSS for an empty account, got %d", len(befund.VorhandeneTSS))
	}
}

// The fake fails TestConnection on purpose: the status must use the light tester.Umgebung path.
func TestGetTSEStatus_NutztLeichtenUmgebungsPfad(t *testing.T) {
	q := Query{
		TSERepo: stubTSERepo{konfiguration: konfiguriert()},
		NewTSEConnectionTester: func(tse.Credentials) (tse.ConnectionTester, error) {
			return tsetest.FakeClient{
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
		t.Error("expected IstKonfiguriert to be true")
	}
	if status.Umgebung != "LIVE" {
		t.Errorf("expected LIVE environment, got %q", status.Umgebung)
	}
}

func TestTestTSEVerbindung_SeriennummerAbweichung(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		TSERepo: stubTSERepo{
			konfiguration: konfiguriert(),
			identitaet:    tse.Kassenidentitaet{Seriennummer: seriennummer},
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
		t.Error("expected SeriennummerKorrekt to be false for a deviating serial number")
	}
}

func TestTestTSEVerbindung_SeriennummerStimmtUeberein(t *testing.T) {
	seriennummer := uuid.New()
	q := Query{
		TSERepo: stubTSERepo{
			konfiguration: konfiguriert(),
			identitaet:    tse.Kassenidentitaet{Seriennummer: seriennummer},
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
		t.Error("expected SeriennummerKorrekt to be true for a matching serial number")
	}
}
