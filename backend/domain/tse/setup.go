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

// TSSErstellt ist das Ergebnis der TSS-Neuanlage: die ID der frisch erzeugten
// TSS, ihr Zustand (CREATED) und der einmalig von fiskaly gelieferte Admin-PUK.
// Der PUK wird nie persistiert oder geloggt — er fliesst nur durch bis in die
// einmalige Anzeige an den Admin.
type TSSErstellt struct {
	ID    string
	PUK   string
	State string
}

// SetupClient kapselt die fiskaly-Operationen der gefuehrten TSE-Einrichtung:
// die lesenden Operationen des Pruef-Schritts (ListTSS/ListClients) und die
// schreibenden Lebenszyklus-Operationen, mit denen der Orchestrator eine TSS
// von der Neuanlage bis zum registrierten Client treibt. ListTSS liefert die
// Umgebung aus der fiskaly-Antwort mit, damit der Befund TEST/LIVE auch bei
// leerem Konto anzeigen kann.
type SetupClient interface {
	ListTSS(ctx context.Context) (Umgebung, []TSSInfo, error)
	ListClients(ctx context.Context, tssID string) ([]ClientInfo, error)

	// CreateTSS legt eine neue TSS an (Zustand CREATED) und liefert deren
	// einmaligen Admin-PUK zurueck.
	CreateTSS(ctx context.Context) (TSSErstellt, error)
	// PersonalisiereTSS ueberfuehrt die TSS von CREATED nach UNINITIALIZED.
	PersonalisiereTSS(ctx context.Context, tssID string) error
	// SetzeAdminPIN setzt mit dem PUK die Admin-PIN der TSS.
	SetzeAdminPIN(ctx context.Context, tssID, puk, pin string) error
	// AuthentifiziereAdmin hebt das aktuelle Zugriffstoken fuer die folgenden
	// Admin-Operationen der TSS auf Admin-Rechte an.
	AuthentifiziereAdmin(ctx context.Context, tssID, pin string) error
	// InitialisiereTSS ueberfuehrt die TSS nach INITIALIZED (signierbereit).
	InitialisiereTSS(ctx context.Context, tssID string) error
	// RegistriereClient registriert einen Client unter clientID mit der
	// uebergebenen serial_number.
	RegistriereClient(ctx context.Context, tssID, clientID, serialNumber string) error
}
