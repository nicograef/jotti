// Command checklanguage implements the three rules behind
// scripts/check-language.sh's language-and-character gate:
//
//   - windows-strings: every Go string literal (not a comment) under
//     windows/** must be pure ASCII — the starter and relay print these
//     straight to a Windows console, which mangles non-ASCII bytes.
//   - cmd-ascii: a packaging/**/*.cmd file must be pure ASCII throughout,
//     comments included — a Windows batch file has no separate
//     doc-comment channel the way Go does.
//   - backend-comments: a Go comment line under backend/** must not carry
//     a German word stem spelled with the two-letter ASCII stand-in for
//     ä/ö/ü (see the stems map below for the exact list) where a real
//     umlaut belongs. It never touches string literals.
//
// The backend-comments rule protects every spelling the code itself uses,
// because a comment quoting a route, an event type, a table or column, an
// enum value or a package name must repeat that name byte for byte —
// identifiers never carry umlauts, and they never change. A match is
// therefore not reported when it is
//
//   - a word in the protected set: the name of any package or identifier
//     the backend declares or references, or the whole text of any Go
//     string literal there — read from the protection sources the caller
//     passes after "--";
//   - part of a token held together by "/", "_", "-", ":" or "." — a route
//     path, a column name, an event type, a qualified name;
//   - inside double quotes or backticks in the comment;
//   - the name after "Package" in a package doc header;
//   - carrying an upper-case letter past its first rune, which German
//     spelling never does but a Go identifier routinely does;
//   - a doc comment's own opening word, naming the declaration below it.
//
// Exit code 0 means no violations, 1 means violations were found and
// printed, 2 means the tool itself failed (bad arguments, a file that
// doesn't parse, an I/O error).
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

