//go:build unit

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func anyContains(hits []string, substr string) bool {
	for _, h := range hits {
		if strings.Contains(h, substr) {
			return true
		}
	}
	return false
}

func TestCheckWindowsStrings_ReportsNormalAndRawStringLiterals(t *testing.T) {
	src := "package p\n\n" +
		"const Normal = \"café\"\n" + // normal string literal
		"const Raw = `weiß`\n" // raw string literal

	path := writeTemp(t, "strings.go", src)
	hits, err := checkWindowsStrings([]string{path})
	if err != nil {
		t.Fatalf("checkWindowsStrings: %v", err)
	}
	if !anyContains(hits, "café") {
		t.Errorf("expected a hit for the normal string literal, got %v", hits)
	}
	if !anyContains(hits, "weiß") {
		t.Errorf("expected a hit for the raw string literal, got %v", hits)
	}
	if len(hits) != 2 {
		t.Errorf("expected exactly 2 hits, got %d: %v", len(hits), hits)
	}
}

func TestCheckWindowsStrings_IgnoresComments(t *testing.T) {
	src := "package p\n\n" +
		"// äöü in a line comment must not be reported.\n" +
		"/* äöü in a block comment must not be reported. */\n" +
		"const OK = \"ascii only\"\n"

	path := writeTemp(t, "comments.go", src)
	hits, err := checkWindowsStrings([]string{path})
	if err != nil {
		t.Fatalf("checkWindowsStrings: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("expected no hits, comments are not string literals: %v", hits)
	}
}

func TestCheckBackendComments_ReportsStemInCommentOnly(t *testing.T) {
	src := "package p\n\n" +
		"// haelt das Ergebnis fest.\n" +
		"func Foo() {\n" +
		"\tfuer := \"fuer\" // identifier and string literal, not a comment word\n" +
		"\t_ = fuer\n" +
		"}\n"

	path := writeTemp(t, "comments.go", src)
	hits, err := checkBackendComments([]string{path}, []string{path})
	if err != nil {
		t.Fatalf("checkBackendComments: %v", err)
	}
	if !anyContains(hits, `"haelt"`) {
		t.Errorf("expected a hit for the comment word \"haelt\", got %v", hits)
	}
	if anyContains(hits, `"fuer"`) {
		t.Errorf("the local identifier and the string literal must not be flagged: %v", hits)
	}
	if len(hits) != 1 {
		t.Errorf("expected exactly 1 hit, got %d: %v", len(hits), hits)
	}
}

func TestCheckBackendComments_ProtectsOnlyDocHeaderName(t *testing.T) {
	// "Störung" opens its own doc comment (Go doc convention: a type's
	// comment starts with the type's exact name) and must be protected
	// there even when the file is checked without being one of its own
	// protection sources — the configuration the gate uses for this
	// package. The same word later in the very same sentence is ordinary
	// prose and must still be flagged.
	src := "package p\n\n" +
		"// Stoerung beschreibt eine Stoerung im System.\n" +
		"type Stoerung struct{}\n"

	path := writeTemp(t, "doc.go", src)
	hits, err := checkBackendComments([]string{path}, nil)
	if err != nil {
		t.Fatalf("checkBackendComments: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected exactly 1 hit, got %d: %v", len(hits), hits)
	}
	if !strings.Contains(hits[0], `"Stoerung"`) || !strings.Contains(hits[0], `"Störung"`) {
		t.Errorf("expected the second, non-header \"Stoerung\" to be flagged: %v", hits)
	}
}

func TestCheckBackendComments_ProtectsMidSentenceIdentifierReference(t *testing.T) {
	// A comment naming a function of another package, which this file
	// neither declares nor lists as a protection source: "Eroeffne" opens
	// the name with a stem, so hasInternalCapital is the only rule that
	// can keep it out of the hits.
	src := "package p\n\n" +
		"// Legt die Sitzung an, wie EroeffneKassensitzung es tut.\n" +
		"func Foo() {}\n"

	path := writeTemp(t, "ref.go", src)
	hits, err := checkBackendComments([]string{path}, nil)
	if err != nil {
		t.Fatalf("checkBackendComments: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("a multi-capital identifier reference must never be flagged: %v", hits)
	}
}

// TestCheckBackendComments_ProtectsCodeSpellings covers one comment per
// class of name a backend comment quotes. Every fixture is its own
// protection source unless the case says otherwise, exactly as the gate
// passes the backend's files.
func TestCheckBackendComments_ProtectsCodeSpellings(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		extraSrc  string
		noSources bool
		wantHits  []string
	}{
		{
			name: "route path in a handler doc comment",
			src: "package p\n\n" +
				"// POST /admin/get-tse-stoerungen liefert das Störungsprotokoll.\n" +
				"const route = \"/get-tse-stoerungen\"\n",
		},
		{
			name: "frozen event type",
			src: "package p\n\n" +
				"// Event kassensitzung-eroeffnet:v1 eröffnet die Sitzung.\n" +
				"const eventTyp = \"kassensitzung-eroeffnet:v1\"\n",
		},
		{
			name: "database column",
			src: "package p\n\n" +
				"// Spalte naechster_versuch_am trägt den Backoff.\n" +
				"const query = \"UPDATE tse_signaturauftraege SET naechster_versuch_am = $1\"\n",
		},
		{
			name: "enum value quoted in the comment, declared elsewhere",
			src: "package p\n\n" +
				"// Bonmodus \"pro_stueck\" und Kategorie \"getraenk\" kommen aus der Domäne.\n" +
				"func Foo() {}\n",
			noSources: true,
		},
		{
			name: "enum value unquoted, declared as a literal",
			src: "package p\n\n" +
				"// Kategorie getraenk zieht den Regelsteuersatz.\n" +
				"const kategorie = \"getraenk\"\n",
		},
		{
			name: "package doc header names its own package",
			src: "// Package dsfinvkpruefung prüft ein DSFinV-K-Export-ZIP.\n" +
				"package dsfinvkpruefung\n",
		},
		{
			name: "package doc header of a package that is not a protection source",
			src: "// Package pruefung liest ein Export-ZIP.\n" +
				"package pruefung\n",
			noSources: true,
		},
		{
			name: "package named in prose, declared in another file",
			src: "package p\n\n" +
				"// Wie in den Paketen kassenfuehrung und pruefung.\n" +
				"func Foo() {}\n",
			extraSrc: "package pruefung\n",
		},
		{
			name: "plain German prose is still reported",
			src: "package p\n\n" +
				"// fuer den naechsten Versuch.\n" +
				"func Foo() {}\n",
			wantHits: []string{`"fuer"`, `"naechsten"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTemp(t, "case.go", tt.src)
			sources := []string{path}
			if tt.noSources {
				sources = nil
			}
			if tt.extraSrc != "" {
				sources = append(sources, writeTemp(t, "other.go", tt.extraSrc))
			}
			hits, err := checkBackendComments([]string{path}, sources)
			if err != nil {
				t.Fatalf("checkBackendComments: %v", err)
			}
			if len(hits) != len(tt.wantHits) {
				t.Fatalf("expected %d hit(s), got %d: %v", len(tt.wantHits), len(hits), hits)
			}
			for _, want := range tt.wantHits {
				if !anyContains(hits, want) {
					t.Errorf("expected a hit for %s, got %v", want, hits)
				}
			}
		})
	}
}

func TestCheckBackendComments_IgnoresGoBuildDirective(t *testing.T) {
	src := "package p\n\n" +
		"//go:generate mockgen -destination=fuer_mock.go\n" +
		"// Ein Kommentar der ueber die Zeile hinausgeht.\n" +
		"func Foo() {}\n"

	path := writeTemp(t, "directive.go", src)
	hits, err := checkBackendComments([]string{path}, []string{path})
	if err != nil {
		t.Fatalf("checkBackendComments: %v", err)
	}
	if anyContains(hits, "fuer_mock") {
		t.Errorf("a //go: directive line must be skipped entirely: %v", hits)
	}
	if !anyContains(hits, `"ueber"`) {
		t.Errorf("the ordinary comment on the next line must still be checked: %v", hits)
	}
}

func TestCheckCmdASCII_HandlesCRLF(t *testing.T) {
	src := "@echo off\r\n" +
		"REM jotti starten — Doppelklick genuegt.\r\n" +
		"echo done\r\n"

	path := writeTemp(t, "start.cmd", src)
	hits, err := checkCmdASCII([]string{path})
	if err != nil {
		t.Fatalf("checkCmdASCII: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected exactly 1 hit on the CRLF file, got %d: %v", len(hits), hits)
	}
	if !strings.HasPrefix(hits[0], path+":2:") {
		t.Errorf("expected the hit on line 2, got %q", hits[0])
	}
	if strings.ContainsRune(hits[0], '\r') {
		t.Errorf("the reported line must not carry the trailing CR: %q", hits[0])
	}
}
