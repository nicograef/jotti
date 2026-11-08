package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nicograef/jotti/windows/starter/host"
)

// checkPorts prueft, ob 80/443 frei sind; leerer String = frei. Laeuft der eigene
// reverse-proxy schon, ist die Belegung der Erfolgs-Fall und kein Fehlalarm.
func checkPorts() string {
	if reverseProxyRunning() {
		fmt.Println("Der jotti-Stack laeuft bereits - Start ist idempotent.")
		return ""
	}
	return host.CheckPorts([]int{80, 443})
}

func reverseProxyRunning() bool {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", reverseProxyContainer).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}
