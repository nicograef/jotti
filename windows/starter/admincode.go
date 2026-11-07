package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nicograef/jotti/windows/starter/core"
)

func containerStartedAt(container string) string {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.StartedAt}}", container).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// containerLogsSince liest die Container-Logs ab startedAt. CombinedOutput, damit
// der Grep unabhaengig vom Stream greift; ein Fehler ergibt die (moeglicherweise
// leere) Ausgabe.
func containerLogsSince(container, startedAt string) string {
	out, _ := exec.Command("docker", "logs", "--since", startedAt, container).CombinedOutput()
	return string(out)
}

// readAdminOTP liefert den Klartext-Code des Initial-Admins aus den Backend-Logs
// SEIT dem aktuellen Container-Start — so verschwindet ein veralteter Marker, sobald
// der Container nach abgeschlossener Einrichtung neu erstellt wurde. Jeder Fehler
// ergibt ("", false) und ist nie fatal.
func readAdminOTP() (string, bool) {
	startedAt := containerStartedAt(backendContainer)
	if startedAt == "" {
		return "", false
	}
	return core.ParseAdminOTP(containerLogsSince(backendContainer, startedAt))
}

func printAdminCode() {
	fmt.Println()
	fmt.Println(core.AdminCodeHinweis(readAdminOTP()))
}
