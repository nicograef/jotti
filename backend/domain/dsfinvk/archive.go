package dsfinvk

import (
	"archive/zip"
	"bytes"
	_ "embed"

	"github.com/nicograef/jotti/backend/domain/event"
)

// gdpduDTD ist die statische, zwingend mitzuliefernde GoBD/GDPdU-Beschreibungs-
// DTD, auf die die index.xml verweist.
//
//go:embed gdpdu-01-09-2004.dtd
var gdpduDTD []byte

// amtlicheIndexXML ist die amtliche, UNVERÄNDERTE index.xml der DSFinV-K v2.4
// (Quelle: docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4). Sie
// deklariert alle 20 Tabellen samt Spalten und Formaten (u. a. Dezimal-KOMMA);
// Prüfsoftware validiert gegen genau diese Datei. Die erzeugten CSVs müssen
// ihr exakt entsprechen — Spaltenabgleich per Test in index_test.go.
//
//go:embed index.xml
var amtlicheIndexXML []byte

const (
	// DTDFilename ist der unveränderliche Dateiname der GDPdU-DTD im Archiv.
	DTDFilename = "gdpdu-01-09-2004.dtd"
	// indexFilename ist der Beschreibungs-Index des Archivs.
	indexFilename = "index.xml"
)

// BuildArchive transformiert Snapshot und Events einer Kassensitzung in ein
// vollständiges DSFinV-K-ZIP: die CSV-Dateien, die beschreibende index.xml und
// die gdpdu-01-09-2004.dtd. Seiteneffektfrei — komponiert Mapper, CSV-Serializer,
// index.xml-Generator und ZIP-Packer. signaturen ist der Signatur-Stand je
// Event-ID aus der Signaturauftrags-Tabelle (die einzige Signaturquelle).
func BuildArchive(snapshot Snapshot, events []event.Event, signaturen map[int]EventSignatur) ([]byte, error) {
	archive, err := Map(snapshot, events, signaturen)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := writeZipFile(zw, indexFilename, amtlicheIndexXML); err != nil {
		return nil, err
	}
	if err := writeZipFile(zw, DTDFilename, gdpduDTD); err != nil {
		return nil, err
	}
	for _, t := range archive.Tables() {
		if err := writeZipFile(zw, t.File, serializeCSV(t)); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeZipFile(zw *zip.Writer, name string, content []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}
