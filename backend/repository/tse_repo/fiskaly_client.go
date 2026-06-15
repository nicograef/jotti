package tse_repo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

const (
	defaultHTTPTimeout   = 10 * time.Second
	defaultRetryAttempts = 3
	tokenExpiryLeeway    = 30 * time.Second
)

type apiError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e apiError) Error() string {
	if e.Code != "" && e.Message != "" {
		return fmt.Sprintf("fiskaly api error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	if e.Code != "" {
		return fmt.Sprintf("fiskaly api error %d (%s)", e.StatusCode, e.Code)
	}
	if e.Message != "" {
		return fmt.Sprintf("fiskaly api error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("fiskaly api error %d", e.StatusCode)
}

type authResponse struct {
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresIn int64  `json:"access_token_expires_in"`
	AccessTokenExpiresAt int64  `json:"access_token_expires_at"`
	AccessTokenClaims    struct {
		Env string `json:"env"`
	} `json:"access_token_claims"`
}

type startTransactionRequest struct {
	State    string `json:"state"`
	ClientID string `json:"client_id"`
}

type finishTransactionRequest struct {
	State    string            `json:"state"`
	ClientID string            `json:"client_id"`
	Schema   rawSchemaEnvelope `json:"schema"`
}

type rawSchemaEnvelope struct {
	Raw rawSchema `json:"raw"`
}

type rawSchema struct {
	ProcessType string `json:"process_type"`
	ProcessData string `json:"process_data"`
}

type transactionResponse struct {
	Number          int             `json:"number"`
	State           string          `json:"state"`
	TSSSerialNumber string          `json:"tss_serial_number"`
	TimeStart       json.RawMessage `json:"time_start"`
	TimeEnd         json.RawMessage `json:"time_end"`
	QRCodeData      string          `json:"qr_code_data"`
	Log             struct {
		Timestamp json.RawMessage `json:"timestamp"`
	} `json:"log"`
	Signature struct {
		Value   string          `json:"value"`
		Counter json.RawMessage `json:"counter"`
	} `json:"signature"`
}

type tssResponse struct {
	State string `json:"state"`
	Env   string `json:"_env"`
}

type clientResponse struct {
	State        string `json:"state"`
	SerialNumber string `json:"serial_number"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type sleepFn func(ctx context.Context, duration time.Duration) error

// fiskalyClient buendelt die HTTP-Maschinerie, die sich der Signier-Client und
// der Setup-Client teilen: Basis-URL, API-Key/-Secret-Auth mit Token-Cache und
// die Retry-Logik. Sie kommt ohne TSS-/Client-ID aus.
type fiskalyClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	now        func() time.Time
	sleep      sleepFn
	maxRetries int

	mu             sync.Mutex
	accessToken    string
	accessTokenEnv tse.Umgebung
	expiresAt      time.Time
}

// FiskalyTSEClient signiert Transaktionen einer konkreten TSS/Client-Kombination.
type FiskalyTSEClient struct {
	*fiskalyClient
	tssID    string
	clientID string
}

var _ tse.TSEClient = (*FiskalyTSEClient)(nil)
var _ tse.ConnectionTester = (*FiskalyTSEClient)(nil)
var _ tse.TransactionRetriever = (*FiskalyTSEClient)(nil)

func newFiskalyClient(baseURL, apiKey, apiSecret string, httpClient *http.Client) (*fiskalyClient, error) {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBaseURL == "" {
		return nil, fmt.Errorf("base url is required")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &fiskalyClient{
		baseURL:    trimmedBaseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: httpClient,
		now:        time.Now,
		sleep:      sleepWithContext,
		maxRetries: defaultRetryAttempts,
	}, nil
}

func NewFiskalyTSEClient(baseURL string, credentials tse.Credentials, httpClient *http.Client) (*FiskalyTSEClient, error) {
	if err := credentials.Validate(); err != nil {
		return nil, err
	}

	base, err := newFiskalyClient(baseURL, credentials.ApiKey, credentials.ApiSecret, httpClient)
	if err != nil {
		return nil, err
	}

	return &FiskalyTSEClient{
		fiskalyClient: base,
		tssID:         credentials.TssID,
		clientID:      credentials.ClientID,
	}, nil
}

// NewFiskalyTSEClientSingleAttempt creates a client that tries each request
// exactly once (no retries). Used in the synchronous Kassier path, where the
// caller imposes a total deadline (tse.SignierDeadline) and failed signing
// falls back to the Nachsignier worker, which owns the full retry strategy.
func NewFiskalyTSEClientSingleAttempt(baseURL string, credentials tse.Credentials, httpClient *http.Client) (*FiskalyTSEClient, error) {
	client, err := NewFiskalyTSEClient(baseURL, credentials, httpClient)
	if err != nil {
		return nil, err
	}
	client.maxRetries = 0
	return client, nil
}

// StartTransaction sendet bewusst kein Schema: processType/processData muessen
// laut DSFinV-K bei StartTransaction immer leer sein (Anhang I).
func (c *FiskalyTSEClient) StartTransaction(ctx context.Context, txID string) (tse.StartResult, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return tse.StartResult{}, fmt.Errorf("tx id is required")
	}

	resp := transactionResponse{}
	err := c.doJSONRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.tssID), url.PathEscape(txID)),
		url.Values{"tx_revision": []string{"1"}},
		startTransactionRequest{
			State:    "ACTIVE",
			ClientID: c.clientID,
		},
		true,
		&resp,
	)
	if err != nil {
		return tse.StartResult{}, err
	}

	return mapStartResult(resp)
}

func (c *FiskalyTSEClient) FinishTransaction(ctx context.Context, txID string, processType string, processData string) (tse.FinishResult, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return tse.FinishResult{}, fmt.Errorf("tx id is required")
	}

	resp := transactionResponse{}
	err := c.doJSONRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.tssID), url.PathEscape(txID)),
		url.Values{"tx_revision": []string{"2"}},
		finishTransactionRequest{
			State:    "FINISHED",
			ClientID: c.clientID,
			Schema: rawSchemaEnvelope{
				Raw: rawSchema{
					ProcessType: processType,
					// Die fiskaly-API verlangt process_data als Base64; Aufrufer
					// liefern Klartext, das Encoding ist allein Sache dieses Clients.
					ProcessData: base64.StdEncoding.EncodeToString([]byte(processData)),
				},
			},
		},
		true,
		&resp,
	)
	if err != nil {
		return tse.FinishResult{}, err
	}

	return mapFinishResult(resp)
}

// RetrieveTransaction fragt den aktuellen Stand einer Transaktion ab (letzte
// Revision). Eine bei fiskaly unbekannte Transaktion wird als
// tse.ErrTransactionNichtGefunden gemeldet.
func (c *FiskalyTSEClient) RetrieveTransaction(ctx context.Context, txID string) (tse.RetrieveResult, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return tse.RetrieveResult{}, fmt.Errorf("tx id is required")
	}

	resp := transactionResponse{}
	err := c.doJSONRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.tssID), url.PathEscape(txID)),
		nil,
		nil,
		true,
		&resp,
	)
	if err != nil {
		var apiErr apiError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return tse.RetrieveResult{}, tse.ErrTransactionNichtGefunden
		}
		return tse.RetrieveResult{}, err
	}

	finishResult, err := mapFinishResult(resp)
	if err != nil {
		return tse.RetrieveResult{}, err
	}

	return tse.RetrieveResult{
		State:        tse.TransactionState(strings.ToUpper(strings.TrimSpace(resp.State))),
		FinishResult: finishResult,
	}, nil
}

// Umgebung liefert die Umgebung (TEST/LIVE) allein aus dem Auth-Token
// (access_token_claims.env). Sie kommt ohne TSS-/Client-Abruf aus und dient der
// reinen Statusanzeige, wo der volle Verbindungstest unnoetig waere.
func (c *fiskalyClient) Umgebung(ctx context.Context) (tse.Umgebung, error) {
	_, env, err := c.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return env, nil
}

func (c *FiskalyTSEClient) TestConnection(ctx context.Context) (tse.VerbindungStatus, error) {
	_, env, err := c.getAccessToken(ctx)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	tssResp := tssResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(c.tssID)),
		nil,
		nil,
		true,
		&tssResp,
	)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	clientResp := clientResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/api/v2/tss/%s/client/%s", url.PathEscape(c.tssID), url.PathEscape(c.clientID)),
		nil,
		nil,
		true,
		&clientResp,
	)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	if env == "" {
		env = tse.Umgebung(strings.ToUpper(strings.TrimSpace(tssResp.Env)))
	}

	// Ein nicht-REGISTERED-Client und ein Seriennummern-Mismatch sind keine
	// Transportfehler — sie werden als Befund im Status transportiert, damit die
	// UI das Ergebnis aufgeschluesselt anzeigen kann. Den Seriennummern-Abgleich
	// uebernimmt die Application-Schicht (sie kennt die Kassen-Seriennummer).
	status := tse.VerbindungStatus{
		Umgebung:           env,
		TSSState:           strings.TrimSpace(tssResp.State),
		ClientState:        strings.TrimSpace(clientResp.State),
		ClientSerialNumber: strings.TrimSpace(clientResp.SerialNumber),
	}
	if err := status.Validate(); err != nil {
		return tse.VerbindungStatus{}, err
	}
	return status, nil
}

func mapStartResult(resp transactionResponse) (tse.StartResult, error) {
	logTime, err := parseUnixTime(resp.Log.Timestamp)
	if err != nil {
		return tse.StartResult{}, err
	}
	if logTime.IsZero() {
		logTime, err = parseUnixTime(resp.TimeStart)
		if err != nil {
			return tse.StartResult{}, err
		}
	}

	signatureCounter, err := parseFlexibleInt(resp.Signature.Counter)
	if err != nil {
		return tse.StartResult{}, err
	}

	return tse.StartResult{
		TransactionNumber: resp.Number,
		LogTime:           logTime,
		SerialNumberTSE:   resp.TSSSerialNumber,
		SignatureCounter:  signatureCounter,
	}, nil
}

func mapFinishResult(resp transactionResponse) (tse.FinishResult, error) {
	logTime, err := parseUnixTime(resp.Log.Timestamp)
	if err != nil {
		return tse.FinishResult{}, err
	}
	logTimeStart, err := parseUnixTime(resp.TimeStart)
	if err != nil {
		return tse.FinishResult{}, err
	}
	logTimeEnd, err := parseUnixTime(resp.TimeEnd)
	if err != nil {
		return tse.FinishResult{}, err
	}
	if logTime.IsZero() {
		logTime = logTimeEnd
	}

	signatureCounter, err := parseFlexibleInt(resp.Signature.Counter)
	if err != nil {
		return tse.FinishResult{}, err
	}

	return tse.FinishResult{
		TransactionNumber: resp.Number,
		Signature:         resp.Signature.Value,
		LogTime:           logTime,
		LogTimeStart:      logTimeStart,
		LogTimeEnd:        logTimeEnd,
		SignatureCounter:  signatureCounter,
		SerialNumberTSE:   resp.TSSSerialNumber,
		QRCodeData:        resp.QRCodeData,
	}, nil
}

func (c *fiskalyClient) doJSONRequest(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	requestBody any,
	withAuth bool,
	responseBody any,
) error {
	var payload []byte
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		payload = encoded
	}

	triedTokenRefresh := false
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		requestURL := c.baseURL + path
		if len(query) > 0 {
			requestURL += "?" + query.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		if withAuth {
			token, _, err := c.getAccessToken(ctx)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt == c.maxRetries {
				return err
			}
			if err := c.sleep(ctx, c.retryDelay(attempt, "")); err != nil {
				return err
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			if closeErr != nil {
				return fmt.Errorf("read response: %w", errors.Join(readErr, closeErr))
			}
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}

		if withAuth && resp.StatusCode == http.StatusUnauthorized && !triedTokenRefresh {
			triedTokenRefresh = true
			c.invalidateToken()
			// Der Token-Refresh ist kein Netz-Retry und verbraucht keinen
			// Versuch — wichtig fuer Single-Attempt-Clients (maxRetries = 0).
			attempt--
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if responseBody == nil {
				return nil
			}
			if err := json.Unmarshal(body, responseBody); err != nil {
				return err
			}
			return nil
		}

		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			if err := c.sleep(ctx, c.retryDelay(attempt, resp.Header.Get("Retry-After"))); err != nil {
				return err
			}
			continue
		}

		return mapAPIError(resp.StatusCode, body)
	}

	return fmt.Errorf("request failed after retries")
}

func (c *fiskalyClient) getAccessToken(ctx context.Context) (string, tse.Umgebung, error) {
	c.mu.Lock()
	if c.accessToken != "" && c.now().Add(tokenExpiryLeeway).Before(c.expiresAt) {
		token := c.accessToken
		env := c.accessTokenEnv
		c.mu.Unlock()
		return token, env, nil
	}
	c.mu.Unlock()

	request := map[string]string{
		"api_key":    c.apiKey,
		"api_secret": c.apiSecret,
	}
	response := authResponse{}
	if err := c.doJSONRequest(ctx, http.MethodPost, "/api/v2/auth", nil, request, false, &response); err != nil {
		return "", "", err
	}

	expiresAt := time.Unix(response.AccessTokenExpiresAt, 0).UTC()
	if response.AccessTokenExpiresAt == 0 {
		expiresAt = c.now().Add(time.Duration(response.AccessTokenExpiresIn) * time.Second)
	}
	if expiresAt.Before(c.now()) {
		expiresAt = c.now().Add(5 * time.Minute)
	}
	env := tse.Umgebung(strings.ToUpper(strings.TrimSpace(response.AccessTokenClaims.Env)))

	c.mu.Lock()
	c.accessToken = response.AccessToken
	c.accessTokenEnv = env
	c.expiresAt = expiresAt
	c.mu.Unlock()

	return response.AccessToken, env, nil
}

func (c *fiskalyClient) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = ""
	c.accessTokenEnv = ""
	c.expiresAt = time.Time{}
}

func parseFlexibleInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}

	if raw[0] == '"' {
		var asString string
		if err := json.Unmarshal(raw, &asString); err != nil {
			return 0, err
		}
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, nil
		}
		parsed, err := strconv.Atoi(asString)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err != nil {
		return 0, err
	}
	parsed, err := asNumber.Int64()
	if err == nil {
		return int(parsed), nil
	}

	asFloat, err := asNumber.Float64()
	if err != nil {
		return 0, err
	}
	return int(asFloat), nil
}

func parseUnixTime(raw json.RawMessage) (time.Time, error) {
	value, err := parseFlexibleInt(raw)
	if err != nil {
		return time.Time{}, err
	}
	if value == 0 {
		return time.Time{}, nil
	}
	return time.Unix(int64(value), 0).UTC(), nil
}

func mapAPIError(statusCode int, body []byte) error {
	errResponse := errorResponse{}
	if err := json.Unmarshal(body, &errResponse); err == nil {
		code := strings.TrimSpace(errResponse.Code)
		if code == "" {
			code = strings.TrimSpace(errResponse.Error)
		}
		return apiError{StatusCode: statusCode, Code: code, Message: strings.TrimSpace(errResponse.Message)}
	}

	message := strings.TrimSpace(string(body))
	return apiError{StatusCode: statusCode, Message: message}
}

func isRetryableStatus(statusCode int) bool {
	return statusCode == 429 || statusCode == 499 || statusCode >= 500
}

func (c *fiskalyClient) retryDelay(attempt int, retryAfterHeader string) time.Duration {
	if parsed := parseRetryAfter(retryAfterHeader); parsed > 0 {
		return parsed
	}
	base := 200 * time.Millisecond
	delay := base * time.Duration(1<<attempt)
	if delay > 5*time.Second {
		return 5 * time.Second
	}
	return delay
}

func parseRetryAfter(value string) time.Duration {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(trimmed); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	if at, err := http.ParseTime(trimmed); err == nil {
		delta := time.Until(at)
		if delta > 0 {
			return delta
		}
	}

	return 0
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
