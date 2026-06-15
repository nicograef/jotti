package tse_repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// FiskalyTSESetupClient fuehrt die lesenden Operationen der gefuehrten
// TSE-Einrichtung aus. Es teilt sich die HTTP-Maschinerie (Auth, Token-Cache,
// Retry) mit dem Signier-Client, kommt aber ohne TSS-/Client-ID aus.
type FiskalyTSESetupClient struct {
	*fiskalyClient
}

var _ tse.SetupClient = (*FiskalyTSESetupClient)(nil)

func NewFiskalyTSESetupClient(baseURL string, credentials tse.SetupCredentials, httpClient *http.Client) (*FiskalyTSESetupClient, error) {
	if err := credentials.Validate(); err != nil {
		return nil, err
	}

	base, err := newFiskalyClient(baseURL, credentials.ApiKey, credentials.ApiSecret, httpClient)
	if err != nil {
		return nil, err
	}

	return &FiskalyTSESetupClient{fiskalyClient: base}, nil
}

type tssListResponse struct {
	Data []tssListItem `json:"data"`
	Env  string        `json:"_env"`
}

type tssListItem struct {
	ID    string `json:"_id"`
	State string `json:"state"`
}

type clientListResponse struct {
	Data []clientListItem `json:"data"`
}

type clientListItem struct {
	ID           string `json:"_id"`
	SerialNumber string `json:"serial_number"`
	State        string `json:"state"`
}

type createTSSResponse struct {
	ID       string `json:"_id"`
	AdminPUK string `json:"admin_puk"`
	State    string `json:"state"`
}

type tssDetailResponse struct {
	AdminPUK string `json:"admin_puk"`
	State    string `json:"state"`
}

type tssStateRequest struct {
	State string `json:"state"`
}

type adminPINRequest struct {
	AdminPUK    string `json:"admin_puk"`
	NewAdminPIN string `json:"new_admin_pin"`
}

type adminAuthRequest struct {
	AdminPIN string `json:"admin_pin"`
}

type registerClientRequest struct {
	SerialNumber string `json:"serial_number"`
}

// ListTSS listet die TSS des Kontos und liefert die Umgebung (TEST/LIVE) aus der
// fiskaly-Antwort mit — letztere ist auch bei leerem Konto gesetzt.
func (c *FiskalyTSESetupClient) ListTSS(ctx context.Context) (tse.Umgebung, []tse.TSSInfo, error) {
	resp := tssListResponse{}
	if err := c.doJSONRequest(ctx, http.MethodGet, "/api/v2/tss", nil, nil, true, &resp); err != nil {
		return "", nil, mapSetupError(err)
	}

	env := tse.Umgebung(strings.ToUpper(strings.TrimSpace(resp.Env)))
	tssList := make([]tse.TSSInfo, 0, len(resp.Data))
	for _, item := range resp.Data {
		tssList = append(tssList, tse.TSSInfo{
			ID:    strings.TrimSpace(item.ID),
			State: strings.TrimSpace(item.State),
		})
	}
	return env, tssList, nil
}

func (c *FiskalyTSESetupClient) ListClients(ctx context.Context, tssID string) ([]tse.ClientInfo, error) {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return nil, fmt.Errorf("tss id is required")
	}

	resp := clientListResponse{}
	path := fmt.Sprintf("/api/v2/tss/%s/client", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodGet, path, nil, nil, true, &resp); err != nil {
		return nil, mapSetupError(err)
	}

	clients := make([]tse.ClientInfo, 0, len(resp.Data))
	for _, item := range resp.Data {
		clients = append(clients, tse.ClientInfo{
			ID:           strings.TrimSpace(item.ID),
			SerialNumber: strings.TrimSpace(item.SerialNumber),
			State:        strings.TrimSpace(item.State),
		})
	}
	return clients, nil
}

// CreateTSS legt unter einer frisch erzeugten UUID eine neue TSS an. fiskaly
// liefert in der Antwort den einmaligen Admin-PUK, mit dem spaeter die Admin-PIN
// gesetzt wird.
func (c *FiskalyTSESetupClient) CreateTSS(ctx context.Context) (tse.TSSErstellt, error) {
	tssID := uuid.NewString()

	resp := createTSSResponse{}
	path := fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodPut, path, nil, struct{}{}, true, &resp); err != nil {
		return tse.TSSErstellt{}, mapSetupError(err)
	}

	id := strings.TrimSpace(resp.ID)
	if id == "" {
		id = tssID
	}
	return tse.TSSErstellt{
		ID:    id,
		PUK:   strings.TrimSpace(resp.AdminPUK),
		State: strings.TrimSpace(resp.State),
	}, nil
}