// stems maps each ASCII-transliterated German word stem to its correctly
// spelled form. Matching is by stem, not whole word (see stemPattern), so
// one entry catches every inflected ending of that stem. A prefixed
// compound needs its own separate entry: the match is anchored to a
// word's start (\b), and German prefixes attach directly with no
// separator, so the bare stem does not match once a prefix is glued on
// the front. The map is per-stem, not a blanket character substitution,
// because German orthography turns a transliterated double-s into "ß" in
// some of these words (schließen, gemäß, größe, mäßig) and keeps it a
// plain double-s in others (müssen) — that distinction exists per word,
// not per character.
var stems = map[string]string{
	"fuer":              "für",
	"ueber":             "über",
	"koenn":             "könn",
	"muess":             "müss",
	"waehrend":          "während",
	"naechst":           "nächst",
	"auftraeg":          "aufträg",
	"aender":            "änder",
	"gemaess":           "gemäß",
	"zurueck":           "zurück",
	"moeglich":          "möglich",
	"spaet":             "spät",
	"pruef":             "prüf",
	"laeuf":             "läuf",
	"haelt":             "hält",
	"groess":            "größ",
	"schliess":          "schließ",
	"genueg":            "genüg",
	"einfuehr":          "einführ",
	"uebernahm":         "übernahm",
	"stoerung":          "störung",
	"laess":             "läss",
	"rueck":             "rück",
	"traeg":             "träg",
	"endgueltig":        "endgültig",
	"getaetigt":         "getätigt",
	"oeffn":             "öffn",
	"geoeffn":           "geöffn",
	"eroeffn":           "eröffn",
	"fuenf":             "fünf",
	"zaehl":             "zähl",
	"haeng":             "häng",
	"faeng":             "fäng",
	"waer":              "wär",
	"fuehr":             "führ",
	"gefuehr":           "geführ",
	"durchgefuehr":      "durchgeführ",
	"auszufuehr":        "auszuführ",
	"herausfuehr":       "herausführ",
	"ueberfuehr":        "überführ",
	"abhaeng":           "abhäng",
	"unabhaeng":         "unabhäng",
	"datenabhaeng":      "datenabhäng",
	"abzuegl":           "abzügl",
	"aeltest":           "ältest",
	"atomaritaet":       "atomarität",
	"aufgeloes":         "aufgelös",
	"aufgeschluess":     "aufgeschlüss",
	"raeum":             "räum",
	"ausgeloes":         "ausgelös",
	"ausloes":           "auslös",
	"loes":              "lös",
	"buendel":           "bündel",
	"buendig":           "bündig",
	"traef":             "träf",
	"betraef":           "beträf",
	"braech":            "bräch",
	"faehig":            "fähig",
	"faehr":             "fähr",
	"faell":             "fäll",
	"faelsch":           "fälsch",
	"frueh":             "früh",
	"fuell":             "füll",
	"erfuell":           "erfüll",
	"befuell":           "befüll",
	"fuett":             "fütt",
	"hoer":              "hör",
	"gehoer":            "gehör",
	"zuhoer":            "zuhör",
	"kuerz":             "kürz",
	"gekuerz":           "gekürz",
	"geschaeft":         "geschäft",
	"getraenk":          "getränk",
	"waehl":             "wähl",
	"gewaehl":           "gewähl",
	"gewaenn":           "gewänn",
	"gewoehn":           "gewöhn",
	"staend":            "ständ",
	"verstaend":         "verständ",
	"vollstaend":        "vollständ",
	"unvollstaend":      "unvollständ",
	"gleichstaend":      "gleichständ",
	"gueltig":           "gültig",
	"ungueltig":         "ungültig",
	"duerf":             "dürf",
	"empfaeng":          "empfäng",
	"ergaeb":            "ergäb",
	"klaer":             "klär",
	"erklaer":           "erklär",
	"maessig":           "mäßig",
	"ermaessig":         "ermäßig",
	"kraeft":            "kräft",
	"noetig":            "nötig",
	"unnoetig":          "unnötig",
	"oberflaech":        "oberfläch",
	"raeng":             "räng",
	"regulaer":          "regulär",
	"saeh":              "säh",
	"saetz":             "sätz",
	"steuersaetz":       "steuersätz",
	"schaedlich":        "schädlich",
	"unschaedlich":      "unschädlich",
	"schlaeg":           "schläg",
	"schlueg":           "schlüg",
	"schluessel":        "schlüssel",
	"schoeb":            "schöb",
	"verschoeb":         "verschöb",
	"sekundaer":         "sekundär",
	"stuend":            "stünd",
	"mehrstuend":        "mehrstünd",
	"entstuend":         "entstünd",
	"stuetz":            "stütz",
	"unterstuetz":       "unterstütz",
	"stueck":            "stück",
	"tatsaechlich":      "tatsächlich",
	"laeng":             "läng",
	"verlaeng":          "verläng",
	"vorgaeng":          "vorgäng",
	"waechs":            "wächs",
	"waecht":            "wächt",
	"wuerd":             "würd",
	"vertrauenswuerdig": "vertrauenswürdig",
	"beruehr":           "berühr",
	"unberuehr":         "unberühr",
	"beschraenk":        "beschränk",
	"unbeschraenk":      "unbeschränk",
	"grosszuegig":       "großzügig",
	"hoeh":              "höh",
	"hoechst":           "höchst",
	"kaem":              "käm",
	"jueng":             "jüng",
	"inaktivitaet":      "inaktivität",
	"kapazitaet":        "kapazität",
	"uebrig":            "übrig",
	"umhuell":           "umhüll",
	"stoess":            "stöß",
	"lueck":             "lück",
	"schaerf":           "schärf",
}

// stemPattern matches a known stem at a word's start, case-insensitively,
// followed by whatever inflection continues the word. Longer stems are
// tried first (see stemKeys), so a more specific prefixed-compound entry
// wins over a shorter stem that would otherwise also match at the same
// starting position.
var stemPattern = regexp.MustCompile(`(?i)\b(` + strings.Join(stemKeys(), "|") + `)(\w*)`)

