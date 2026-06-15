package tse

import (
	"context"
	"errors"
	"strings"
)

// ErrSetupAuthFehlgeschlagen zeigt an, dass sich die Setup-Operationen mit dem
// uebergebenen API-Key/-Secret nicht authentifizieren konnten — fast immer
// falsche Zugangsdaten. Die Application-Schicht macht daraus eine
// verstaendliche Meldung fuer den Admin.
var ErrSetupAuthFehlgeschlagen = errors.New("tse setup authentication failed")

// SetupCredentials authentifiziert die gefuehrte TSE-Einrichtung. Anders als
// Credentials kommt das Setup ohne TSS-/Client-ID aus: beide entstehen erst im
// Verlauf der Einrichtung.
type SetupCredentials struct {
	ApiKey    string
	ApiSecret string
}

func (c SetupCredentials) Validate() error {
	if strings.TrimSpace(c.ApiKey) == "" || strings.TrimSpace(c.ApiSecret) == "" {
		return ErrUnvollstaendigeCredentials
	}
	return nil
}

// TSSInfo ist der fuer die Einrichtung relevante Ausschnitt einer fiskaly-TSS.
type TSSInfo struct {
	ID    string
	State string
}

// ClientInfo ist der fuer die Einrichtung relevante Ausschnitt eines
// fiskaly-Clients einer TSS.
type ClientInfo struct {
	ID           string
	SerialNumber string
	State        string
}

// SetupClient kapselt die fiskaly-Operationen der gefuehrten TSE-Einrichtung.
// Phase 3 nutzt nur die lesenden Operationen; die schreibenden Lebenszyklus-
// Operationen (TSS anlegen, initialisieren, Client registrieren) kommen spaeter
// dazu. ListTSS liefert die Umgebung aus der fiskaly-Antwort mit, damit der
// Befund TEST/LIVE auch bei leerem Konto anzeigen kann.
type SetupClient interface {
	ListTSS(ctx context.Context) (Umgebung, []TSSInfo, error)
	ListClients(ctx context.Context, tssID string) ([]ClientInfo, error)
}
