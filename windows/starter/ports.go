package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/nicograef/jotti/windows/starter/core"
)

// checkPorts prueft, ob 80/443 frei sind; leerer String = frei. Laeuft der eigene
// reverse-proxy schon, ist die Belegung der Erfolgs-Fall und kein Fehlalarm.
func checkPorts(composePath string) string {
	if reverseProxyRunning(composePath) {
		fmt.Println("Der jotti-Stack laeuft bereits - Start ist idempotent.")
		return ""
	}

	for _, port := range []int{80, 443} {
		if portAvailable(port) {
			continue
		}
		return core.PortBelegtDiagnose(port, lookupPortOwners(port))
	}

	return ""
}

func reverseProxyRunning(composePath string) bool {
	out, err := exec.Command("docker", "compose", "-f", composePath, "ps", "-q", "--status", "running", "reverse-proxy").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func portAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// lookupPortOwners ermittelt per Get-NetTCPConnection den haltenden Prozess eines
// belegten Ports; nil bei Fehlschlag — dann greift die generische Diagnose.
func lookupPortOwners(port int) []core.PortOwner {
	script := fmt.Sprintf(
		"Get-NetTCPConnection -State Listen -LocalPort %d -ErrorAction SilentlyContinue | "+
			"Select-Object LocalPort,OwningProcess,@{n='ProcessName';e={(Get-Process -Id $_.OwningProcess).ProcessName}} | "+
			"ConvertTo-Json", port)
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return nil
	}
	owners, err := core.ParsePortOwners(out)
	if err != nil {
		return nil
	}
	return owners
}
