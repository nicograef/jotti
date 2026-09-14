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
// (noch) nicht existiert — der Signatur-Worker startet sie dann neu.
var ErrTransactionNichtGefunden = errors.New("tse transaction not found")

// AuftragsFehler kennzeichnet einen auftragsspezifischen Signierfehler (etwa von
// fiskaly zurückgewiesene processData): der Worker verbucht einen Fehlversuch und
// überspringt den Auftrag, ein Gift-Auftrag staut nie die Queue. Jeder nicht so
// gekennzeichnete Fehler gilt als TSE-weit, bricht den Durchlauf ab und schaltet
// den Worker in den Störungszustand.
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

// TSEClient bildet das atomare Transaktionsmuster ab: Start eröffnet die
// Transaktion (processType/processData sind laut DSFinV-K bei Start immer
// leer), Finish schließt sie mit dem finalen Schema ab. Beide Aufrufe
// adressieren die Transaktion über die von jotti erzeugte tx-ID (UUIDv4).
type TSEClient interface {
	StartTransaction(ctx context.Context, txID string) (StartResult, error)
	FinishTransaction(ctx context.Context, txID string, processType string, processData string) (FinishResult, error)
}

// ConnectionTester: TestConnection ist die volle Diagnose (TSS- und Client-Abruf,
// Seriennummer), Umgebung der leichte Pfad allein aus dem Auth-Token.
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

// RetrieveResult ist der bei der TSE gespeicherte Stand einer Transaktion; die
// Signaturdaten trägt es nur bei abgeschlossenen Transaktionen.
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

// VerbindungStatus: Umgebung, TSSState, ClientState und ClientSerialNumber füllt
// die Repository-Schicht aus den fiskaly-Antworten; SeriennummerKorrekt setzt die
// Application-Schicht, die die jotti-Kassen-Seriennummer mit der
// Client-serial_number abgleicht.
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
