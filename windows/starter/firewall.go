package main

import (
	"fmt"
	"os/exec"
)

// ensureFirewall setzt idempotent die eingehende Freigabe fuer TCP 80/443. Ein
// Fehlschlag ist kein Abbruchgrund — nur eine Warnung mit manuellem Hinweis.
func ensureFirewall() {
	if firewallRuleExists() {
		return
	}
	if err := addFirewallRule(); err != nil {
		fmt.Printf("WARNUNG: Firewall-Regel konnte nicht automatisch gesetzt werden (%v). "+
			"Bitte eingehende Verbindungen auf TCP 80 und 443 manuell erlauben.\n", err)
	}
}

func firewallRuleExists() bool {
	return exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=jotti").Run() == nil
}

func addFirewallRule() error {
	return exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=jotti", "dir=in", "action=allow", "protocol=TCP",
		"localport=80,443", "remoteip=localsubnet", "profile=any").Run()
}
