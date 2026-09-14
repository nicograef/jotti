package core

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// ParseLatestRelease zieht den tag_name aus der GitHub-Releases-Antwort. Ein leerer
// oder fehlender tag_name gilt als Fehler, damit der Aufrufer nicht auf einen leeren
// Versions-String hin "neue Version" meldet.
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

// IsNewerVersion meldet, ob latest hoeher als current ist (vMAJOR.MINOR.PATCH, "v"
// optional, Suffixe abgeschnitten). Ist eine Seite kein Semver — etwa der Dev-Build
// "dev" —, wird bewusst nichts gemeldet.
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

// IsDowngrade meldet, ob exeVersion streng aelter als dataVersion ist. Ohne Semver
// auf einer Seite ("dev", "latest", leer beim Erststart) greift die Sperre nicht —
// die Reihenfolge ist dann unbekannt. Spiegelt is_downgrade aus
// scripts/prod-update.sh.
func IsDowngrade(exeVersion, dataVersion string) bool {
	return IsNewerVersion(exeVersion, dataVersion)
}

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
