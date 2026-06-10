package settings

import (
	"fmt"
	"strings"
	"time"
)

type TSEKonfiguration struct {
	ApiKey    string
	ApiSecret string
	TssID     string
	ClientID  string
	UpdatedAt time.Time
}

func (t TSEKonfiguration) Validate() error {
	hasApiKey := strings.TrimSpace(t.ApiKey) != ""
	hasApiSecret := strings.TrimSpace(t.ApiSecret) != ""
	hasTssID := strings.TrimSpace(t.TssID) != ""
	hasClientID := strings.TrimSpace(t.ClientID) != ""

	hasAny := hasApiKey || hasApiSecret || hasTssID || hasClientID
	hasAll := hasApiKey && hasApiSecret && hasTssID && hasClientID
	if hasAny && !hasAll {
		return fmt.Errorf("all tse fields must be set together")
	}

	if len(t.ApiKey) > 500 {
		return fmt.Errorf("api_key is too long")
	}
	if len(t.ApiSecret) > 500 {
		return fmt.Errorf("api_secret is too long")
	}
	if len(t.TssID) > 255 {
		return fmt.Errorf("tss_id is too long")
	}
	if len(t.ClientID) > 255 {
		return fmt.Errorf("client_id is too long")
	}
	if t.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	return nil
}

func (t TSEKonfiguration) IstKonfiguriert() bool {
	return strings.TrimSpace(t.ApiKey) != "" &&
		strings.TrimSpace(t.ApiSecret) != "" &&
		strings.TrimSpace(t.TssID) != "" &&
		strings.TrimSpace(t.ClientID) != ""
}

func NewTSEKonfiguration(apiKey, apiSecret, tssID, clientID string) (TSEKonfiguration, error) {
	t := TSEKonfiguration{
		ApiKey:    strings.TrimSpace(apiKey),
		ApiSecret: strings.TrimSpace(apiSecret),
		TssID:     strings.TrimSpace(tssID),
		ClientID:  strings.TrimSpace(clientID),
		UpdatedAt: time.Now(),
	}

	if err := t.Validate(); err != nil {
		return TSEKonfiguration{}, err
	}

	return t, nil
}
