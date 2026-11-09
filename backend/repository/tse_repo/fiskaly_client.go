package tse_repo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

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

// FiskalyTSEClient signs transactions for one TSS/client pair.
type FiskalyTSEClient struct {
	*fiskalyClient
	tssID    string
	clientID string
}

var _ tse.TSEClient = (*FiskalyTSEClient)(nil)
var _ tse.ConnectionTester = (*FiskalyTSEClient)(nil)
var _ tse.TransactionRetriever = (*FiskalyTSEClient)(nil)

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

// tssZustandsCodes400 are fiskaly error codes that report a TSS-wide state despite HTTP 400 (docs/rechtsquellen/fiskaly/fiskaly-SIGN-DE-API-v2-openapi.json).
// They affect every signing, not a single order, so they stay TSE-wide.
var tssZustandsCodes400 = map[string]bool{
	"E_TSS_NOT_INITIALIZED": true,
	"E_TSS_DISABLED":        true,
	"E_CLIENT_DEREGISTERED": true,
	"E_CLIENT_NOT_FOUND":    true,
	"E_TX_LIMIT_REACHED":    true,
}

// klassifiziereSignierFehler marks HTTP 400/409/422 rejections of the concrete transaction as tse.AuftragsFehler, except tssZustandsCodes400.
// All other errors (connection, 401/403, 404, 423, 429, 5xx) stay unmarked and count as TSE-wide for the worker.
func klassifiziereSignierFehler(err error) error {
	var apiErr apiError
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.StatusCode {
	case http.StatusBadRequest:
		if tssZustandsCodes400[apiErr.Code] {
			return err
		}
		return tse.AuftragsFehler{Err: err}
	case http.StatusConflict, http.StatusUnprocessableEntity:
		return tse.AuftragsFehler{Err: err}
	default:
		return err
	}
}

// txPfad is the escaped path of one transaction of this client's TSS.
func (c *FiskalyTSEClient) txPfad(txID string) (string, error) {
	txID = strings.TrimSpace(txID)
	if txID == "" {
		return "", fmt.Errorf("tx id is required")
	}
	return fmt.Sprintf("/api/v2/tss/%s/tx/%s", url.PathEscape(c.tssID), url.PathEscape(txID)), nil
}

// StartTransaction sends no schema: processType and processData must be empty at start (DSFinV-K Anhang I, docs/compliance.md §3.2).
func (c *FiskalyTSEClient) StartTransaction(ctx context.Context, txID string) (tse.StartResult, error) {
	pfad, err := c.txPfad(txID)
	if err != nil {
		return tse.StartResult{}, err
	}

	resp := transactionResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodPut,
		pfad,
		url.Values{"tx_revision": []string{"1"}},
		startTransactionRequest{
			State:    "ACTIVE",
			ClientID: c.clientID,
		},
		mitAuth,
		&resp,
	)
	if err != nil {
		return tse.StartResult{}, klassifiziereSignierFehler(err)
	}

	return mapStartResult(resp)
}

func (c *FiskalyTSEClient) FinishTransaction(ctx context.Context, txID string, processType string, processData string) (tse.FinishResult, error) {
	pfad, err := c.txPfad(txID)
	if err != nil {
		return tse.FinishResult{}, err
	}

	resp := transactionResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodPut,
		pfad,
		url.Values{"tx_revision": []string{"2"}},
		finishTransactionRequest{
			State:    "FINISHED",
			ClientID: c.clientID,
			Schema: rawSchemaEnvelope{
				Raw: rawSchema{
					ProcessType: processType,
					// fiskaly expects process_data as Base64; callers pass plain text.
					ProcessData: base64.StdEncoding.EncodeToString([]byte(processData)),
				},
			},
		},
		mitAuth,
		&resp,
	)
	if err != nil {
		return tse.FinishResult{}, klassifiziereSignierFehler(err)
	}

	return mapFinishResult(resp)
}

// RetrieveTransaction fetches the latest revision of a transaction; an unknown one yields tse.ErrTransactionNichtGefunden.
func (c *FiskalyTSEClient) RetrieveTransaction(ctx context.Context, txID string) (tse.RetrieveResult, error) {
	pfad, err := c.txPfad(txID)
	if err != nil {
		return tse.RetrieveResult{}, err
	}

	resp := transactionResponse{}
	err = c.doJSONRequest(
		ctx,
		http.MethodGet,
		pfad,
		nil,
		nil,
		mitAuth,
		&resp,
	)
	if err != nil {
		// Only E_TX_NOT_FOUND means the transaction does not exist (yet); E_TSS_NOT_FOUND (wrong TSS ID) stays TSE-wide.
		// A 404 without error code counts as an unknown transaction.
		var apiErr apiError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound && apiErr.Code != "E_TSS_NOT_FOUND" {
			return tse.RetrieveResult{}, tse.ErrTransactionNichtGefunden
		}
		return tse.RetrieveResult{}, err
	}

	finishResult, err := mapFinishResult(resp)
	if err != nil {
		return tse.RetrieveResult{}, err
	}

	return tse.RetrieveResult{
		State:        tse.TransactionState(fiskalyZustand(resp.State)),
		FinishResult: finishResult,
	}, nil
}

// Umgebung reads the environment (TEST/LIVE) from the auth token alone (access_token_claims.env).
// It serves the status display, where a full connection test is unnecessary.
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
		mitAuth,
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
		mitAuth,
		&clientResp,
	)
	if err != nil {
		return tse.VerbindungStatus{}, err
	}

	if env == "" {
		env = tse.Umgebung(strings.ToUpper(strings.TrimSpace(tssResp.Env)))
	}

	// A non-REGISTERED client or a serial number mismatch is a finding in the status, not a transport error, so the UI can itemise it.
	// The application layer compares serial numbers because it knows the Kassen-Seriennummer.
	status := tse.VerbindungStatus{
		Umgebung:           env,
		TSSState:           tse.TSSZustand(fiskalyZustand(tssResp.State)),
		ClientState:        tse.ClientZustand(fiskalyZustand(clientResp.State)),
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

// fiskalyZustand normalizes a fiskaly state once at the boundary, so every comparison behind it can match exact upper-case values.
func fiskalyZustand(state string) string {
	return strings.ToUpper(strings.TrimSpace(state))
}
