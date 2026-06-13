package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"jotti-starter/core"
)

// DockerCliPath ist die Docker-Desktop-CLI fuer den Engine-Wechsel; sie liegt am
// Standardpfad neben "Docker Desktop.exe" (core.DockerDesktopPath).
const DockerCliPath = `C:\Program Files\Docker\Docker\DockerCli.exe`

// Timeouts fuer die Auto-Fix- und Health-Schleifen; grosszuegig fuer den
// WSL2-/VM-Kaltstart und die Erst-Migrationen auf Altgeraeten.
const (
	dockerStartTimeout = 120 * time.Second
	healthTimeout      = 120 * time.Second
)

// ensureDocker stellt sicher, dass der Docker-Daemon im Linux-Container-Modus
// antwortet, und repariert die haeufigen Stoerungen selbst (Admin-Rechte
// vorausgesetzt). Rueckgabe: leerer String bei Erfolg, sonst eine deutsche
// Diagnose mit Handlungshinweis.
func ensureDocker() string {
	if _, err := exec.LookPath("docker"); err != nil {
		return core.DiagnoseDockerCLIFehlt
	}

	osType, err := dockerOSType()
	if err != nil {
		// Daemon antwortet nicht — Docker Desktop selbst starten, falls installiert.
		if ok, _ := fileExists(core.DockerDesktopPath); !ok {
			return core.DiagnoseDockerNichtInstalliert
		}
		fmt.Println("Docker Desktop wird gestartet ...")
		if err := startDockerDesktop(); err != nil {
			return core.DiagnoseDockerStartFehlgeschlagen
		}
		if !waitForDockerDaemon(dockerStartTimeout) {
			return core.DiagnoseDockerStartFehlgeschlagen
		}
		osType, err = dockerOSType()
		if err != nil {
			return core.DiagnoseDockerStartFehlgeschlagen
		}
	}

	if osType == "windows" {
		fmt.Println("Docker laeuft im Windows-Container-Modus — schalte auf Linux-Container um ...")
		if err := switchToLinuxEngine(); err != nil {
			return core.DiagnoseEngineSwitchFehlgeschlagen
		}
		if osType, err = dockerOSType(); err != nil || osType == "windows" {
			return core.DiagnoseEngineSwitchFehlgeschlagen
		}
	}

	return ""
}

// dockerOSType liefert den Container-Modus des Daemons ("linux"/"windows"); ein
// Fehler bedeutet, dass der Daemon (noch) nicht antwortet.
func dockerOSType() (string, error) {
	out, err := exec.Command("docker", "info", "-f", "{{.OSType}}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// startDockerDesktop startet die GUI-Anwendung detached; sie faehrt den Daemon
// im Hintergrund hoch.
func startDockerDesktop() error {
	return exec.Command(core.DockerDesktopPath).Start()
}

// waitForDockerDaemon pollt "docker info" mit Fortschrittsanzeige, bis der Daemon
// antwortet oder timeout ablaeuft.
func waitForDockerDaemon(timeout time.Duration) bool {
	fmt.Print("Warte auf den Docker-Daemon ")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := exec.Command("docker", "info").Run(); err == nil {
			fmt.Println(" bereit.")
			return true
		}
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
	fmt.Println()
	return false
}

// switchToLinuxEngine schaltet Docker Desktop auf Linux-Container um.
func switchToLinuxEngine() error {
	return exec.Command(DockerCliPath, "-SwitchLinuxEngine").Run()
}

// checkPorts prueft, ob 80/443 frei sind. Laeuft der eigene reverse-proxy bereits
// (Day-2-Pfad), ist die Belegung der Erfolgs-Fall und kein Fehlalarm. Rueckgabe:
// leerer String wenn frei (oder eigener Stack), sonst die Belegt-Diagnose mit
// — wenn ermittelbar — exaktem Verursacher.
func checkPorts(composePath string) string {
	if reverseProxyRunning(composePath) {
		fmt.Println("Der jotti-Stack laeuft bereits — Start ist idempotent.")
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

// reverseProxyRunning meldet, ob der eigene reverse-proxy-Container schon laeuft.
func reverseProxyRunning(composePath string) bool {
	out, err := exec.Command("docker", "compose", "-f", composePath, "ps", "-q", "--status", "running", "reverse-proxy").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// portAvailable prueft per net.Listen, ob der TCP-Port frei ist.
func portAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// lookupPortOwners ermittelt per Get-NetTCPConnection (JSON) den haltenden
// Prozess eines belegten Ports. Schlaegt der Lookup oder das Parsen fehl, liefert
// die Funktion nil — die Diagnose faellt dann auf den generischen Fallback zurueck.
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

// ensureFirewall setzt idempotent die eingehende Freigabe fuer TCP 80/443 (aufs
// lokale Subnetz beschraenkt, profilunabhaengig). Ein Fehlschlag ist kein
// Abbruchgrund — nur eine Warnung mit manuellem Hinweis.
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

// detectLANIP ermittelt die LAN-IP fuer die LAN_IP-Env des Caddy-Containers.
// Schlaegt die Erkennung fehl, laeuft der Start trotzdem weiter — Caddy rendert
// dann nur die Fallback-Site.
func detectLANIP() string {
	ip, err := core.SelectLANIP(outboundIP(), localInterfaces())
	if err != nil {
		fmt.Printf("Hinweis: LAN-IP konnte nicht ermittelt werden (%v) — die Zugangsadresse fuers WLAN "+
			"erscheint erst, sobald eine LAN-IP erkannt wird.\n", err)
		return ""
	}
	fmt.Printf("LAN-IP: %s\n", ip)
	return ip
}

// outboundIP liefert die IP des Default-Route-Interfaces ueber einen UDP-"Connect"
// (es wird kein Paket gesendet). Leerer String, wenn keine Route existiert.
func outboundIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return addr.IP.String()
	}
	return ""
}

// localInterfaces sammelt die IPv4-Adressen aller Interfaces fuer die
// Fallback-Heuristik in core.SelectLANIP.
func localInterfaces() []core.NetInterface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	result := make([]core.NetInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		var ips []string
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					ips = append(ips, ip4.String())
				}
			}
		}
		if len(ips) > 0 {
			result = append(result, core.NetInterface{Name: iface.Name, IPs: ips})
		}
	}
	return result
}