func stemKeys() []string {
	keys := make([]string, 0, len(stems))
	for k := range stems {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	return keys
}

func isAllCaps(s string) bool {
	return s == strings.ToUpper(s) && s != strings.ToLower(s)
}

// hint proposes the correctly spelled word for a stemPart+suffix match. It
// returns ok=false for an all-caps match: re-casing German's "ß" to
// uppercase has no single correct ASCII answer (it can be "SS" or "ẞ"
// depending on convention), so guessing one is worse than not hinting.
func hint(stemPart, suffix string) (string, bool) {
	whole := stemPart + suffix
	if isAllCaps(whole) {
		return "", false
	}
	correct, ok := stems[strings.ToLower(stemPart)]
	if !ok {
		return "", false
	}
	if r := []rune(stemPart); len(r) > 0 && unicode.IsUpper(r[0]) {
		cr := []rune(correct)
		cr[0] = unicode.ToUpper(cr[0])
		correct = string(cr)
	}
	return correct + suffix, true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: checklanguage <windows-strings|cmd-ascii> <file>...")
		fmt.Fprintln(os.Stderr, "       checklanguage backend-comments <file>... -- <protection-source>...")
		os.Exit(2)
	}
	mode := os.Args[1]
	files := os.Args[2:]

	var hits []string
	var err error
	switch mode {
	case "windows-strings":
		hits, err = checkWindowsStrings(files)
	case "cmd-ascii":
		hits, err = checkCmdASCII(files)
	case "backend-comments":
		checked, protection := splitAtDoubleDash(files)
		hits, err = checkBackendComments(checked, protection)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", mode)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, h := range hits {
		fmt.Println(h)
	}
	if len(hits) > 0 {
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
// of kind STRING) — normal ("...") and raw (`...`) alike. Comments are never
// visited because ast.Inspect walks the syntax tree, not the token stream —
// they are excluded by construction.
func checkWindowsStrings(files []string) ([]string, error) {
	fset := token.NewFileSet()
	var hits []string
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return hits, fmt.Errorf("parse %s: %w", path, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			if hasNonASCII(lit.Value) {
				pos := fset.Position(lit.Pos())
				hits = append(hits, fmt.Sprintf("%s:%d: non-ASCII byte in Go string literal: %s", path, pos.Line, lit.Value))
			}
			return true
		})
	}
	return hits, nil
}

// checkCmdASCII reports non-ASCII bytes anywhere in a .cmd file, including
// REM comment lines: Windows batch files carry no doc-comment convention
// that jotti's German prose could live in instead, and the printed jotti
// console output stays ASCII throughout, so the whole file stays ASCII.
// CRLF-safe: the trailing "\r" a CRLF line leaves after splitting on "\n"
// is itself ASCII, so it never trips hasNonASCII.
func checkCmdASCII(files []string) ([]string, error) {
	var hits []string
	for _, path := range files {
		content, err := os.ReadFile(path) //nolint:gosec // path comes from the tool's own file-list argument (git ls-files), not untrusted input
		if err != nil {
			return hits, fmt.Errorf("read %s: %w", path, err)
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if hasNonASCII(line) {
				hits = append(hits, fmt.Sprintf("%s:%d: non-ASCII byte: %s", path, i+1, strings.TrimRight(line, "\r")))
			}
		}
	}
	return hits, nil
}

// isDirectiveLine reports whether a line-comment line is a compiler
// directive (//go:build, //go:generate, ...): the "//" is followed
// immediately by "go:", with no space. Directive lines are structural, not
// prose, and are skipped before the stem scan.
func isDirectiveLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "//go:")
}

// docHeaders maps a doc *ast.CommentGroup to the exact identifier name it
// documents, for every declaration in the given files that has one: a
// func, a type, a single-name var/const, or a single-name struct field or
// interface method. Go doc convention has such a comment open with that
// exact name; checkBackendComments protects only that opening word (see
// startsAtFirstWord), not every other occurrence of the same text — a
// type's own doc header aside, its name is an ordinary word in prose
// everywhere else and still needs a real umlaut when one is missing.
func docHeaders(files []*ast.File) map[*ast.CommentGroup]string {
	headers := make(map[*ast.CommentGroup]string)
	add := func(doc *ast.CommentGroup, name string) {
		if doc != nil && name != "" {
			headers[doc] = name
		}
	}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.FuncDecl:
				add(d.Doc, d.Name.Name)
			case *ast.TypeSpec:
				add(d.Doc, d.Name.Name)
			case *ast.ValueSpec:
				if len(d.Names) == 1 {
					add(d.Doc, d.Names[0].Name)
				}
			case *ast.Field:
				if len(d.Names) == 1 {
					add(d.Doc, d.Names[0].Name)
				}
			case *ast.GenDecl:
				// A single-spec block ("// Foo counts widgets.\nvar Foo int")
				// carries its doc on the GenDecl, not the spec.
				if len(d.Specs) == 1 {
					switch s := d.Specs[0].(type) {
					case *ast.TypeSpec:
						add(d.Doc, s.Name.Name)
					case *ast.ValueSpec:
						if len(s.Names) == 1 {
							add(d.Doc, s.Names[0].Name)
						}
					}
				}
			}
			return true
		})
	}
	return headers
}

