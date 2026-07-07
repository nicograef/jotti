package core

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// ParseLatestRelease zieht den tag_name aus der Antwort der GitHub-Releases-API
// (/releases/latest). Nur dieses Feld interessiert; alles andere wird ignoriert.
// Ein leerer oder fehlender tag_name gilt als Fehler, damit der Aufrufer nicht auf
// einen leeren Versions-String hin "neue Version" meldet.
func ParseLatestRelease(data []byte) (string, error) {
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return "", err
	}
	if strings.TrimSpace(release.TagName) == "" {
		return "", errors.New("kein tag_name im Release")
	}
	return release.TagName, nil
}

// IsNewerVersion meldet, ob latest eine echte, hoehere Release-Version als current
// ist. Beide werden als vMAJOR.MINOR.PATCH gelesen (fuehrendes "v" optional,
// Vorabversions-/Build-Suffixe werden abgeschnitten). Laesst sich eine Seite nicht
// als Semver lesen — etwa der Dev-Build "dev" oder "dev-<sha>" —, wird bewusst
// nichts gemeldet: Entwickler-Builds sollen keinen Update-Hinweis ausloesen.
func IsNewerVersion(current, latest string) bool {
	c, okc := parseSemver(current)
	l, okl := parseSemver(latest)
	if !okc || !okl {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// IsDowngrade meldet, ob exeVersion eine streng aeltere Version als dataVersion
// ist. Schlaegt die Semver-Aufloesung einer Seite fehl (z. B. "dev", "latest",
// leerer String beim Erststart), gilt die Downgrade-Sperre nicht — ohne
// gepinntes Semver ist die Reihenfolge unbekannt. Spiegelt die
// is_downgrade-Logik aus scripts/prod-update.sh.
func IsDowngrade(exeVersion, dataVersion string) bool {
	e, oke := parseSemver(exeVersion)
	d, okd := parseSemver(dataVersion)
	if !oke || !okd {
		return false
	}
	for i := 0; i < 3; i++ {
		if e[i] != d[i] {
			return e[i] < d[i]
		}
	}
	return false // gleich ist kein Downgrade
}

// parseSemver liest "v1.2.3" (oder "1.2.3") in [major, minor, patch]. Ein
// Vorabversions- oder Build-Suffix ("1.2.3-rc1", "1.2.3+meta") wird vor dem Parsen
// abgeschnitten. Fehlt eine der drei Komponenten oder ist sie nicht numerisch,
// schlaegt das Parsen fehl (ok == false).
func parseSemver(s string) ([3]int, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var v [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		v[i] = n
	}
	return v, true
}
