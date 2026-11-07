//go:build !windows

package main

import (
	"fmt"
	"os"
)

// prepareStateDir legt im Linux-Dev-Lauf nur den ordnerlokalen Zustand an; die
// Windows-Zugriffsrechte setzt statedir_windows.go.
func prepareStateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil { //nolint:gosec // the state dir comes from PROGRAMDATA or the compose dir, the operator's own environment
		return fmt.Errorf("%s konnte nicht angelegt werden: %w", dir, err)
	}
	return nil
}
