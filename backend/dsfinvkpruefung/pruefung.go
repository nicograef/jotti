// Package dsfinvkpruefung checks a DSFinV-K 2.4 export ZIP against structure and content rules.
// It parses independently of the generator (api/fiskal/dsfinvk) so a shared mistake still
// surfaces; see docs/compliance.md §6.1.
package dsfinvkpruefung

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"sort"
)

// Befund is one structure or content violation; Datei stays empty for package-wide findings.
type Befund struct {
	Datei   string
	Regel   string
	Meldung string
}

func (b Befund) String() string {
	if b.Datei == "" {
		return fmt.Sprintf("[%s] %s", b.Regel, b.Meldung)
	}
	return fmt.Sprintf("[%s] %s: %s", b.Regel, b.Datei, b.Meldung)
}

// Pruefen returns no findings for a conforming archive; an error means the ZIP itself is unreadable.
func Pruefen(r io.ReaderAt, size int64) ([]Befund, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("zip nicht lesbar: %w", err)
	}
	return pruefeArchiv(zr)
}

// PruefenBytes checks an in-memory archive, the form the export produces.
func PruefenBytes(archiv []byte) ([]Befund, error) {
	return Pruefen(bytes.NewReader(archiv), int64(len(archiv)))
}

func pruefeArchiv(zr *zip.Reader) ([]Befund, error) {
	dateien := dateiInhalte(zr)

	var befunde []Befund
	befunde = append(befunde, pruefeDateinamen(dateien)...)
	befunde = append(befunde, pruefePaketpflichtdateien(dateien)...)

	tabellen, indexBefunde := pruefeIndexXML(dateien[indexDatei])
	befunde = append(befunde, indexBefunde...)

	befunde = append(befunde, pruefeTabellenGegenIndex(dateien, tabellen)...)

	befunde = append(befunde, pruefeInhalt(dateien, tabellen)...)

	return befunde, nil
}

// dateiInhalte maps an unreadable file to nil content so the later rules report it.
func dateiInhalte(zr *zip.Reader) map[string][]byte {
	out := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			out[f.Name] = nil
			continue
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			out[f.Name] = nil
			continue
		}
		out[f.Name] = data
	}
	return out
}

// sortierteNamen sorts the file names so findings come in a stable order.
func sortierteNamen(dateien map[string][]byte) []string {
	namen := make([]string, 0, len(dateien))
	for name := range dateien {
		namen = append(namen, name)
	}
	sort.Strings(namen)
	return namen
}
