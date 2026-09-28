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

// ErrTransactionNichtGefunden tells the signing worker to start the transaction fresh.
var ErrTransactionNichtGefunden = errors.New("tse transaction not found")

// AuftragsFehler marks a job-specific signing error, such as processData rejected by fiskaly.
// Any unmarked error counts as TSE-wide; see docs/handbuch.md §3.13 (Signatur-Worker).
type AuftragsFehler struct {
	Err error
}

func (e AuftragsFehler) Error() string { return e.Err.Error() }

func (e AuftragsFehler) Unwrap() error { return e.Err }

func IstAuftragsFehler(err error) bool {
	var auftragsFehler AuftragsFehler
	return errors.As(err, &auftragsFehler)
}

type Credentials struct {
	ApiKey    string
	ApiSecret string
	TssID     string
	ClientID  string
}

func (c Credentials) Validate() error {
	if strings.TrimSpace(c.ApiKey) == "" ||
		strings.TrimSpace(c.ApiSecret) == "" ||
		strings.TrimSpace(c.TssID) == "" ||
		strings.TrimSpace(c.ClientID) == "" {
		return ErrUnvollstaendigeCredentials
	}
	return nil
}

// TSEClient implements the atomic pattern: Start carries no process data (DSFinV-K Anhang I), Finish the final schema.
// See docs/compliance.md §3.2.
type TSEClient interface {
	StartTransaction(ctx context.Context, txID string) (StartResult, error)
	FinishTransaction(ctx context.Context, txID string, processType string, processData string) (FinishResult, error)
}

// ConnectionTester: TestConnection is the full diagnosis, Umgebung the light path from the auth token alone.
type ConnectionTester interface {
	TestConnection(ctx context.Context) (VerbindungStatus, error)
	Umgebung(ctx context.Context) (Umgebung, error)
}

// TransactionRetriever returns ErrTransactionNichtGefunden for an unknown transaction.
type TransactionRetriever interface {
	RetrieveTransaction(ctx context.Context, txID string) (RetrieveResult, error)
}

type TransactionState string

const (
	TransactionStateActive    TransactionState = "ACTIVE"
	TransactionStateFinished  TransactionState = "FINISHED"
	TransactionStateCancelled TransactionState = "CANCELLED"
)

// RetrieveResult carries signature data only for finished transactions.
type RetrieveResult struct {
	State TransactionState
	FinishResult
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

// VerbindungStatus: the repository fills the fiskaly fields; the application layer sets SeriennummerKorrekt.
// It compares the jotti Kassen-Seriennummer with the client serial_number.
type VerbindungStatus struct {
	Umgebung            Umgebung
	TSSState            string
	ClientState         string
	ClientSerialNumber  string
	SeriennummerKorrekt bool
}

func (v VerbindungStatus) Validate() error {
	if strings.TrimSpace(string(v.Umgebung)) == "" {
		return fmt.Errorf("umgebung is required")
	}
	if strings.TrimSpace(v.TSSState) == "" {
		return fmt.Errorf("tss state is required")
	}
	if strings.TrimSpace(v.ClientState) == "" {
		return fmt.Errorf("client state is required")
	}
	if strings.TrimSpace(v.ClientSerialNumber) == "" {
		return fmt.Errorf("client serial number is required")
	}
	return nil
}
