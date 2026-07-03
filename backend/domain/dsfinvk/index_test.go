//go:build unit

package dsfinvk

import (
	"encoding/xml"
	"testing"

	"github.com/nicograef/jotti/backend/domain/event"
)

// amtlicheTabelle bildet die Table-Deklaration der amtlichen index.xml ab.
type amtlicheTabelle struct {
	URL            string `xml:"URL"`
	VariableLength struct {
		Columns []struct {
			Name string `xml:"Name"`
		} `xml:"VariableColumn"`
	} `xml:"VariableLength"`
}

type amtlicherIndex struct {
	Media struct {
		Tables []amtlicheTabelle `xml:"Table"`
	} `xml:"Media"`
}

// Die ausgelieferte index.xml ist die amtliche, unveränderte Vorlage der
// DSFinV-K v2.4. Die erzeugten Tabellen müssen ihr exakt entsprechen: jede
// deklarierte Datei existiert, in deklarierter Reihenfolge, mit identischen
// Spaltennamen in identischer Reihenfolge.
func TestArchivEntsprichtAmtlicherIndexXML(t *testing.T) {
	var amtlich amtlicherIndex
	if err := xml.Unmarshal(amtlicheIndexXML, &amtlich); err != nil {
		t.Fatalf("amtliche index.xml nicht parsebar: %v", err)
	}
	if len(amtlich.Media.Tables) != 20 {
		t.Fatalf("amtliche index.xml deklariert %d Tabellen, erwartet 20", len(amtlich.Media.Tables))
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
			if tbl.Columns[c].name != declCol.Name {
				t.Errorf("%s Spalte %d: %q, amtlich deklariert %q", tbl.File, c, tbl.Columns[c].name, declCol.Name)
			}
		}
	}
}

// Die amtliche index.xml deklariert das Komma als Dezimalsymbol — alle
// Betrags-, Mengen- und Prozentformate müssen dem entsprechen.
func TestZahlenformateNutzenKommaAlsDezimalsymbol(t *testing.T) {
	if got := formatAmount(-150); got != "-1,50" {
		t.Errorf("formatAmount(-150) = %q, want -1,50", got)
	}
	if got := formatQuantity(2); got != "2,000" {
		t.Errorf("formatQuantity(2) = %q, want 2,000", got)
	}
	if got := formatPercent(19); got != "19,00" {
		t.Errorf("formatPercent(19) = %q, want 19,00", got)
	}
}
