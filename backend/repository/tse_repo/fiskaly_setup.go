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
	// Fiskalische Stammdaten der TSS-Ressource fuer den DSFinV-K-Export. fiskaly
	// nennt das Log-Time-Format signature_timestamp_format und die Seriennummer
	// tss_serial_number (SHA-256 des Public Key, hex-kodiert).
	TSSSerialNumber          string `json:"tss_serial_number"`
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

// GetAdminPUK liest den Admin-PUK einer TSS erneut aus der TSS-Ressource.
// fiskaly liefert ihn dort nur, solange die TSS im Zustand CREATED ist.
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

// RetrieveTSSStammdaten liest die fiskalischen Stammdaten der TSS-Ressource
// (Signaturalgorithmus, Public Key, Zertifikat, Log-Time-Format) fuer den
// DSFinV-K-Export. Reine Leseoperation auf derselben TSS-Ressource wie
// GetAdminPUK.
func (c *FiskalyTSESetupClient) RetrieveTSSStammdaten(ctx context.Context, tssID string) (tse.TSSStammdaten, error) {
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return tse.TSSStammdaten{}, fmt.Errorf("tss id is required")
	}
	resp := tssDetailResponse{}
	path := fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(tssID))
	if err := c.doJSONRequest(ctx, http.MethodGet, path, nil, nil, true, &resp); err != nil {
		return tse.TSSStammdaten{}, mapSetupError(err)
	}
	return tse.TSSStammdaten{
		Seriennummer:        strings.TrimSpace(resp.TSSSerialNumber),
		SignaturAlgorithmus: strings.TrimSpace(resp.SignatureAlgorithm),
		PublicKey:           strings.TrimSpace(resp.PublicKey),
		Zertifikat:          strings.TrimSpace(resp.Certificate),
		LogTimeFormat:       strings.TrimSpace(resp.SignatureTimestampFormat),
	}, nil
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

// SetAdminPIN setzt mit dem Admin-PUK die Admin-PIN der TSS. Derselbe Endpunkt
// (PATCH /tss/{id}/admin) setzt eine verlorene PIN neu bzw. entsperrt eine nach
// fuenf Fehlversuchen gesperrte PIN und funktioniert auch auf einer bereits
// personalisierten TSS (UNINITIALIZED/INITIALIZED).
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

// ReaktiviereClient reaktiviert einen bereits vorhandenen, aber DEREGISTERED
// Client per PATCH mit state=REGISTERED. Es wird kein neuer Client angelegt — die
// serial_number ist je TSS eindeutig.
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

// mapSetupError uebersetzt bekannte fiskaly-Fehler in Domain-Sentinels, damit
// die Application-Schicht verstaendliche Meldungen erzeugen kann: einen
// Auth-Fehler (falsche Zugangsdaten oder abgelehnte/gesperrte Admin-PIN) und das
// Erreichen des TSS-Limits (E_TSS_LIMIT_REACHED, in TEST fuenf aktive TSS). Eine
// nach fuenf Fehlversuchen gesperrte Admin-PIN (E_ADMIN_PIN_BLOCKED) liefert
// fiskaly mit Status 423; sie wird hier ebenfalls als Auth-Fehler gemeldet, damit
// die Uebernahme in die PIN-Sackgasse (mit PUK-Reset als Ausweg) statt in einen
// technischen Fehler laeuft. Alle anderen Fehler bleiben unveraendert.
func mapSetupError(err error) error {
	var apiErr apiError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden {
			return tse.ErrSetupAuthFehlgeschlagen
		}
		if apiErr.Code == "E_ADMIN_PIN_BLOCKED" {
			return tse.ErrSetupAuthFehlgeschlagen
		}
		if apiErr.Code == "E_TSS_LIMIT_REACHED" {
			return tse.ErrSetupTSSLimitErreicht
		}
	}
	return err
}