// HoleAdminPUK liest den Admin-PUK einer TSS erneut aus der TSS-Ressource.
// fiskaly liefert ihn dort nur, solange die TSS im Zustand CREATED ist.
func (c *FiskalyTSESetupClient) HoleAdminPUK(ctx context.Context, tssID string) (string, error) {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return "", fmt.Errorf("tss id is required")
	}
	resp := tssDetailResponse{}
	path := fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodGet, path, nil, nil, true, &resp); err != nil {
		return "", mapSetupError(err)
	}
	return strings.TrimSpace(resp.AdminPUK), nil
}

// PersonalisiereTSS ueberfuehrt die TSS von CREATED nach UNINITIALIZED.
func (c *FiskalyTSESetupClient) PersonalisiereTSS(ctx context.Context, tssID string) error {
	return c.patchTSSState(ctx, tssID, "UNINITIALIZED")
}

// InitialisiereTSS ueberfuehrt die TSS nach INITIALIZED und macht sie damit
// signierbereit. Sie setzt eine vorher erfolgte Admin-Authentifizierung voraus.
func (c *FiskalyTSESetupClient) InitialisiereTSS(ctx context.Context, tssID string) error {
	return c.patchTSSState(ctx, tssID, "INITIALIZED")
}

func (c *FiskalyTSESetupClient) patchTSSState(ctx context.Context, tssID, state string) error {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return fmt.Errorf("tss id is required")
	}
	path := fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodPatch, path, nil, tssStateRequest{State: state}, true, nil); err != nil {
		return mapSetupError(err)
	}
	return nil
}

// SetzeAdminPIN setzt mit dem Admin-PUK die Admin-PIN der TSS.
func (c *FiskalyTSESetupClient) SetzeAdminPIN(ctx context.Context, tssID, puk, pin string) error {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return fmt.Errorf("tss id is required")
	}
	path := fmt.Sprintf("/api/v2/tss/%s/admin", url.PathEscape(tssID))
	body := adminPINRequest{AdminPUK: puk, NewAdminPIN: pin}
	if err := c.doJSONRequest(ctx, http.MethodPatch, path, nil, body, true, nil); err != nil {
		return mapSetupError(err)
	}
	return nil
}

// AuthentifiziereAdmin hebt das aktuelle Zugriffstoken fuer die folgenden
// Admin-Operationen der TSS (Initialisieren, Client registrieren) auf
// Admin-Rechte an.
func (c *FiskalyTSESetupClient) AuthentifiziereAdmin(ctx context.Context, tssID, pin string) error {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return fmt.Errorf("tss id is required")
	}
	path := fmt.Sprintf("/api/v2/tss/%s/admin/auth", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodPost, path, nil, adminAuthRequest{AdminPIN: pin}, true, nil); err != nil {
		return mapSetupError(err)
	}
	return nil
}

// RegistriereClient registriert einen Client unter clientID mit der uebergebenen
// serial_number (der jotti-Kassen-Seriennummer).
func (c *FiskalyTSESetupClient) RegistriereClient(ctx context.Context, tssID, clientID, serialNumber string) error {
	tssID = strings.TrimSpace(tssID)
	clientID = strings.TrimSpace(clientID)
	if tssID == "" || clientID == "" {
		return fmt.Errorf("tss id and client id are required")
	}
	path := fmt.Sprintf("/api/v2/tss/%s/client/%s", url.PathEscape(tssID), url.PathEscape(clientID))
	body := registerClientRequest{SerialNumber: serialNumber}
	if err := c.doJSONRequest(ctx, http.MethodPut, path, nil, body, true, nil); err != nil {
		return mapSetupError(err)
	}
	return nil
}

// mapSetupError uebersetzt einen Auth-Fehler (falsche Zugangsdaten) in das
// Domain-Sentinel, damit die Application-Schicht eine verstaendliche Meldung
// erzeugen kann. Alle anderen Fehler bleiben unveraendert.
func mapSetupError(err error) error {
	var apiErr apiError
	if errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden) {
		return tse.ErrSetupAuthFehlgeschlagen
	}
	return err
}
