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

// The HTTP transport shared by the fiskaly signing and setup clients: auth with token cache, retries, error mapping.

const (
	defaultHTTPTimeout   = 10 * time.Second
	defaultRetryAttempts = 3
	tokenExpiryLeeway    = 30 * time.Second
	maxRetryDelay        = 5 * time.Second
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

type errorResponse struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type sleepFn func(ctx context.Context, duration time.Duration) error

// fiskalyClient holds the HTTP machinery shared by the signing and setup clients: base URL, key/secret auth with token cache, retries.
// It needs no TSS or client ID.
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

	requestURL := c.baseURL + path
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	triedTokenRefresh := false
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		antwort, err := c.sendOnce(ctx, method, requestURL, payload, withAuth)
		var netz netzFehler
		if errors.As(err, &netz) {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt == c.maxRetries {
				return netz.err
			}
			if err := c.sleep(ctx, c.retryDelay(attempt, "")); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}

		if withAuth && antwort.status == http.StatusUnauthorized && !triedTokenRefresh {
			triedTokenRefresh = true
			c.invalidateToken()
			// A token refresh is no network retry and uses no attempt, which matters for single-attempt clients (maxRetries = 0).
			attempt--
			continue
		}

		if antwort.status >= 200 && antwort.status < 300 {
			if responseBody == nil {
				return nil
			}
			return json.Unmarshal(antwort.body, responseBody)
		}

		if isRetryableStatus(antwort.status) && attempt < c.maxRetries {
			if err := c.sleep(ctx, c.retryDelay(attempt, antwort.retryAfter)); err != nil {
				return err
			}
			continue
		}

		return mapAPIError(antwort.status, antwort.body)
	}

	return fmt.Errorf("request failed after retries")
}

// httpAntwort is what one exchange with fiskaly tells the retry loop.
type httpAntwort struct {
	status     int
	body       []byte
	retryAfter string
}

// netzFehler marks a failed HTTP round trip, the only error the retry loop resends on.
type netzFehler struct{ err error }

func (e netzFehler) Error() string { return e.err.Error() }
func (e netzFehler) Unwrap() error { return e.err }

// sendOnce performs one HTTP exchange and reads the whole body.
func (c *fiskalyClient) sendOnce(ctx context.Context, method, requestURL string, payload []byte, withAuth bool) (httpAntwort, error) {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(payload))
	if err != nil {
		return httpAntwort{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	if withAuth {
		token, _, err := c.getAccessToken(ctx)
		if err != nil {
			return httpAntwort{}, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return httpAntwort{}, netzFehler{err: err}
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		if closeErr != nil {
			return httpAntwort{}, fmt.Errorf("read response: %w", errors.Join(readErr, closeErr))
		}
		return httpAntwort{}, readErr
	}
	if closeErr != nil {
		return httpAntwort{}, closeErr
	}

	return httpAntwort{status: resp.StatusCode, body: body, retryAfter: resp.Header.Get("Retry-After")}, nil
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
	delay := parseRetryAfter(retryAfterHeader)
	if delay <= 0 {
		delay = 200 * time.Millisecond * time.Duration(1<<attempt)
	}
	return min(delay, maxRetryDelay)
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
