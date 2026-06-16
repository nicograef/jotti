//go:build unit

package dsfinvk

import "testing"

func TestSerializeCSVFormat(t *testing.T) {
	table := Table{
		Columns: []column{alpha("TEXT"), num("BETRAG", 2), alpha("MENGE")},
		Records: [][]string{
			{"plain", "5.00", "ok"},
			{"semi;colon", "1.50", `inner"quote`},
			{"line\nbreak", "0.00", "ende"},
		},
	}

	got := string(serializeCSV(table))
	want := "TEXT;BETRAG;MENGE\r\n" +
		"plain;5.00;ok\r\n" +
		`"semi;colon";1.50;"inner""quote"` + "\r\n" +
		"\"line\nbreak\";0.00;ende\r\n"

	if got != want {
		t.Fatalf("serializeCSV() =\n%q\nwant\n%q", got, want)
	}
}

func TestEscapeCSVField(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"plain", "plain"},
		{"with space", "with space"},
		{"semi;col", `"semi;col"`},
		{`quote"d`, `"quote""d"`},
		{"line\nbreak", "\"line\nbreak\""},
		{"carriage\rreturn", "\"carriage\rreturn\""},
	}

	for _, tt := range tests {
		if got := escapeCSVField(tt.in); got != tt.want {
			t.Errorf("escapeCSVField(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
