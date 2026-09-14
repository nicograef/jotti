package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nicograef/jotti/windows/starter/core"
)

func backendContainerID(composePath string) string {
	out, err := exec.Command("docker", "compose", "-f", composePath, "ps", "-q", "backend").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func containerStartedAt(cid string) string {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.StartedAt}}", cid).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// containerLogsSince liest die Container-Logs ab startedAt. CombinedOutput, damit
// der Grep unabhaengig vom Stream greift; ein Fehler ergibt die (moeglicherweise
// leere) Ausgabe.
func containerLogsSince(cid, startedAt string) string {
	out, _ := exec.Command("docker", "logs", "--since", startedAt, cid).CombinedOutput()
	return string(out)
}

// readAdminOTP liefert den Klartext-Code des Initial-Admins aus den Backend-Logs
// SEIT dem aktuellen Container-Start — so verschwindet ein veralteter Marker, sobald
// der Container nach abgeschlossener Einrichtung neu erstellt wurde. Jeder Fehler
// ergibt ("", false) und ist nie fatal.
func readAdminOTP(composePath string) (string, bool) {
	cid := backendContainerID(composePath)
	if cid == "" {
		return "", false
	}
	startedAt := containerStartedAt(cid)
	if startedAt == "" {
		return "", false
	}
	return core.ParseAdminOTP(containerLogsSince(cid, startedAt))
}

func printAdminCode(composePath string) {
	fmt.Println()
	fmt.Println(core.AdminCodeHinweis(readAdminOTP(composePath)))
}
