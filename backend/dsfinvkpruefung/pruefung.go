// Package dsfinvkpruefung prüft ein DSFinV-K-Export-ZIP eigenständig gegen die
// Strukturregeln der DSFinV-K 2.4 und liefert eine Befundliste. Es ist der
// Gegenspieler des Erzeugers (backend/api/fiskal/dsfinvk): der Erzeuger baut das
// Archiv, diese Prüfung liest es zurück und stellt sicher, dass Paket, Dateinamen,
// CSV-Format und index.xml/DTD-Struktur der Spezifikation entsprechen.
//
// Die Prüfung ist bewusst unabhängig vom Erzeuger implementiert (eigener CSV- und
// index.xml-Parser, eigene DTD-Regeln), damit ein Formatfehler im Erzeuger auch
// dann auffällt, wenn beide dieselbe Konstante teilten. Sie führt keine
// fachliche/betragsmäßige Plausibilisierung durch — das leisten die Golden-File-
// Tests des Erzeugers —, sondern ausschließlich die formale Paket- und
// Strukturkonformität.
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

// Befund ist ein einzelner Strukturverstoß. Datei ist die betroffene Archivdatei
// (leer für paketweite Befunde), Regel benennt die verletzte Strukturregel und
// Meldung beschreibt den konkreten Verstoß.
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

// Pruefen liest das DSFinV-K-Export-ZIP (io.ReaderAt plus Größe) und prüft es
// gegen die Strukturregeln der DSFinV-K 2.4. Rückgabe ist die Befundliste; ein
// befundfreies (leeres) Ergebnis bedeutet: strukturell konform. Ein Fehler wird
// nur zurückgegeben, wenn das ZIP selbst nicht lesbar ist (kein gültiges Archiv).
func Pruefen(r io.ReaderAt, size int64) ([]Befund, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("zip nicht lesbar: %w", err)
	}
	return pruefeArchiv(zr)
}

// PruefenBytes ist der bequeme Einstieg für ein vollständig im Speicher liegendes
// Archiv (der Regelfall in jotti: der Export erzeugt []byte). Delegiert an Pruefen.
func PruefenBytes(archiv []byte) ([]Befund, error) {
	return Pruefen(bytes.NewReader(archiv), int64(len(archiv)))
}

// pruefeArchiv wendet alle Strukturregeln auf ein bereits geöffnetes Archiv an
// und sammelt die Befunde. Reihenfolge: Paket (vorhandene Dateien, Dateinamen),
// dann index.xml/DTD, dann je Tabelle die CSV-Struktur.
func pruefeArchiv(zr *zip.Reader) ([]Befund, error) {
	dateien := dateiInhalte(zr)

	var befunde []Befund
	befunde = append(befunde, pruefeDateinamen(dateien)...)
	befunde = append(befunde, pruefePaketpflichtdateien(dateien)...)

	// index.xml ist die deklarative Beschreibung; ihre DTD-Struktur bestimmt, welche
	// Tabellen mit welchen Spalten das Archiv enthält. Ohne lesbare index.xml sind
	// die CSV-Prüfungen gegenstandslos.
	tabellen, indexBefunde := pruefeIndexXML(dateien[indexDatei])
	befunde = append(befunde, indexBefunde...)

	befunde = append(befunde, pruefeTabellenGegenIndex(dateien, tabellen)...)

	return befunde, nil
}

// dateiInhalte liest alle regulären Archivdateien in eine Map Name -> Inhalt.
// Verzeichniseinträge werden übersprungen. Ein Lesefehler einer einzelnen Datei
// wird als leerer Inhalt geführt; die nachgelagerten Regeln melden dann den
// strukturellen Mangel.
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
		data, _ := io.ReadAll(rc)
		_ = rc.Close()
		out[f.Name] = data
	}
	return out
}

// sortierteNamen liefert die Dateinamen einer Map in deterministischer Reihenfolge
// (für stabile, reproduzierbare Befundlisten).
func sortierteNamen(dateien map[string][]byte) []string {
	namen := make([]string, 0, len(dateien))
	for name := range dateien {
		namen = append(namen, name)
	}
	sort.Strings(namen)
	return namen
}