// configVolume ist das von Compose verwaltete jotti-config-Volume. Compose
// stellt benannten Volumes den Projektnamen (jotti-local) voran, daher lautet der
// Host-Name, den der Starter ansprechen muss, "jotti-local_jotti-config". Das
// Volume ist in der Compose read-only in postgres gemountet — das bindet seinen
// Lebenszyklus an die Daten: `docker compose down -v` entfernt beide zusammen, das
// Secret kann die Daten, die es entsperrt, also nie ueberleben (kein Lockout).
const configVolume = "jotti-local_jotti-config"

// configVolumePath ist der Pfad der gespiegelten .env im Volume.
const configVolumePath = "/config/.env"

// configHelperImage liest/schreibt das Volume in einem Wegwerf-Container. Es ist
// bewusst dasselbe postgres-Image wie im Stack (keine zusaetzliche Abhaengigkeit,
// nach dem ersten `up` ohnehin lokal vorhanden) — beim Bump in den Compose-Dateien
// hier mitziehen, sonst wird ein zweites Image gezogen.
const configHelperImage = "postgres:17.8"

// errSecretFehltMitDaten signalisiert den Fail-Safe-Abbruch: vorhandene Daten, aber
// nirgends ein Secret. Die ausfuehrliche Anleitung gibt materializeEnvFromVolume
// selbst aus (wie die Preflight-Diagnosen) — run() beendet daraufhin nur noch mit
// Code 1, ohne die Meldung mit einem Fehlerpraefix zu doppeln.
var errSecretFehltMitDaten = errors.New("start abgebrochen: keine Zugangsdaten zu vorhandenen Daten gefunden")

// materializeEnvFromVolume macht das jotti-config-Volume zur Quelle der Wahrheit
// fuers Install-Secret und schreibt den Host-.env-Spiegel, den `compose
// --env-file` und das Relay lesen. Laeuft nur unter Windows und nach ensureDocker
// (der Volume-Read braucht einen laufenden Daemon). Das Secret wird in fester
// Reihenfolge gesucht (Volume → ordnerlokale Kandidaten in localDirs); ein adoptierter
// Treffer wird ins Volume geschrieben. Wird nichts gefunden, obwohl bereits Daten
// existieren, bricht der Start ab (Fail-Safe), statt frische Secrets neben die Daten
// zu erzeugen und sie damit auszusperren.
func materializeEnvFromVolume(envPath string, localDirs []string) error {
	volumeContent, err := readConfigVolume()
	if err != nil {
		return err
	}
	dataExists, err := volumeExists(postgresDataVolume)
	if err != nil {
		return err
	}
	res := core.ResolveEnv(volumeContent, readEnvCandidates(localDirs), dataExists)
	if res.Abort {
		fmt.Println(core.DiagnoseSecretFehltMitDaten)
		return errSecretFehltMitDaten
	}
	if res.Seed {
		if err := writeConfigVolume(res.Content); err != nil {
			return err
		}
		fmt.Println("Zugangsdaten im jotti-Datentresor gesichert.")
	}
	return writeEnvFile(envPath, []byte(res.Content))
}

