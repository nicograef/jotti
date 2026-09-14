// Package dsfinvkpruefung prüft ein DSFinV-K-Export-ZIP gegen die Struktur- und
// Inhaltsregeln der DSFinV-K 2.4 und liefert eine Befundliste.
//
// Bewusst unabhängig vom Erzeuger (backend/api/fiskal/dsfinvk) implementiert — eigener
// CSV- und index.xml-Parser, eigene DTD-Regeln —, damit ein Formatfehler im Erzeuger
// auch dann auffällt, wenn beide dieselbe Konstante teilten. Keine betragsmäßige
// Plausibilisierung; die leisten die Golden-File-Tests des Erzeugers.
//
// Referenz: DSFinV-K 2.4 (docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4).
package dsfinvkpruefung

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"sort"
)

// Befund ist ein einzelner Struktur- oder Inhaltsverstoß; Datei ist die betroffene
// Archivdatei und bleibt leer für paketweite Befunde.
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

// Pruefen prüft ein DSFinV-K-Export-ZIP (io.ReaderAt plus Größe) gegen die DSFinV-K 2.4.
// Eine leere Befundliste bedeutet konform; einen Fehler gibt es nur, wenn das ZIP selbst
// nicht lesbar ist.
func Pruefen(r io.ReaderAt, size int64) ([]Befund, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("zip nicht lesbar: %w", err)
	}
	return pruefeArchiv(zr)
}

// PruefenBytes prüft ein im Speicher liegendes Archiv (der Export erzeugt []byte).
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

// dateiInhalte liest alle regulären Archivdateien in eine Map Name -> Inhalt; ein
// Lesefehler wird als leerer Inhalt geführt, die nachgelagerten Regeln melden den Mangel.
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

// sortierteNamen sortiert die Dateinamen deterministisch (stabile Befundlisten).
func sortierteNamen(dateien map[string][]byte) []string {
	namen := make([]string, 0, len(dateien))
	for name := range dateien {
		namen = append(namen, name)
	}
	sort.Strings(namen)
	return namen
}
