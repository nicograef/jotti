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

// ErrTransactionNichtGefunden zeigt an, dass eine Transaktion bei der TSE
// (noch) nicht existiert — der Nachsignier-Worker startet sie dann neu.
var ErrTransactionNichtGefunden = errors.New("tse transaction not found")

// SignierDeadline ist die Gesamt-Deadline fuer den synchronen Signierversuch
// im Kassier-Pfad (Auth + Start + Finish, max. 1 Versuch pro Request).
// Laeuft sie ab, greift der Ausfallpfad (Nachsignier-Auftrag); die volle
// Retry-Strategie lebt nur im Nachsignier-Worker.
const SignierDeadline = 5 * time.Second

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

	hasAll := hasApiKey && hasApiSecret && hasTssID && hasClientID
	if !hasAll {
		return ErrUnvollstaendigeCredentials
	}
	return nil
}

// TSEClient bildet das atomare Transaktionsmuster ab: Start eroeffnet die
// Transaktion (processType/processData sind laut DSFinV-K bei Start immer
// leer), Finish schliesst sie mit dem finalen Schema ab. Beide Aufrufe
// adressieren die Transaktion ueber die von jotti erzeugte tx-ID (UUIDv4).
type TSEClient interface {
	StartTransaction(ctx context.Context, txID string) (StartResult, error)
	FinishTransaction(ctx context.Context, txID string, processType string, processData string) (FinishResult, error)
}

// ConnectionTester prueft eine konfigurierte TSE. TestConnection ist die volle
// Diagnose (TSS- und Client-Abruf, Seriennummer); Umgebung ist der leichte Pfad
// fuer reine Statusanzeigen und kommt allein aus dem Auth-Token, ohne TSS-/
// Client-Abruf.
type ConnectionTester interface {
	TestConnection(ctx context.Context) (VerbindungStatus, error)
	Umgebung(ctx context.Context) (Umgebung, error)
}

// TransactionRetriever fragt den Ist-Zustand einer Transaktion bei der TSE ab.
// Existiert die Transaktion nicht, wird ErrTransactionNichtGefunden geliefert.
type TransactionRetriever interface {
	RetrieveTransaction(ctx context.Context, txID string) (RetrieveResult, error)
}

type TransactionState string

const (
	TransactionStateActive    TransactionState = "ACTIVE"
	TransactionStateFinished  TransactionState = "FINISHED"
	TransactionStateCancelled TransactionState = "CANCELLED"
)

// RetrieveResult ist der bei der TSE gespeicherte Stand einer Transaktion:
// ihr Zustand plus — bei abgeschlossenen Transaktionen — die Signaturdaten
// in derselben Form wie ein FinishResult.
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

// VerbindungStatus ist das Ergebnis des Verbindungstests. Umgebung, TSSState,
// ClientState und ClientSerialNumber werden von der Repository-Schicht aus den
// fiskaly-Antworten befuellt. SeriennummerKorrekt setzt die Application-Schicht,
// die die jotti-Kassen-Seriennummer kennt und sie mit der Client-serial_number
// abgleicht.
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
