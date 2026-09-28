package dsfinvk

import (
	"encoding/xml"
	"testing"

	"github.com/nicograef/jotti/backend/domain/event"
)

// amtlicheTabelle mirrors a Table declaration of the official index.xml.
type amtlicheTabelle struct {
	URL            string `xml:"URL"`
	VariableLength struct {
		Columns []struct {
			Name      string `xml:"Name"`
			MaxLength int    `xml:"MaxLength"`
		} `xml:"VariableColumn"`
	} `xml:"VariableLength"`
}

type amtlicherIndex struct {
	Media struct {
		Tables []amtlicheTabelle `xml:"Table"`
	} `xml:"Media"`
}

// The generated tables must match the unmodified official v2.4 index.xml exactly:
// every declared file, in declared order, with identical columns in identical order.
func TestArchivEntsprichtAmtlicherIndexXML(t *testing.T) {
	var amtlich amtlicherIndex
	if err := xml.Unmarshal(amtlicheIndexXML, &amtlich); err != nil {
		t.Fatalf("amtliche index.xml nicht parsebar: %v", err)
	}
	if len(amtlich.Media.Tables) != 20 {
		t.Errorf("amtliche index.xml deklariert %d Tabellen, erwartet 20", len(amtlich.Media.Tables))
	}

	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)}, nil)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	tables := archive.Tables()

	if len(tables) != len(amtlich.Media.Tables) {
		t.Fatalf("Archiv enthält %d Tabellen, amtlich deklariert sind %d", len(tables), len(amtlich.Media.Tables))
	}

	for i, decl := range amtlich.Media.Tables {
		tbl := tables[i]
		if tbl.File != decl.URL {
			t.Errorf("Tabelle %d: Datei %q, amtlich deklariert %q", i, tbl.File, decl.URL)
			continue
		}
		if len(tbl.Columns) != len(decl.VariableLength.Columns) {
			t.Errorf("%s: %d Spalten, amtlich deklariert %d", tbl.File, len(tbl.Columns), len(decl.VariableLength.Columns))
			continue
		}
		for c, declCol := range decl.VariableLength.Columns {
			if tbl.Columns[c] != declCol.Name {
				t.Errorf("%s Spalte %d: %q, amtlich deklariert %q", tbl.File, c, tbl.Columns[c], declCol.Name)
			}
		}
	}
}

// The official index.xml declares the comma as decimal symbol; all amount, quantity
// and percent formats must match it.
func TestZahlenformateNutzenKommaAlsDezimalsymbol(t *testing.T) {
	if got := formatAmount(-150); got != "-1,50" {
		t.Errorf("formatAmount(-150) = %q, want -1,50", got)
	}
	if got := formatQuantity(2); got != "2,000" {
		t.Errorf("formatQuantity(2) = %q, want 2,000", got)
	}
}

// amtlicheMaxLength reads a column's MaxLength from the embedded official index.xml,
// the length the mapper truncates to.
func amtlicheMaxLength(t *testing.T, datei string, spalte string) int {
	t.Helper()

	var amtlich amtlicherIndex
	if err := xml.Unmarshal(amtlicheIndexXML, &amtlich); err != nil {
		t.Fatalf("amtliche index.xml nicht parsebar: %v", err)
	}

	for _, tbl := range amtlich.Media.Tables {
		if tbl.URL != datei {
			continue
		}
		for _, deklariert := range tbl.VariableLength.Columns {
			if deklariert.Name != spalte {
				continue
			}
			if deklariert.MaxLength <= 0 {
				t.Fatalf("Spalte %s/%s deklariert keine MaxLength", datei, spalte)
			}
			return deklariert.MaxLength
		}
	}

	t.Fatalf("Spalte %s/%s fehlt in der amtlichen index.xml", datei, spalte)
	return 0
}
