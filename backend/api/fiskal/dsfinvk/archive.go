package dsfinvk

import (
	"archive/zip"
	"bytes"
	_ "embed"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// gdpduDTD is the mandatory GDPdU DTD that index.xml references.
//
//go:embed gdpdu-01-09-2004.dtd
var gdpduDTD []byte

// amtlicheIndexXML is the unmodified official DSFinV-K v2.4 index.xml audit software validates against.
//
//go:embed index.xml
var amtlicheIndexXML []byte

const (
	DTDFilename   = "gdpdu-01-09-2004.dtd"
	indexFilename = "index.xml"
)

// BuildArchive zips a Kassensitzung's CSVs with the embedded index.xml and DTD.
// signaturen is the per-event state from the Signaturauftrag table, the only signature source.
func BuildArchive(snapshot Snapshot, events []event.Event, signaturen map[int]tse.EventSignatur) ([]byte, error) {
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
