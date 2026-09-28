package tse

import (
	"context"
	"errors"
	"strings"
)

// ErrSetupAuthFehlgeschlagen almost always means a wrong API key or secret.
var ErrSetupAuthFehlgeschlagen = errors.New("tse setup authentication failed")

// ErrSetupTSSLimitErreicht is fiskaly's E_TSS_LIMIT_REACHED (five active TSS in TEST).
// fiskaly removes inactive TEST TSS by itself.
var ErrSetupTSSLimitErreicht = errors.New("tse setup tss limit reached")

// SetupCredentials has no TSS or client ID because the setup creates both.
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

// TSSErstellt carries the one-time admin PUK of a CREATED TSS.
// The PUK is never persisted or logged; it only reaches the one-time admin display.
type TSSErstellt struct {
	ID    string
	PUK   string
	State string
}

// SetupClient holds the fiskaly operations of the guided TSE setup.
// ListTSS returns the Umgebung so the report shows TEST/LIVE even for an empty account.
type SetupClient interface {
	ListTSS(ctx context.Context) (Umgebung, []TSSInfo, error)
	ListClients(ctx context.Context, tssID string) ([]ClientInfo, error)

	RetrieveTSSStammdaten(ctx context.Context, tssID string) (Stammdaten, error)

	// CreateTSS returns a CREATED TSS with its one-time admin PUK.
	CreateTSS(ctx context.Context) (TSSErstellt, error)
	// GetAdminPUK works only while the TSS is CREATED (admin PIN not yet set).
	// This lets an aborted setup resume without new user input.
	GetAdminPUK(ctx context.Context, tssID string) (string, error)
	// PersonalisiereTSS moves the TSS from CREATED to UNINITIALIZED.
	PersonalisiereTSS(ctx context.Context, tssID string) error
	// SetAdminPIN also resets a lost PIN or unblocks one locked after five failed attempts.
	// This works on an already personalised TSS too.
	SetAdminPIN(ctx context.Context, tssID, puk, pin string) error
	// AuthentifiziereAdmin elevates the current access token for the following admin operations.
	AuthentifiziereAdmin(ctx context.Context, tssID, pin string) error
	// InitialisiereTSS moves the TSS to INITIALIZED, ready to sign.
	InitialisiereTSS(ctx context.Context, tssID string) error
	RegistriereClient(ctx context.Context, tssID, clientID, serialNumber string) error
	// ReaktiviereClient re-registers a DEREGISTERED client, since serial_number is unique per TSS.
	// It requires admin authentication.
	ReaktiviereClient(ctx context.Context, tssID, clientID string) error
}
