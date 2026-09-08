// Command checklanguage backs scripts/check-language.sh: it applies the two
// rules that need a real Go parser instead of a line-based grep — non-ASCII
// bytes inside Go string literals (not comments) and transliterated umlauts
// inside Go comment lines — plus the plain-text ASCII rule for
// packaging/**/*.cmd, which needs no parser at all.
//
// It is not part of any Go module (no go.mod above scripts/): `go run` builds
// it as a standalone "command-line-arguments" program, matching the
// architectural decision to add no new dependency for the AST gates.
//
// Usage: go run ./scripts/checklanguage <mode> <file>...
//
//	windows-strings  — report non-ASCII bytes in Go string literals
//	cmd-ascii        — report non-ASCII bytes anywhere in a .cmd file
//	backend-comments — report transliterated umlaut words in Go comment lines
//	backend-fix      — rewrite those words to real umlauts, in place
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// umlauts maps each transliterated form on the enforced word list to its
// correct German spelling. The mapping is per-word, not a generic character
// substitution: German orthography turns "ss" into
// "ß" only in some of these words (gemäß, schließen, größe), not others
// (müssen, genügt) — hardcoding each word keeps that correct.
var umlauts = map[string]string{
	"fuer":        "für",
	"ueber":       "über",
	"koennen":     "können",
	"muessen":     "müssen",
	"waehrend":    "während",
	"naechst":     "nächst",
	"auftraege":   "aufträge",
	"aenderung":   "änderung",
	"gemaess":     "gemäß",
	"zurueck":     "zurück",
	"moeglich":    "möglich",
	"spaeter":     "später",
	"aendern":     "ändern",
	"pruefen":     "prüfen",
	"laeuft":      "läuft",
	"haelt":       "hält",
	"groesse":     "größe",
	"schliessen":  "schließen",
	"genuegt":     "genügt",
	"einfuehrung": "einführung",
}

var wordPattern = regexp.MustCompile(`(?i)\b(` + strings.Join(wordKeys(), "|") + `)\b`)

func wordKeys() []string {
	keys := make([]string, 0, len(umlauts))
	for k := range umlauts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// replacement re-cases repl to match the casing of matched: fully upper stays
// fully upper, a capitalized word stays capitalized, everything else is
// left as the lowercase spelling from the umlauts map. Rune-aware because
// repl's first character can itself be a multi-byte umlaut (e.g. "über").
func replacement(matched, repl string) string {
	if matched == strings.ToUpper(matched) && matched != strings.ToLower(matched) {
		return strings.ToUpper(repl)
	}
	first := []rune(matched)[0]
	if unicode.IsUpper(first) {
		r := []rune(repl)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	return repl
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: checklanguage <windows-strings|cmd-ascii|backend-comments|backend-fix> <file>...")
		os.Exit(2)
	}
	mode := os.Args[1]
	files := os.Args[2:]

	var violations int
	var err error
	switch mode {
	case "windows-strings":
		violations, err = checkWindowsStrings(files)
	case "cmd-ascii":
		violations, err = checkCmdASCII(files)
	case "backend-comments":
		violations, err = checkBackendComments(files)
	case "backend-fix":
		violations, err = fixBackendComments(files)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", mode)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if violations > 0 {
		os.Exit(1)
	}
}

func hasNonASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7F {
			return true
		}
	}
	return false
}

// checkWindowsStrings reports non-ASCII bytes in Go string literals (BasicLit
// of kind STRING). Comments are never visited because ast.Inspect walks the
// syntax tree, not the token stream — they are excluded by construction.
func checkWindowsStrings(files []string) (int, error) {
	fset := token.NewFileSet()
	violations := 0
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return violations, fmt.Errorf("parse %s: %w", path, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			if hasNonASCII(lit.Value) {
				pos := fset.Position(lit.Pos())
				fmt.Printf("%s:%d: non-ASCII byte in Go string literal: %s\n", path, pos.Line, lit.Value)
				violations++
			}
			return true
		})
	}
	return violations, nil
}

// checkCmdASCII reports non-ASCII bytes anywhere in a .cmd file, including
// REM comment lines: Windows batch files carry no doc-comment convention
// that jotti's German prose could live in instead, and the printed jotti
// console output stays ASCII throughout, so the whole file stays ASCII.
func checkCmdASCII(files []string) (int, error) {
	violations := 0
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return violations, fmt.Errorf("read %s: %w", path, err)
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if hasNonASCII(line) {
				fmt.Printf("%s:%d: non-ASCII byte: %s\n", path, i+1, strings.TrimRight(line, "\r"))
				violations++
			}
		}
	}
	return violations, nil
}

// commentEdits returns, for one parsed file, the byte-range edits that
// convert transliterated umlaut words in comment text to their real
// spelling. Shared by the report-only and the in-place fix mode so both
// apply the exact same rule.
type edit struct {
	start, end int
	text       string
}

func commentEdits(fset *token.FileSet, f *ast.File) []edit {
	var edits []edit
	for _, group := range f.Comments {
		for _, c := range group.List {
			newText := wordPattern.ReplaceAllStringFunc(c.Text, func(match string) string {
				return replacement(match, umlauts[strings.ToLower(match)])
			})
			if newText != c.Text {
				start := fset.Position(c.Slash).Offset
				edits = append(edits, edit{start: start, end: start + len(c.Text), text: newText})
			}
		}
	}
	return edits
}

// checkBackendComments reports transliterated words on Go comment lines. A
// block comment's ast.Comment.Text carries embedded "\n"s, so the reported
// line is the comment's start line plus the newline count before the match.
func checkBackendComments(files []string) (int, error) {
	fset := token.NewFileSet()
	violations := 0
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return violations, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, group := range f.Comments {
			for _, c := range group.List {
				startLine := fset.Position(c.Slash).Line
				for i, line := range strings.Split(c.Text, "\n") {
					for _, match := range wordPattern.FindAllString(line, -1) {
						repl := replacement(match, umlauts[strings.ToLower(match)])
						fmt.Printf("%s:%d: transliterated umlaut %q (use %q) in comment\n", path, startLine+i, match, repl)
						violations++
					}
				}
			}
		}
	}
	return violations, nil
}

// fixBackendComments rewrites transliterated umlaut words in comments to
// their real spelling, applying edits back-to-front so earlier byte offsets
// stay valid. It never touches string literals or identifiers: edits are
// confined to the exact byte span of each *ast.Comment, taken from the
// parser's own position info.
func fixBackendComments(files []string) (int, error) {
	fset := token.NewFileSet()
	changed := 0
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return changed, fmt.Errorf("parse %s: %w", path, err)
		}
		edits := commentEdits(fset, f)
		if len(edits) == 0 {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return changed, fmt.Errorf("read %s: %w", path, err)
		}
		sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
		for _, e := range edits {
			content = append(content[:e.start], append([]byte(e.text), content[e.end:]...)...)
			changed++
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return changed, fmt.Errorf("write %s: %w", path, err)
		}
	}
	return changed, nil
}
