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

// FiskalyTSESetupClient runs the guided TSE setup operations, both read-only checks and state transitions.
// It shares the HTTP machinery with the signing client but needs no TSS or client ID.
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
	// DSFinV-K export master data; on the TSS resource the serial number (SHA-256 of the public key, hex) is serial_number.
	// tss_serial_number exists only on transaction responses.
	TSSSerialNumber          string `json:"serial_number"`
	SignatureAlgorithm       string `json:"signature_algorithm"`
	PublicKey                string `json:"public_key"`
	Certificate              string `json:"certificate"`
	SignatureTimestampFormat string `json:"signature_timestamp_format"`
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

type clientStateRequest struct {
	State string `json:"state"`
}

// ListTSS lists the account's TSS and returns the environment (TEST/LIVE), which fiskaly sets even for an empty account.
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

// CreateTSS creates a TSS under a fresh UUID.
// fiskaly returns the one-time admin PUK needed to set the admin PIN.
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

// GetAdminPUK re-reads the admin PUK, which fiskaly returns only while the TSS is CREATED.
func (c *FiskalyTSESetupClient) GetAdminPUK(ctx context.Context, tssID string) (string, error) {
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

// RetrieveTSSStammdaten reads the TSS master data (signature algorithm, public key, certificate, log time format) for the DSFinV-K export.
func (c *FiskalyTSESetupClient) RetrieveTSSStammdaten(ctx context.Context, tssID string) (tse.Stammdaten, error) {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return tse.Stammdaten{}, fmt.Errorf("tss id is required")
	}
	resp := tssDetailResponse{}
	path := fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodGet, path, nil, nil, true, &resp); err != nil {
		return tse.Stammdaten{}, mapSetupError(err)
	}
	return tse.NewStammdaten(resp.TSSSerialNumber, resp.SignatureAlgorithm, resp.PublicKey, resp.Certificate, resp.SignatureTimestampFormat), nil
}

// PersonalisiereTSS moves the TSS from CREATED to UNINITIALIZED.
func (c *FiskalyTSESetupClient) PersonalisiereTSS(ctx context.Context, tssID string) error {
	return c.patchTSSState(ctx, tssID, "UNINITIALIZED")
}

// InitialisiereTSS moves the TSS to INITIALIZED, ready to sign.
// It requires a prior AuthentifiziereAdmin.
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

// SetAdminPIN sets the admin PIN using the admin PUK.
// The same endpoint (PATCH /tss/{id}/admin) resets a lost PIN or unblocks one after five failures, also on a personalised TSS.
func (c *FiskalyTSESetupClient) SetAdminPIN(ctx context.Context, tssID, puk, pin string) error {
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

// AuthentifiziereAdmin elevates the current access token to admin rights for the following TSS admin operations (initialise, register client).
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

// RegistriereClient registers a client under clientID with the Kassen-Seriennummer as serial_number.
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

// ReaktiviereClient reactivates an existing DEREGISTERED client via PATCH state=REGISTERED.
// It creates no new client because serial_number is unique per TSS.
func (c *FiskalyTSESetupClient) ReaktiviereClient(ctx context.Context, tssID, clientID string) error {
	tssID = strings.TrimSpace(tssID)
	clientID = strings.TrimSpace(clientID)
	if tssID == "" || clientID == "" {
		return fmt.Errorf("tss id and client id are required")
	}
	path := fmt.Sprintf("/api/v2/tss/%s/client/%s", url.PathEscape(tssID), url.PathEscape(clientID))
	body := clientStateRequest{State: "REGISTERED"}
	if err := c.doJSONRequest(ctx, http.MethodPatch, path, nil, body, true, nil); err != nil {
		return mapSetupError(err)
	}
	return nil
}

// mapSetupError maps the TSS limit codes and auth failures to domain sentinels; a blocked admin PIN counts as auth failure so the takeover offers the PUK reset.
// Codes are checked before the status because fiskaly sends the limit codes with 403 and E_ADMIN_PIN_BLOCKED with 423 (docs/rechtsquellen/fiskaly/fiskaly-SIGN-DE-API-v2-openapi.json).
func mapSetupError(err error) error {
	var apiErr apiError
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.Code {
	case "E_TSS_LIMIT_REACHED", "E_TSS_LIMIT_PER_DAY_REACHED":
		return tse.ErrSetupTSSLimitErreicht
	case "E_ADMIN_PIN_BLOCKED":
		return tse.ErrSetupAuthFehlgeschlagen
	}
	if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden {
		return tse.ErrSetupAuthFehlgeschlagen
	}
	return err
}
