package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nicograef/jotti/windows/starter/core"
)

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
