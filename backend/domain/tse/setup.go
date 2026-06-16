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

// ErrSetupTSSLimitErreicht zeigt an, dass das fiskaly-Konto die Obergrenze
// aktiver TSS erreicht hat (in TEST fuenf; fiskaly: E_TSS_LIMIT_REACHED) und
// keine weitere TSS angelegt werden kann. Alte TEST-TSS werden von fiskaly bei
// Inaktivitaet automatisch bereinigt.
var ErrSetupTSSLimitErreicht = errors.New("tse setup tss limit reached")

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

// TSSStammdaten sind die fiskalischen Stammdaten der TSS-Ressource, die der
// DSFinV-K-Export braucht: Signaturalgorithmus, Public Key, Zertifikat,
// Log-Time-Format (fiskaly: signature_timestamp_format) und die API-Version. Sie
// aendern sich ueber die Lebensdauer der TSS nicht.
type TSSStammdaten struct {
	SignaturAlgorithmus string
	PublicKey           string
	Zertifikat          string
	LogTimeFormat       string
	Version             string
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

	// RetrieveTSSStammdaten liest die fiskalischen Stammdaten der TSS-Ressource
	// (Signaturalgorithmus, Public Key, Zertifikat, Log-Time-Format, Version) fuer
	// den DSFinV-K-Export. Reine Leseoperation.
	RetrieveTSSStammdaten(ctx context.Context, tssID string) (TSSStammdaten, error)

	// CreateTSS legt eine neue TSS an (Zustand CREATED) und liefert deren
	// einmaligen Admin-PUK zurueck.
	CreateTSS(ctx context.Context) (TSSErstellt, error)
	// HoleAdminPUK liest den Admin-PUK einer TSS erneut aus. fiskaly liefert ihn
	// nur, solange die TSS im Zustand CREATED ist (Admin-PIN noch nicht gesetzt);
	// danach ist er nicht mehr abrufbar. Dient der Wiederaufnahme nach einem
	// Abbruch im Zustand CREATED ohne erneute Nutzereingabe.
	HoleAdminPUK(ctx context.Context, tssID string) (string, error)
	// PersonalisiereTSS ueberfuehrt die TSS von CREATED nach UNINITIALIZED.
	PersonalisiereTSS(ctx context.Context, tssID string) error
	// SetzeAdminPIN setzt mit dem PUK die Admin-PIN der TSS. Derselbe Endpunkt
	// setzt eine verlorene PIN neu bzw. entsperrt eine nach fuenf Fehlversuchen
	// gesperrte PIN — auch auf einer bereits personalisierten TSS.
	SetzeAdminPIN(ctx context.Context, tssID, puk, pin string) error
	// AuthentifiziereAdmin hebt das aktuelle Zugriffstoken fuer die folgenden
	// Admin-Operationen der TSS auf Admin-Rechte an.
	AuthentifiziereAdmin(ctx context.Context, tssID, pin string) error
	// InitialisiereTSS ueberfuehrt die TSS nach INITIALIZED (signierbereit).
	InitialisiereTSS(ctx context.Context, tssID string) error
	// RegistriereClient registriert einen Client unter clientID mit der
	// uebergebenen serial_number.
	RegistriereClient(ctx context.Context, tssID, clientID, serialNumber string) error
	// ReaktiviereClient reaktiviert einen vorhandenen, aber DEREGISTERED Client
	// per state=REGISTERED. Die serial_number ist je TSS eindeutig, daher wird
	// kein neuer Client mit derselben Seriennummer angelegt — derselbe clientID
	// wird wieder aktiviert. Setzt eine vorherige Admin-Authentifizierung voraus.
	ReaktiviereClient(ctx context.Context, tssID, clientID string) error
}