// readEnvCandidates liest den .env-Inhalt aus jedem Kandidatenverzeichnis in
// Reihenfolge. Ein fehlender oder unlesbarer Eintrag liefert leeren Inhalt — leere
// Kandidaten ueberspringt die Auswahl in core.ResolveEnv.
func readEnvCandidates(dirs []string) []string {
	candidates := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		data, _ := os.ReadFile(filepath.Join(dir, ".env"))
		candidates = append(candidates, string(data))
	}
	return candidates
}

// readConfigVolume liest die gespiegelte .env aus dem jotti-config-Volume ueber
// einen Wegwerf-Container. Ein leerer String bedeutet "kein Secret vorhanden":
// entweder fehlt das Volume (Erststart) oder es existiert, enthaelt aber noch
// keine .env. Nur ein echter Docker-Fehler (Daemon/Image) wird durchgereicht.
func readConfigVolume() (string, error) {
	if exec.Command("docker", "volume", "inspect", configVolume).Run() != nil {
		return "", nil // Volume fehlt → Erststart, ohne ein leeres Volume anzulegen
	}
	out, err := exec.Command("docker", "run", "--rm", "--entrypoint", "cat",
		"-v", configVolume+":/config", configHelperImage, configVolumePath).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil // Volume vorhanden, .env aber (noch) nicht geschrieben
		}
		return "", fmt.Errorf("lesen aus dem Datentresor fehlgeschlagen: %w", err)
	}
	return string(out), nil
}

// writeConfigVolume schreibt content in die .env des jotti-config-Volumes. Das
// `docker run -v` legt das Volume bei Bedarf an; Compose verwendet danach dasselbe
// (gleicher Name).
func writeConfigVolume(content string) error {
	cmd := exec.Command("docker", "run", "--rm", "-i", "--entrypoint", "sh",
		"-v", configVolume+":/config", configHelperImage, "-c", "cat > "+configVolumePath)
	cmd.Stdin = strings.NewReader(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schreiben in den Datentresor fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// composeUp faehrt den Stack hoch und reicht die Ausgabe live durch (der Pull
// bleibt sichtbar). LAN_IP wird nur gesetzt, wenn es erkannt wurde — ohne LAN_IP
// rendert Caddy nur die Fallback-Site. Der Proxy wird danach neu erzeugt, weil
// das Caddyfile nur im Entrypoint aus LAN_IP gerendert wird (wie make local-up).
func composeUp(composePath, envPath, lanIP string) error {
	env := os.Environ()
	if lanIP != "" {
		env = append(env, "LAN_IP="+lanIP)
	}
	if err := runCompose(env, composePath, envPath, "up", "-d", "--build"); err != nil {
		return err
	}
	return runCompose(env, composePath, envPath, "up", "-d", "--no-deps", "--force-recreate", "reverse-proxy")
}

// runCompose ruft `docker compose` mit explizitem --env-file auf den Host-Spiegel
// auf. Nach der UAC-Elevation ist das Arbeitsverzeichnis System32, deshalb wird
// die .env-Quelle fuer die ${...}-Interpolation explizit benannt statt implizit
// aus dem Projektverzeichnis geladen.
func runCompose(env []string, composePath, envPath string, args ...string) error {
	full := append([]string{"compose", "-f", composePath, "--env-file", envPath}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// waitForHealth pollt https://localhost/api/health (TLS-Verify aus — localhost
// trifft Caddys interne CA), bis HTTP 200 kommt oder healthTimeout ablaeuft. Nur
// 200 gilt als bereit; das Backend liefert 503, solange die DB nicht antwortet.
func waitForHealth() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Print("Warte darauf, dass jotti bereit ist ")
	deadline := time.Now().Add(healthTimeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://localhost/api/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println(" bereit.")
				return nil
			}
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
	fmt.Println()
	return fmt.Errorf("jotti wurde nicht innerhalb von %s bereit. Bitte die Logs pruefen "+
		"(docker compose logs) und jotti erneut starten", healthTimeout)
}
