package core

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type PortOwner struct {
	LocalPort   int
	PID         int
	ProcessName string
}

// ParsePortOwners liest die JSON-Ausgabe von Get-NetTCPConnection. PowerShell
// liefert bei genau einem Treffer ein Objekt, bei mehreren ein Array; ein nicht
// aufloesbarer Prozessname (JSON null) wird zu "". Leere Ausgabe ergibt eine leere
// Liste, unlesbares JSON einen Fehler — dann greift die generische Port-Diagnose.
func ParsePortOwners(jsonOut []byte) ([]PortOwner, error) {
	trimmed := bytes.TrimPrefix(bytes.TrimSpace(jsonOut), []byte("\xef\xbb\xbf"))
	trimmed = bytes.TrimSpace(trimmed)
	if len(trimmed) == 0 {
		return nil, nil
	}

	type rawOwner struct {
		LocalPort     int    `json:"LocalPort"`
		OwningProcess int    `json:"OwningProcess"`
		ProcessName   string `json:"ProcessName"`
	}

	var raws []rawOwner
	switch trimmed[0] {
	case '[':
		if err := json.Unmarshal(trimmed, &raws); err != nil {
			return nil, fmt.Errorf("Port-Belegung (JSON-Array) nicht lesbar: %w", err)
		}
	case '{':
		var single rawOwner
		if err := json.Unmarshal(trimmed, &single); err != nil {
			return nil, fmt.Errorf("Port-Belegung (JSON-Objekt) nicht lesbar: %w", err)
		}
		raws = []rawOwner{single}
	default:
		return nil, fmt.Errorf("unerwartete Port-Belegungs-Ausgabe: %q", trimmed)
	}

	owners := make([]PortOwner, 0, len(raws))
	for _, r := range raws {
		owners = append(owners, PortOwner{
			LocalPort:   r.LocalPort,
			PID:         r.OwningProcess,
			ProcessName: r.ProcessName,
		})
	}
	return owners, nil
}
