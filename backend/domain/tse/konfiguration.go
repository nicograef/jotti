package tse

import (
	"fmt"
	"strings"
	"time"
)

type Konfiguration struct {
	ApiKey    string
	ApiSecret string
	TssID     string
	ClientID  string
	UpdatedAt time.Time
}

// Credentials bildet die TSE-Konfiguration auf die kanonische
// Credentials-Form ab — die einzige Stelle, an der die vier Felder
// gemappt werden.
func (t Konfiguration) Credentials() Credentials {
	return Credentials{
		ApiKey:    t.ApiKey,
		ApiSecret: t.ApiSecret,
		TssID:     t.TssID,
		ClientID:  t.ClientID,
	}
}

func (t Konfiguration) Validate() error {
	// Sonderfall: komplett leer ist gueltig (TSE schlicht nicht konfiguriert).
	// Sind Felder gesetzt, gilt die kanonische Vier-Felder-Regel aus
	// Credentials (alle oder keines).
	if !t.leer() {
		if err := t.Credentials().Validate(); err != nil {
			return err
		}
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

func (t Konfiguration) IstKonfiguriert() bool {
	return t.Credentials().Validate() == nil
}

func (t Konfiguration) leer() bool {
	return strings.TrimSpace(t.ApiKey) == "" &&
		strings.TrimSpace(t.ApiSecret) == "" &&
		strings.TrimSpace(t.TssID) == "" &&
		strings.TrimSpace(t.ClientID) == ""
}

func NewKonfiguration(apiKey, apiSecret, tssID, clientID string) (Konfiguration, error) {
	t := Konfiguration{
		ApiKey:    strings.TrimSpace(apiKey),
		ApiSecret: strings.TrimSpace(apiSecret),
		TssID:     strings.TrimSpace(tssID),
		ClientID:  strings.TrimSpace(clientID),
		UpdatedAt: time.Now().UTC(),
	}

	if err := t.Validate(); err != nil {
		return Konfiguration{}, err
	}

	return t, nil
}
