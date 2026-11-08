package host

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FirewallRule beschreibt eine eingehende TCP-Freigabe aus dem lokalen Subnetz.
// Program leer heisst: die Regel gilt fuer jedes Programm.
type FirewallRule struct {
	Name    string
	Ports   []int
	Program string
}

// EnsureFirewallRule setzt die Regel idempotent. Ein Fehlschlag ist kein
// Abbruchgrund — nur eine Warnung mit manuellem Hinweis. Meldet, ob die Regel
// danach besteht.
func EnsureFirewallRule(rule FirewallRule) bool {
	if firewallRuleExists(rule.Name) {
		return true
	}
	if err := exec.Command("netsh", addFirewallRuleArgs(rule)...).Run(); err != nil {
		fmt.Printf("WARNUNG: Firewall-Regel konnte nicht automatisch gesetzt werden (%v). "+
			"Bitte eingehende Verbindungen auf TCP %s manuell erlauben.\n", err, portList(rule.Ports, " und "))
		return false
	}
	return true
}

func firewallRuleExists(name string) bool {
	return exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name="+name).Run() == nil
}

func addFirewallRuleArgs(rule FirewallRule) []string {
	args := []string{"advfirewall", "firewall", "add", "rule",
		"name=" + rule.Name, "dir=in", "action=allow", "protocol=TCP",
		"localport=" + portList(rule.Ports, ","), "remoteip=localsubnet", "profile=any"}
	if rule.Program != "" {
		args = append(args, "program="+rule.Program)
	}
	return args
}

func portList(ports []int, sep string) string {
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = strconv.Itoa(p)
	}
	return strings.Join(parts, sep)
}
