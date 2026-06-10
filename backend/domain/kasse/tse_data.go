package kasse

import (
	"fmt"
	"strings"
)

// TSEData stores fiscal signature data in event payloads.
// JSON tags are allowed here because event-data structs are persisted in the event store.
type TSEData struct {
	TransactionNumber int    `json:"tseTransactionNumber"`
	SignatureCounter  int    `json:"tseSignatureCounter"`
	SerialNumberTSE   string `json:"tseSerialNumber"`
	LogTimeStart      string `json:"tseLogTimeStart"`
	LogTimeEnd        string `json:"tseLogTimeEnd"`
	Signature         string `json:"tseSignature"`
	ProcessType       string `json:"tseProcessType"`
	QRCodeData        string `json:"tseQrCodeData,omitempty"`
}

func (t TSEData) Validate() error {
	if t.TransactionNumber <= 0 {
		return fmt.Errorf("tse transaction number must be > 0")
	}
	if t.SignatureCounter < 0 {
		return fmt.Errorf("tse signature counter must be >= 0")
	}
	if strings.TrimSpace(t.SerialNumberTSE) == "" {
		return fmt.Errorf("tse serial number is required")
	}
	if strings.TrimSpace(t.LogTimeStart) == "" {
		return fmt.Errorf("tse log time start is required")
	}
	if strings.TrimSpace(t.LogTimeEnd) == "" {
		return fmt.Errorf("tse log time end is required")
	}
	if strings.TrimSpace(t.Signature) == "" {
		return fmt.Errorf("tse signature is required")
	}
	if strings.TrimSpace(t.ProcessType) == "" {
		return fmt.Errorf("tse process type is required")
	}

	return nil
}
