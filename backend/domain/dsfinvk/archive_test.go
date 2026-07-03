//go:build unit

package dsfinvk

import (
	"archive/zip"
	"bytes"
	"sort"
	"testing"

	"github.com/nicograef/jotti/backend/domain/event"
)

func TestBuildArchiveContents(t *testing.T) {
	zipBytes, err := BuildArchive(testSnapshot(), []event.Event{barverkaufEvent(t)}, nil)
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}

	got := make([]string, 0, len(reader.File))
	for _, f := range reader.File {
		got = append(got, f.Name)
	}
	sort.Strings(got)

	want := []string{
		"allocation_groups.csv", "businesscases.csv", "cash_per_currency.csv", "cashpointclosing.csv",
		"cashregister.csv", "datapayment.csv", "gdpdu-01-09-2004.dtd", "index.xml",
		"itemamounts.csv", "lines.csv", "lines_vat.csv", "location.csv",
		"pa.csv", "payment.csv", "references.csv", "slaves.csv",
		"subitems.csv", "transactions.csv", "transactions_tse.csv", "transactions_vat.csv",
		"tse.csv", "vat.csv",
	}
	if !equalStrings(got, want) {
		t.Errorf("archive files = %v\nwant %v", got, want)
	}

	// Die index.xml im Archiv ist byte-identisch mit der amtlichen Vorlage.
	for _, f := range reader.File {
		if f.Name != "index.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open index.xml: %v", err)
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			t.Fatalf("read index.xml: %v", err)
		}
		rc.Close()
		if !bytes.Equal(buf.Bytes(), amtlicheIndexXML) {
			t.Error("index.xml im Archiv weicht von der amtlichen Vorlage ab")
		}
	}
}

func TestBuildArchiveEmptySession(t *testing.T) {
	_, err := BuildArchive(testSnapshot(), nil, nil)
	if err != ErrKeineVorgaenge {
		t.Fatalf("BuildArchive() error = %v, want ErrKeineVorgaenge", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
