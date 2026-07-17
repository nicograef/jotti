package core

import (
	"fmt"
	"strings"
)

// Standardpfade der Docker-Desktop-Installation unter Windows; die Diagnosen
// nennen sie, der Shell-Layer (windows/starter, das ausfuehrbare main-Paket)
// startet die Anwendung darueber.
const (
	DockerBinPath     = `C:\Program Files\Docker\Docker\resources\bin`
	DockerDesktopPath = `C:\Program Files\Docker\Docker\Docker Desktop.exe`
)

// Statische Preflight-Diagnosen, deutsch und ASCII-transliteriert fuer die
// Windows-Konsole (wie die Laufzeit-Strings in windows/relay). Jede nennt den
// naechsten Handlungsschritt.
const (
	DiagnoseDockerCLIFehlt = "Docker wurde nicht gefunden. Bitte Docker Desktop installieren und sicherstellen, dass \"" +
		DockerBinPath + "\" im PATH liegt, dann jotti erneut starten."

	DiagnoseDockerNichtInstalliert = "Docker Desktop ist nicht installiert (erwartet unter \"" +
		DockerDesktopPath + "\"). Bitte Docker Desktop installieren und jotti erneut starten."

	DiagnoseDockerStartFehlgeschlagen = "Docker Desktop antwortet nicht. Bitte Docker Desktop manuell starten, " +
		"warten bis das Wal-Symbol ruhig steht, und jotti erneut starten."

	DiagnoseEngineSwitchFehlgeschlagen = "Docker laeuft im Windows-Container-Modus und konnte nicht automatisch auf " +
		"Linux-Container umgeschaltet werden. Bitte in Docker Desktop \"Switch to Linux containers\" waehlen und jotti erneut starten."

	// DiagnoseSecretFehltMitDaten ist die Fail-Safe-Meldung: es gibt bereits Daten,
	// aber an keinem Suchort ein Install-Secret. jotti bricht ab, statt frische
	// Secrets neben die Daten zu erzeugen (das wuerde sie aussperren), und nennt
	// die gesuchten Orte samt Rettungsweg (alte .env an den kanonischen Ort legen).
	DiagnoseSecretFehltMitDaten = "Es sind bereits jotti-Daten vorhanden, aber es wurden keine Zugangsdaten (.env) gefunden. " +
		"jotti startet NICHT, um die vorhandenen Daten nicht mit neuen, falschen Zugangsdaten auszusperren.\n" +
		"Gesucht wurde im jotti-Datentresor, unter \"%PROGRAMDATA%\\jotti\\.env\" und neben jotti-start.exe.\n" +
		"Bitte die .env aus der vorherigen jotti-Installation (frueher im Programmordner neben jotti-start.exe) " +
		"nach \"%PROGRAMDATA%\\jotti\\.env\" kopieren und jotti erneut starten."
)

// typischePortVerursacher nennt haeufige Beleger von 80/443 fuer den Fall, dass
// der genaue Verursacher nicht ermittelt werden konnte.
const typischePortVerursacher = "z. B. VMware Workstation, IIS (World Wide Web Publishing Service) oder Skype"

// PortBelegtDiagnose erzeugt die deutsche Diagnose fuer einen belegten Port.
// Sind die Verursacher bekannt, werden Prozessname und PID genannt; sonst greift
// der generische Fallback mit typischen Verursachern.
func PortBelegtDiagnose(port int, owners []PortOwner) string {
	if len(owners) == 0 {
		return fmt.Sprintf(
			"Port %d ist bereits belegt, der Verursacher konnte nicht ermittelt werden (%s). "+
				"Bitte das betreffende Programm beenden und jotti erneut starten.",
			port, typischePortVerursacher)
	}

	beschreibungen := make([]string, 0, len(owners))
	for _, o := range owners {
		name := strings.TrimSpace(o.ProcessName)
		if name == "" {
			name = "unbekanntes Programm"
		}
		beschreibungen = append(beschreibungen, fmt.Sprintf("'%s' (PID %d)", name, o.PID))
	}
	return fmt.Sprintf(
		"Port %d ist durch %s belegt. Bitte dieses Programm beenden und jotti erneut starten.",
		port, strings.Join(beschreibungen, ", "))
}
