// Package layout kennt die Ablage der Einzeldatei: Programmordner je Version und Datenverzeichnis.
package layout

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Paths sind die Pfade einer Version im Programmordner und im Datenverzeichnis.
type Paths struct {
	Program     string // %ProgramFiles%\jotti\<version>
	PostgresBin string // <Program>\postgres\bin
	App         string // <Program>\app
	Data        string // %PROGRAMDATA%\jotti
	Env         string // <Data>\.env
	LastVersion string // <Data>\last-version
	PGData      string // <Data>\pgdata
	Caddy       string // <Data>\caddy
}

// Dateinamen im Datenverzeichnis, die auch der Docker-Starter schreibt.
const (
	lastVersionFile = "last-version"
	pgdataDir       = "pgdata"
)

// Paths leitet alle Pfade aus %ProgramFiles%, %PROGRAMDATA% und der Version ab.
func NewPaths(programFiles, programData, version string) Paths {
	program := filepath.Join(programFiles, "jotti", version)
	data := filepath.Join(programData, "jotti")
	return Paths{
		Program:     program,
		PostgresBin: filepath.Join(program, "postgres", "bin"),
		App:         filepath.Join(program, "app"),
		Data:        data,
		Env:         filepath.Join(data, ".env"),
		LastVersion: filepath.Join(data, lastVersionFile),
		PGData:      filepath.Join(data, pgdataDir),
		Caddy:       filepath.Join(data, "caddy"),
	}
}

// DockerInstallation meldet eine Docker-Installation im Datenverzeichnis: Der Starter
// schreibt last-version erst nach einem gesunden Start, die Einzeldatei hält ihre Daten
// in pgdata. last-version ohne pgdata gehört also zu Daten im Docker-Volume.
func DockerInstallation(dataDir string) (bool, error) {
	lastVersion, err := exists(filepath.Join(dataDir, lastVersionFile))
	if err != nil {
		return false, err
	}
	pgdata, err := exists(filepath.Join(dataDir, pgdataDir))
	if err != nil {
		return false, err
	}
	return lastVersion && !pgdata, nil
}

// extractedMarker liegt im Programmordner, sobald Extract ihn vollständig befüllt hat.
const extractedMarker = ".entpackt"

// Extract entpackt postgres.zip nach <dst>\postgres und kopiert app/ nach <dst>\app.
// Ein vollständig entpackter Ordner trägt die Marker-Datei; ein zweiter Aufruf ändert dann nichts.
func Extract(src fs.FS, dst string) error {
	marker := filepath.Join(dst, extractedMarker)
	done, err := exists(marker)
	if err != nil || done {
		return err
	}
	// Ein abgebrochener Lauf hinterlässt halbe Ordner ohne Marker.
	for _, dir := range []string{"postgres", "app"} {
		if err := os.RemoveAll(filepath.Join(dst, dir)); err != nil {
			return err
		}
	}

	zipData, err := fs.ReadFile(src, "postgres.zip")
	if err != nil {
		return err
	}
	if err := unzip(zipData, filepath.Join(dst, "postgres")); err != nil {
		return fmt.Errorf("postgres.zip: %w", err)
	}
	app, err := fs.Sub(src, "app")
	if err != nil {
		return err
	}
	if err := os.CopyFS(filepath.Join(dst, "app"), app); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	return os.WriteFile(marker, nil, 0o600)
}

func unzip(data []byte, dst string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		name := path.Clean(f.Name)
		if !fs.ValidPath(name) || strings.Contains(name, `\`) {
			return fmt.Errorf("ungueltiger Pfad im Archiv: %q", f.Name)
		}
		target := filepath.Join(dst, filepath.FromSlash(name))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o750); err != nil {
				return err
			}
			continue
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return err
	}
	in, err := f.Open()
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck // read-only
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil { //nolint:gosec // G110: das Archiv ist eingebettet und gepinnt
		_ = out.Close()
		return err
	}
	return out.Close()
}

func exists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
