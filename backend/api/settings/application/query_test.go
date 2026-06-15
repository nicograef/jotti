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
