package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nicograef/jotti/windows/starter/core"
)

const (
	latestReleaseAPI = "https://api.github.com/repos/nicograef/jotti/releases/latest"
	releasesPage     = "https://github.com/nicograef/jotti/releases/latest"
	// updateCheckTimeout haelt den Online-Check kurz, damit ein Start auch offline
	// nicht spuerbar verzoegert wird.
	updateCheckTimeout = 3 * time.Second
)

// notifyIfUpdateAvailable weist nach einem gesunden Start auf eine neuere Version
// hin. Jeder Fehler wird still verschluckt — der Check darf den Start nie stoeren.
// Bewusst nur melden, nie automatisch aktualisieren (kein Auto-Update einer
// laufenden Kasse).
func notifyIfUpdateAvailable() {
	latest, err := fetchLatestRelease()
	if err != nil {
		return
	}
	if !core.IsNewerVersion(version, latest) {
		return
	}
	fmt.Println()
	fmt.Printf("Neue Version %s verfuegbar: %s\n", latest, releasesPage)
	fmt.Println("  Zum Aktualisieren: jotti beenden, neues ZIP herunterladen und jotti erneut starten.")
}

// fetchLatestRelease holt den tag_name des neuesten Releases. Der User-Agent ist
// Pflicht — GitHub lehnt Requests ohne ab.
func fetchLatestRelease() (string, error) {
	req, err := http.NewRequest(http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "jotti-starter/"+version)

	client := &http.Client{Timeout: updateCheckTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unerwarteter Status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return core.ParseLatestRelease(body)
}
