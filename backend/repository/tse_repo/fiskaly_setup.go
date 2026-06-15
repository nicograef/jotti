package tse_repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
