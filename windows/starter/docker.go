package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nicograef/jotti/windows/starter/core"
)

const dockerCliPath = `C:\Program Files\Docker\Docker\DockerCli.exe`

// Grosszuegig fuer den WSL2-/VM-Kaltstart auf Altgeraeten.
const dockerStartTimeout = 120 * time.Second

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
	return exec.Command(dockerCliPath, "-SwitchLinuxEngine").Run()
}
