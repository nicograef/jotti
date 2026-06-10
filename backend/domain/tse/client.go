package tse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Umgebung string

const (
	UmgebungTest Umgebung = "TEST"
	UmgebungLive Umgebung = "LIVE"
)

var ErrUnvollstaendigeCredentials = errors.New("tse credentials are incomplete")

type Credentials struct {
	ApiKey    string
	ApiSecret string
	TssID     string
	ClientID  string
}

func (c Credentials) Validate() error {
	hasApiKey := strings.TrimSpace(c.ApiKey) != ""
	hasApiSecret := strings.TrimSpace(c.ApiSecret) != ""
	hasTssID := strings.TrimSpace(c.TssID) != ""
	hasClientID := strings.TrimSpace(c.ClientID) != ""

	hasAny := hasApiKey || hasApiSecret || hasTssID || hasClientID
	hasAll := hasApiKey && hasApiSecret && hasTssID && hasClientID
	if hasAny && !hasAll {
		return ErrUnvollstaendigeCredentials
	}
	if !hasAll {
		return ErrUnvollstaendigeCredentials
	}
	return nil
}

type TSEClient interface {
	StartTransaction(ctx context.Context, kassenID string, processType string, processData string) (StartResult, error)
	UpdateTransaction(ctx context.Context, kassenID string, transactionNumber int, processData string) error
	FinishTransaction(ctx context.Context, kassenID string, transactionNumber int, processType string, processData string) (FinishResult, error)
}

type ConnectionTester interface {
	TestConnection(ctx context.Context) (VerbindungStatus, error)
}

type StartResult struct {
	TransactionNumber int
	LogTime           time.Time
	SerialNumberTSE   string
	SignatureCounter  int
}

type FinishResult struct {
	TransactionNumber int
	Signature         string
	LogTime           time.Time
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	SignatureCounter  int
	SerialNumberTSE   string
	QRCodeData        string
}

type VerbindungStatus struct {
	Umgebung Umgebung
	TSSState string
}

func (v VerbindungStatus) Validate() error {
	if strings.TrimSpace(string(v.Umgebung)) == "" {
		return fmt.Errorf("umgebung is required")
	}
	if strings.TrimSpace(v.TSSState) == "" {
		return fmt.Errorf("tss state is required")
	}
	return nil
}