// startsAtFirstWord reports whether the text before matchStart in line is
// only a comment marker ("//", "/*", " * " on a block-comment continuation
// line) and whitespace — i.e. the match is the very first word of the
// line, the position Go doc convention puts a declaration's name in.
func startsAtFirstWord(line string, matchStart int) bool {
	return strings.Trim(line[:matchStart], "/* \t") == ""
}

// hasInternalCapital reports whether s has an uppercase letter anywhere
// after its first rune. German capitalizes only the first letter of a
// word, compound nouns included (Störungsprotokoll, not
// StörungsProtokoll), so a word with a capital further in is structurally
// a multi-part Go identifier (WriteEventWithDruckauftraege), never plain
// prose — that holds regardless of where in a comment it appears, unlike
// the doc-header check, which only covers a declaration's own opening
// word.
func hasInternalCapital(s string) bool {
	r := []rune(s)
	for i := 1; i < len(r); i++ {
		if unicode.IsUpper(r[i]) {
			return true
		}
	}
	return false
}

// splitAtDoubleDash splits an argument list at the first "--": the files
// to check come first, the files that make up the protected word set come
// after. The two lists differ at both ends — backend/sqlc/dbgen/** is
// generated and never checked, yet its queries carry the column names
// comments quote; this tool's own package is checked but must never
// protect, because its stems map is a list of misspellings, not of names.
func splitAtDoubleDash(args []string) (checked, protection []string) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

// isWordRune reports whether r is a word character, matching what "\w"
// means to Go's regexp package: ASCII letters, digits and "_".
func isWordRune(r rune) bool {
	return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

// tokenSeparators hold a word together with its neighbours into one
// identifier-shaped token instead of leaving it a word of prose: a route
// path (/admin/get-tse-stoerungen), a table or column name
// (naechster_versuch_am), an enum value (pro_stueck), an event type
// (kassensitzung-eroeffnet:v1) or a qualified name (kasse.Stoerung).
const tokenSeparators = "/_-:."

func isTokenByte(b byte) bool {
	return isWordRune(rune(b)) || strings.IndexByte(tokenSeparators, b) >= 0
}

// enclosingToken returns the token around the match at [start,end): the
// surrounding run of word characters and tokenSeparators, with leading
// and trailing separators trimmed. A sentence's closing full stop is
// therefore not part of the token, while the "/" and "-" of a route path
// are.
func enclosingToken(line string, start, end int) string {
	for start > 0 && isTokenByte(line[start-1]) {
		start--
	}
	for end < len(line) && isTokenByte(line[end]) {
		end++
	}
	return strings.Trim(line[start:end], tokenSeparators)
}

// protectedWords collects every spelling the backend's own code uses: the
// name of every package and of every declared or referenced identifier,
// plus the whole text of every Go string literal. Comments quote these
// names verbatim — a route, an event type, a table, a column, an enum
// value — and no declaration spells one with an umlaut, so a comment word
// that matches one is a reference and not misspelled prose. A literal
// contributes its text as one word and is never split into parts: a part
// of a compound name ("pruefen" out of "/admin/tse-setup-pruefen") is an
// ordinary German word on its own, and a comment that means the compound
// writes it whole, where enclosingToken already protects it. Keys are
// lower-cased, and so must lookups be.
func protectedWords(files []*ast.File) map[string]bool {
	protected := make(map[string]bool)
	for _, f := range files {
		protected[strings.ToLower(f.Name.Name)] = true
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				protected[strings.ToLower(x.Name)] = true
			case *ast.BasicLit:
				if x.Kind == token.STRING {
					protected[strings.ToLower(strings.Trim(x.Value, "`\""))] = true
				}
			}
			return true
		})
	}
	return protected
}

