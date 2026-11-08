package host

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/nicograef/jotti/windows/starter/core"
)

// CheckPorts prueft, ob die Ports frei sind, und liefert fuer den ersten belegten
// die Diagnose; leerer String = alle frei.
func CheckPorts(ports []int) string {
	for _, port := range ports {
		if portAvailable(port) {
			continue
		}
		return core.PortBelegtDiagnose(port, lookupPortOwners(port))
	}
	return ""
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
