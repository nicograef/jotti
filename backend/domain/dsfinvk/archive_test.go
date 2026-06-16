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
		"lines.csv", "lines_vat.csv", "location.csv", "payment.csv",
		"references.csv", "transactions.csv", "transactions_tse.csv", "transactions_vat.csv",
		"tse.csv", "vat.csv",
	}
	if !equalStrings(got, want) {
		t.Errorf("archive files = %v\nwant %v", got, want)
	}

	for _, absent := range []string{"slaves.csv", "pa.csv"} {
		if contains(got, absent) {
			t.Errorf("archive must not contain %s", absent)
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

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