// inQuotes reports whether the match at [start,end) sits between a pair
// of double quotes or of backticks on the comment line — how a comment
// marks a value it quotes ("pro_stueck", `regel`). German prose quotes
// with the typographic „…" pair instead, which this deliberately leaves
// unprotected.
func inQuotes(line string, start, end int) bool {
	for _, quote := range []byte{'"', '`'} {
		open := -1
		for i := 0; i < len(line); i++ {
			if line[i] != quote {
				continue
			}
			if open < 0 {
				open = i
				continue
			}
			if open < start && end <= i {
				return true
			}
			open = -1
		}
	}
	return false
}

// isPackageDocName reports whether the match opens a package doc header
// ("// Package dsfinvkpruefung prüft ..."), the one position Go doc
// convention puts a package's own name in.
func isPackageDocName(line string, matchStart int) bool {
	prefix := strings.Trim(line[:matchStart], "/* \t")
	return prefix == "Package" || prefix == "package"
}

// isReference reports whether the stem match at [start,end) spells a name
// the code itself uses, which a comment has to repeat byte for byte: a
// word the backend declares or writes in a string literal, a word inside
// an identifier-shaped token, a quoted value, a package doc's own package
// name, or a word with an inner capital that German spelling never
// produces (WriteEventWithDruckauftraege) but a Go identifier routinely
// does.
func isReference(line string, start, end int, protected map[string]bool) bool {
	word := line[start:end]
	return protected[strings.ToLower(word)] ||
		strings.ContainsAny(enclosingToken(line, start, end), tokenSeparators) ||
		inQuotes(line, start, end) ||
		isPackageDocName(line, start) ||
		hasInternalCapital(word)
}

type parsedFile struct {
	path string
	file *ast.File
}

// checkBackendComments reports transliterated word stems on the comment
// lines of files, skipping every match that names code rather than
// misspelling prose (isReference) and every doc comment's own opening
// word (docHeaders). The protected word set comes from protectionSources
// alone; their own comments are never checked. A block comment's
// ast.Comment.Text carries embedded "\n"s, so the reported line is the
// comment's start line plus the newline count before the match.
func checkBackendComments(files, protectionSources []string) ([]string, error) {
	fset := token.NewFileSet()
	parsed := make([]parsedFile, 0, len(files))
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		parsed = append(parsed, parsedFile{path: path, file: f})
	}
	sources := make([]*ast.File, 0, len(protectionSources))
	for _, path := range protectionSources {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		sources = append(sources, f)
	}
	astFiles := make([]*ast.File, 0, len(parsed))
	for _, p := range parsed {
		astFiles = append(astFiles, p.file)
	}
	headers := docHeaders(astFiles)
	protected := protectedWords(sources)

	var hits []string
	for _, p := range parsed {
		path, f := p.path, p.file
		for _, group := range f.Comments {
			headerName, isHeader := headers[group]
			for commentIdx, c := range group.List {
				startLine := fset.Position(c.Slash).Line
				for lineOffset, line := range strings.Split(c.Text, "\n") {
					if isDirectiveLine(line) {
						continue
					}
					for _, m := range stemPattern.FindAllStringSubmatchIndex(line, -1) {
						whole := line[m[0]:m[1]]
						stemPart := line[m[2]:m[3]]
						suffix := line[m[4]:m[5]]
						isDocHeader := isHeader && commentIdx == 0 && lineOffset == 0 &&
							whole == headerName && startsAtFirstWord(line, m[0])
						if isDocHeader || isReference(line, m[0], m[1], protected) {
							continue
						}
						lineNo := startLine + lineOffset
						if repl, ok := hint(stemPart, suffix); ok {
							hits = append(hits, fmt.Sprintf("%s:%d: transliterated stem %q (use %q) in comment", path, lineNo, whole, repl))
						} else {
							hits = append(hits, fmt.Sprintf("%s:%d: transliterated stem %q in comment (all-caps, fix by hand)", path, lineNo, whole))
						}
					}
				}
			}
		}
	}
	return hits, nil
}
