package core

import (
	"fmt"
	"strings"
)

// Standardpfade der Docker-Desktop-Installation unter Windows.
const (
	DockerBinPath     = `C:\Program Files\Docker\Docker\resources\bin`
	DockerDesktopPath = `C:\Program Files\Docker\Docker\Docker Desktop.exe`
)

// Preflight-Diagnosen, ASCII-transliteriert fuer die Windows-Konsole; jede nennt
// den naechsten Handlungsschritt.
const (
	DiagnoseDockerCLIFehlt = "Docker wurde nicht gefunden. Bitte Docker Desktop installieren und sicherstellen, dass \"" +
		DockerBinPath + "\" im PATH liegt, dann jotti erneut starten."

	DiagnoseDockerNichtInstalliert = "Docker Desktop ist nicht installiert (erwartet unter \"" +
		DockerDesktopPath + "\"). Bitte Docker Desktop installieren und jotti erneut starten."

	DiagnoseDockerStartFehlgeschlagen = "Docker Desktop antwortet nicht. Bitte Docker Desktop manuell starten, " +
		"warten bis das Wal-Symbol ruhig steht, und jotti erneut starten."

	DiagnoseEngineSwitchFehlgeschlagen = "Docker laeuft im Windows-Container-Modus und konnte nicht automatisch auf " +
		"Linux-Container umgeschaltet werden. Bitte in Docker Desktop \"Switch to Linux containers\" waehlen und jotti erneut starten."

	// DiagnoseSecretFehltMitDaten ist die Fail-Safe-Meldung: Daten ohne Secret. jotti
	// bricht ab, statt frische Secrets neben die Daten zu erzeugen — das wuerde sie
	// aussperren.
	DiagnoseSecretFehltMitDaten = "Es sind bereits jotti-Daten vorhanden, aber es wurden keine Zugangsdaten (.env) gefunden. " +
		"jotti startet NICHT, um die vorhandenen Daten nicht mit neuen, falschen Zugangsdaten auszusperren.\n" +
		"Gesucht wurde im jotti-Datentresor, unter \"C:\\ProgramData\\jotti\\.env\" und neben jotti-start.exe.\n" +
		"Eine .env einer anderen jotti-Installation nach \"C:\\ProgramData\\jotti\\.env\" " +
		"kopieren und jotti erneut starten."
)

const typischePortVerursacher = "z. B. VMware Workstation, IIS (World Wide Web Publishing Service) oder Skype"

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

// DiagnoseZustandsverzeichnisFremd verweigert ein vorgefundenes Zustandsverzeichnis,
// das weder jotti noch ein Administrator angelegt hat; befund nennt den Grund.
func DiagnoseZustandsverzeichnisFremd(dir, befund string) string {
	return fmt.Sprintf("Start verweigert: Der Ordner \"%s\" %s. "+
		"Ein anderer Benutzer dieses PCs koennte ihn angelegt und darin Zugangsdaten hinterlegt haben.\n"+
		"  Bitte den Ordner im Explorer loeschen (Administrator-Abfrage bestaetigen) und jotti-start.exe erneut starten. "+
		"Einen Unterordner backups vorher an einen sicheren Ort kopieren.", dir, befund)
}

// DiagnoseZustandsverzeichnisRechte meldet, dass die Zugriffsrechte des
// Zustandsverzeichnisses nicht geprueft oder gesetzt werden konnten.
func DiagnoseZustandsverzeichnisRechte(dir string, err error) string {
	return fmt.Sprintf("Start verweigert: Die Zugriffsrechte fuer \"%s\" konnten nicht eingerichtet werden (%v). "+
		"Ohne sie koennten andere Benutzer dieses PCs die Zugangsdaten und Backups lesen.\n"+
		"  Bitte jotti-start.exe erneut starten und die Administrator-Abfrage bestaetigen. "+
		"Bleibt der Fehler, den Ordner im Explorer loeschen (einen Unterordner backups vorher an einen sicheren Ort kopieren) "+
		"und jotti-start.exe erneut starten.", dir, err)
}
