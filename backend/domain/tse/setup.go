package tse

import (
	"context"
	"errors"
	"strings"
)

// ErrSetupAuthFehlgeschlagen: Authentifizierung mit API-Key/-Secret
// fehlgeschlagen — fast immer falsche Zugangsdaten.
var ErrSetupAuthFehlgeschlagen = errors.New("tse setup authentication failed")

// ErrSetupTSSLimitErreicht: das fiskaly-Konto hat die Obergrenze aktiver TSS
// erreicht (in TEST fünf; fiskaly: E_TSS_LIMIT_REACHED). Alte TEST-TSS räumt
// fiskaly bei Inaktivität selbst ab.
var ErrSetupTSSLimitErreicht = errors.New("tse setup tss limit reached")

// SetupCredentials kommt ohne TSS-/Client-ID aus: beide entstehen erst im Verlauf
// der Einrichtung.
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

type TSSInfo struct {
	ID    string
	State string
}

type ClientInfo struct {
	ID           string
	SerialNumber string
	State        string
}

// TSSErstellt ist das Ergebnis der TSS-Neuanlage (Zustand CREATED, einmaliger
// Admin-PUK). Der PUK wird nie persistiert oder geloggt — er fließt nur bis in
// die einmalige Anzeige an den Admin.
type TSSErstellt struct {
	ID    string
	PUK   string
	State string
}

// SetupClient kapselt die fiskaly-Operationen der geführten TSE-Einrichtung.
// ListTSS liefert die Umgebung aus der fiskaly-Antwort mit, damit der Befund
// TEST/LIVE auch bei leerem Konto anzeigen kann.
type SetupClient interface {
	ListTSS(ctx context.Context) (Umgebung, []TSSInfo, error)
	ListClients(ctx context.Context, tssID string) ([]ClientInfo, error)

	RetrieveTSSStammdaten(ctx context.Context, tssID string) (Stammdaten, error)

	// CreateTSS legt eine neue TSS an (Zustand CREATED) mit einmaligem Admin-PUK.
	CreateTSS(ctx context.Context) (TSSErstellt, error)
	// GetAdminPUK liest den Admin-PUK erneut aus. fiskaly liefert ihn nur, solange
	// die TSS im Zustand CREATED ist (Admin-PIN noch nicht gesetzt) — das trägt die
	// Wiederaufnahme nach einem Abbruch ohne erneute Nutzereingabe.
	GetAdminPUK(ctx context.Context, tssID string) (string, error)
	// PersonalisiereTSS überführt die TSS von CREATED nach UNINITIALIZED.
	PersonalisiereTSS(ctx context.Context, tssID string) error
	// SetAdminPIN setzt mit dem PUK die Admin-PIN der TSS. Derselbe Endpunkt
	// setzt eine verlorene PIN neu bzw. entsperrt eine nach fünf Fehlversuchen
	// gesperrte PIN — auch auf einer bereits personalisierten TSS.
	SetAdminPIN(ctx context.Context, tssID, puk, pin string) error
	// AuthentifiziereAdmin hebt das aktuelle Zugriffstoken für die folgenden
	// Admin-Operationen der TSS auf Admin-Rechte an.
	AuthentifiziereAdmin(ctx context.Context, tssID, pin string) error
	// InitialisiereTSS überführt die TSS nach INITIALIZED (signierbereit).
	InitialisiereTSS(ctx context.Context, tssID string) error
	RegistriereClient(ctx context.Context, tssID, clientID, serialNumber string) error
	// ReaktiviereClient setzt einen DEREGISTERED Client auf state=REGISTERED. Die
	// serial_number ist je TSS eindeutig, ein zweiter Client mit derselben
	// Seriennummer entsteht darum nie. Setzt Admin-Authentifizierung voraus.
	ReaktiviereClient(ctx context.Context, tssID, clientID string) error
}
