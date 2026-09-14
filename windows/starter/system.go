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

	"github.com/nicograef/jotti/windows/starter/core"
)

const DockerCliPath = `C:\Program Files\Docker\Docker\DockerCli.exe`

// Grosszuegig fuer den WSL2-/VM-Kaltstart und die Erst-Migrationen auf Altgeraeten.
const (
	dockerStartTimeout = 120 * time.Second
	healthTimeout      = 120 * time.Second
)

// ensureDocker stellt den Docker-Daemon im Linux-Container-Modus sicher und
// repariert haeufige Stoerungen selbst. Leerer String = Erfolg, sonst die Diagnose.
func ensureDocker() string {
	if _, err := exec.LookPath("docker"); err != nil {
		return core.DiagnoseDockerCLIFehlt
	}

	osType, err := dockerOSType()
	if err != nil {
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
		fmt.Println("Docker laeuft im Windows-Container-Modus - schalte auf Linux-Container um ...")
		if err := switchToLinuxEngine(); err != nil {
			return core.DiagnoseEngineSwitchFehlgeschlagen
		}
		if osType, err = dockerOSType(); err != nil || osType == "windows" {
			return core.DiagnoseEngineSwitchFehlgeschlagen
		}
	}

	return ""
}

func dockerOSType() (string, error) {
	out, err := exec.Command("docker", "info", "-f", "{{.OSType}}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func startDockerDesktop() error {
	return exec.Command(core.DockerDesktopPath).Start()
}

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

func switchToLinuxEngine() error {
	return exec.Command(DockerCliPath, "-SwitchLinuxEngine").Run()
}

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

// detectLANIP ermittelt die LAN-IP fuer die LAN_IP-Env des Caddy-Containers; ohne
// sie laeuft der Start weiter, Caddy rendert dann nur die Fallback-Site.
func detectLANIP() string {
	ip, err := core.SelectLANIP(outboundIP(), localInterfaces())
	if err != nil {
		fmt.Printf("Hinweis: LAN-IP konnte nicht ermittelt werden (%v) - die Zugangsadresse fuers WLAN "+
			"erscheint erst, sobald eine LAN-IP erkannt wird.\n", err)
		return ""
	}
	fmt.Printf("LAN-IP: %s\n", ip)
	return ip
}

// outboundIP liefert die IP des Default-Route-Interfaces ueber einen UDP-"Connect"
// (es wird kein Paket gesendet).
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

// configVolume ist das von Compose verwaltete jotti-config-Volume. Compose stellt
// benannten Volumes den Projektnamen (jotti-local) voran, daher lautet der Name,
// den der Starter ansprechen muss, "jotti-local_jotti-config".
const configVolume = "jotti-local_jotti-config"

const configVolumePath = "/config/.env"

// configHelperImage liest/schreibt das Volume in einem Wegwerf-Container: bewusst
// dasselbe postgres-Image wie im Stack — beim Bump in den Compose-Dateien hier
// mitziehen, sonst wird ein zweites Image gezogen.
const configHelperImage = "postgres:17.8"

// errSecretFehltMitDaten signalisiert den Fail-Safe-Abbruch: vorhandene Daten, aber
// nirgends ein Secret. run() gibt dafuer core.DiagnoseSecretFehltMitDaten aus statt
// der Sentinel-Meldung.
var errSecretFehltMitDaten = errors.New("start abgebrochen: keine Zugangsdaten zu vorhandenen Daten gefunden")

// materializeEnvFromVolume macht das jotti-config-Volume zur Quelle der Wahrheit
// fuers Install-Secret und schreibt den Host-.env-Spiegel fuer `compose --env-file`
// und das Relay. Nur nach ensureDocker aufrufen — der Volume-Read braucht einen
// laufenden Daemon. Suchreihenfolge: Volume, dann localDirs; ein adoptierter Treffer
// wird ins Volume geschrieben. Daten ohne Secret brechen den Start ab, statt sie
// auszusperren.
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

func readEnvCandidates(dirs []string) []string {
	candidates := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		data, _ := os.ReadFile(filepath.Join(dir, ".env"))
		candidates = append(candidates, string(data))
	}
	return candidates
}

// readConfigVolume liest die gespiegelte .env aus dem jotti-config-Volume. Leerer
// String = kein Secret vorhanden; nur ein echter Docker-Fehler wird durchgereicht.
func readConfigVolume() (string, error) {
	exists, err := volumeExists(configVolume)
	if err != nil {
		return "", err
	}
	if !exists {
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

// ensureConfigVolume legt das jotti-config-Volume — falls es fehlt — mit den
// Compose-Labels an, bevor zum ersten Mal hineingeschrieben wird. Ohne sie meldet
// Compose das per `docker run -v` angelegte Volume bei jedem Start als fremd
// ("volume ... not created by Docker Compose"). Kein `external: true`: das wuerde
// `down -v` verhindern und die Garantie brechen, dass das Secret die Daten nie
// ueberlebt. Idempotent — Labels sind nach dem Anlegen unveraenderlich.
func ensureConfigVolume() error {
	exists, err := volumeExists(configVolume)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	out, err := exec.Command("docker", "volume", "create",
		"--label", "com.docker.compose.project=jotti-local",
		"--label", "com.docker.compose.volume=jotti-config",
		configVolume).CombinedOutput()
	if err != nil {
		return fmt.Errorf("jotti-config-Volume anlegen fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func writeConfigVolume(content string) error {
	if err := ensureConfigVolume(); err != nil {
		return err
	}
	cmd := exec.Command("docker", "run", "--rm", "-i", "--entrypoint", "sh",
		"-v", configVolume+":/config", configHelperImage, "-c", "cat > "+configVolumePath)
	cmd.Stdin = strings.NewReader(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schreiben in den Datentresor fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// composeUp faehrt den Stack hoch und reicht die Ausgabe live durch. Der Proxy wird
// danach neu erzeugt, weil das Caddyfile nur im Entrypoint aus LAN_IP gerendert wird.
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

// runCompose benennt die .env-Quelle explizit per --env-file: nach der
// UAC-Elevation ist das Arbeitsverzeichnis System32, nicht das Projektverzeichnis.
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
