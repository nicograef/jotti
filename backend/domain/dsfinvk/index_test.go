//go:build unit

package dsfinvk

import (
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/event"
)

func TestBuildIndexXMLDescriptor(t *testing.T) {
	tables := []Table{
		{
			File:        "transactions.csv",
			LogicalName: "Bonkopf",
			Description: "Ein Datensatz je Kassenbon",
			Columns:     []column{alpha("Z_KASSE_ID"), num("Z_NR", 0), num("UMS_BRUTTO", 2)},
		},
	}

	xml := string(buildIndexXML(dataSupplier{name: "TSV Beispiel", location: "Musterdorf", comment: "jotti DSFinV-K-Export"}, tables))

	mustContain := []string{
		`<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">`,
		"<Version>2.4</Version>",
		"<Name>TSV Beispiel</Name>",
		"<URL>transactions.csv</URL>",
		"<Name>Bonkopf</Name>",
		"<DecimalSymbol>.</DecimalSymbol>",
		"<DigitGroupingSymbol>,</DigitGroupingSymbol>",
		"<ColumnDelimiter>;</ColumnDelimiter>",
		"<Range><From>2</From></Range>",
		"<Name>Z_KASSE_ID</Name>\n          <AlphaNumeric/>",
		"<Name>UMS_BRUTTO</Name>\n          <Numeric><Accuracy>2</Accuracy></Numeric>",
	}
	for _, want := range mustContain {
		if !strings.Contains(xml, want) {
			t.Errorf("index.xml missing %q\n---\n%s", want, xml)
		}
	}
}

// TestBuildIndexXMLDeclaresExactArchiveTables stellt sicher, dass der Descriptor
// genau die im Archiv vorhandenen Tabellen deklariert und die für jotti
// gegenstandslosen Tabellen weglässt.
func TestBuildIndexXMLDeclaresExactArchiveTables(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)}, nil)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	xml := string(buildIndexXML(dataSupplier{name: "TSV"}, archive.Tables()))

	for _, tbl := range archive.Tables() {
		if strings.Count(xml, "<URL>"+tbl.File+"</URL>") != 1 {
			t.Errorf("index.xml must declare %s exactly once", tbl.File)
		}
	}
	for _, absent := range []string{"slaves.csv", "pa.csv", "itemamounts.csv", "subitems.csv"} {
		if strings.Contains(xml, absent) {
			t.Errorf("index.xml must not declare %s", absent)
		}
	}
}
