//go:build !windows

package host

import (
	"fmt"
	"os"
)

// PrepareStateDir legt im Linux-Dev-Lauf nur den ordnerlokalen Zustand an; die
// Windows-Zugriffsrechte setzt statedir_windows.go.
func PrepareStateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("%s konnte nicht angelegt werden: %w", dir, err)
	}
	return nil
}

// PreparePrivateDir legt im Linux-Dev-Lauf einen nur fuer den Besitzer lesbaren Ordner an.
func PreparePrivateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("%s konnte nicht angelegt werden: %w", dir, err)
	}
	return nil
}
