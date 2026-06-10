package tse_repo

import (
	"bytes"
	"context"
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

var ErrUpdateTransactionNichtUnterstuetzt = errors.New("update transaction is not supported in atomares muster")

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

type upsertTransactionRequest struct {
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

type errorResponse struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type sleepFn func(ctx context.Context, duration time.Duration) error

type FiskalyTSEClient struct {
	baseURL     string
	credentials tse.Credentials
	httpClient  *http.Client
	now         func() time.Time
	sleep       sleepFn
	maxRetries  int

	mu             sync.Mutex
	accessToken    string
	accessTokenEnv tse.Umgebung
	expiresAt      time.Time
}

var _ tse.TSEClient = (*FiskalyTSEClient)(nil)
var _ tse.ConnectionTester = (*FiskalyTSEClient)(nil)

func NewFiskalyTSEClient(baseURL string, credentials tse.Credentials, httpClient *http.Client) (*FiskalyTSEClient, error) {
	if err := credentials.Validate(); err != nil {
		return nil, err
	}

	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBaseURL == "" {
		return nil, fmt.Errorf("base url is required")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &FiskalyTSEClient{
		baseURL:     trimmedBaseURL,
		credentials: credentials,
		httpClient:  httpClient,
		now:         time.Now,
		sleep:       sleepWithContext,
		maxRetries:  defaultRetryAttempts,
	}, nil
}

func (c *FiskalyTSEClient) StartTransaction(ctx context.Context, txID string, processType string, processData string) (tse.StartResult, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return tse.StartResult{}, fmt.Errorf("tx id is required")
	}

	resp := transactionResponse{}
	err := c.doJSONRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.credentials.TssID), txID),
		url.Values{"tx_revision": []string{"1"}},
		upsertTransactionRequest{
			State:    "ACTIVE",
			ClientID: c.credentials.ClientID,
			Schema: rawSchemaEnvelope{
				Raw: rawSchema{
					ProcessType: processType,
					ProcessData: processData,
				},
			},
		},
		true,
		&resp,
	)
	if err != nil {
		return tse.StartResult{}, err
	}

	return mapStartResult(resp)
}

func (c *FiskalyTSEClient) UpdateTransaction(_ context.Context, _ string, _ int, _ string) error {
	return ErrUpdateTransactionNichtUnterstuetzt
}

func (c *FiskalyTSEClient) FinishTransaction(ctx context.Context, txID string, _ int, processType string, processData string) (tse.FinishResult, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return tse.FinishResult{}, fmt.Errorf("tx id is required")
	}

	resp := transactionResponse{}
	err := c.doJSONRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.credentials.TssID), url.PathEscape(txID)),
		url.Values{"tx_revision": []string{"2"}},
		upsertTransactionRequest{
			State:    "FINISHED",
			ClientID: c.credentials.ClientID,
			Schema: rawSchemaEnvelope{
				Raw: rawSchema{
					ProcessType: processType,
					ProcessData: processData,
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

func (c *FiskalyTSEClient) TestConnection(ctx context.Context) (tse.VerbindungStatus, error) {
	_, env, err := c.getAccessToken(ctx)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	resp := tssResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/api/v2/tss/%s", url.PathEscape(c.credentials.TssID)),
		nil,
		nil,
		true,
		&resp,
	)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	if env == "" {
		env = tse.Umgebung(strings.ToUpper(strings.TrimSpace(resp.Env)))
	}

	status := tse.VerbindungStatus{
		Umgebung: env,
		TSSState: strings.TrimSpace(resp.State),
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

func (c *FiskalyTSEClient) doJSONRequest(
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

func (c *FiskalyTSEClient) getAccessToken(ctx context.Context) (string, tse.Umgebung, error) {
	c.mu.Lock()
	if c.accessToken != "" && c.now().Add(tokenExpiryLeeway).Before(c.expiresAt) {
		token := c.accessToken
		env := c.accessTokenEnv
		c.mu.Unlock()
		return token, env, nil
	}
	c.mu.Unlock()

	request := map[string]string{
		"api_key":    c.credentials.ApiKey,
		"api_secret": c.credentials.ApiSecret,
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

func (c *FiskalyTSEClient) invalidateToken() {
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

func (c *FiskalyTSEClient) retryDelay(attempt int, retryAfterHeader string) time.Duration {
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
