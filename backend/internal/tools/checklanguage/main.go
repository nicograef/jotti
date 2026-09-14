// Command checklanguage implements the three rules of scripts/check-language.sh:
//
//   - windows-strings: Go string literals (not comments) under windows/** must be
//     pure ASCII — the starter and relay print them straight to a Windows
//     console, which mangles non-ASCII bytes.
//   - cmd-ascii: packaging/**/*.cmd must be pure ASCII throughout, comments
//     included — a batch file has no separate doc-comment channel.
//   - backend-comments: a Go comment under backend/** must not spell a German
//     word stem with the ASCII stand-in for ä/ö/ü (stems map below). String
//     literals are never touched; matches that name code are skipped
//     (isReference, docHeaders).
//
// Exit codes: 0 clean, 1 violations printed, 2 the tool itself failed.
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

// stems maps each ASCII-transliterated German word stem to its correct spelling.
// Matching is by stem (see stemPattern), so one entry catches every inflection; a
// prefixed compound needs its own entry, because the match is anchored at a
// word's start (\b) and German prefixes attach with no separator. Per-stem rather
// than a blanket character substitution: a transliterated double-s becomes "ß" in
// some words (schließen, gemäß, größe, mäßig) and stays a double-s in others
// (müssen).
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

// stemPattern matches a known stem at a word's start, case-insensitively, plus
// the inflection that follows. Longer stems are tried first (see stemKeys), so a
// prefixed compound wins over a shorter stem at the same position.
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

// hint proposes the correct spelling for a stemPart+suffix match. ok=false for an
// all-caps match: upper-casing "ß" has no single correct answer ("SS" or "ẞ"), so
// guessing one is worse than not hinting.
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

// checkWindowsStrings reports non-ASCII bytes in Go string literals, normal and
// raw alike. Comments are excluded by construction: ast.Inspect walks the syntax
// tree, not the token stream.
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

// checkCmdASCII reports non-ASCII bytes anywhere in a .cmd file, REM comments
// included: a batch file has no doc-comment channel for German prose, and the
// printed console output stays ASCII. CRLF-safe — the trailing "\r" is ASCII.
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

// isDirectiveLine reports whether a line comment is a compiler directive ("//"
// followed by "go:", no space). Directives are structural, not prose, and are
// skipped before the stem scan.
func isDirectiveLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "//go:")
}

// docHeaders maps each doc comment to the identifier name it documents (func,
// type, single-name var/const, struct field, interface method). Go doc convention
// opens such a comment with that exact name; only that opening word is protected
// (see startsAtFirstWord) — everywhere else the same word is ordinary prose and
// still needs its umlaut.
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
				// A single-spec block carries its doc on the GenDecl, not on the spec.
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

// startsAtFirstWord reports whether only a comment marker and whitespace precede
// matchStart — the position Go doc convention puts a declaration's name in.
func startsAtFirstWord(line string, matchStart int) bool {
	return strings.Trim(line[:matchStart], "/* \t") == ""
}

// hasInternalCapital reports whether s has an uppercase letter after its first
// rune. German capitalizes only a word's first letter, compound nouns included
// (Störungsprotokoll, not StörungsProtokoll), so an inner capital marks a Go
// identifier (WriteEventWithDruckauftraege), never prose — anywhere in a comment.
func hasInternalCapital(s string) bool {
	r := []rune(s)
	for i := 1; i < len(r); i++ {
		if unicode.IsUpper(r[i]) {
			return true
		}
	}
	return false
}

// splitAtDoubleDash splits the argument list at the first "--": files to check
// first, protection sources after. The two differ — backend/sqlc/dbgen/** is never
// checked but holds column names comments quote; this package is checked but must
// never protect, its stems map being a list of misspellings.
func splitAtDoubleDash(args []string) (checked, protection []string) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

// isWordRune matches what "\w" means to Go's regexp: ASCII letters, digits, "_".
func isWordRune(r rune) bool {
	return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

// tokenSeparators hold a word together with its neighbours into one
// identifier-shaped token: /admin/get-tse-stoerungen, naechster_versuch_am,
// pro_stueck, kassensitzung-eroeffnet:v1, kasse.Stoerung.
const tokenSeparators = "/_-:."

func isTokenByte(b byte) bool {
	return isWordRune(rune(b)) || strings.IndexByte(tokenSeparators, b) >= 0
}

// enclosingToken returns the run of word characters and tokenSeparators around
// [start,end), leading and trailing separators trimmed — a closing full stop is
// not part of the token, the "/" and "-" of a route path are.
func enclosingToken(line string, start, end int) string {
	for start > 0 && isTokenByte(line[start-1]) {
		start--
	}
	for end < len(line) && isTokenByte(line[end]) {
		end++
	}
	return strings.Trim(line[start:end], tokenSeparators)
}

// protectedWords collects every spelling the backend's own code uses: package
// names, declared or referenced identifiers, and the whole text of every string
// literal. A comment quoting such a name repeats it verbatim, and no declaration
// spells one with an umlaut. A literal contributes its text as one word and is
// never split: a part of a compound ("pruefen" out of "/admin/tse-setup-pruefen")
// is an ordinary German word, and enclosingToken already protects the compound.
// Keys are lower-cased, and so must lookups be.
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

// inQuotes reports whether the match sits between a pair of double quotes or
// backticks on the line — how a comment marks a value ("pro_stueck", `regel`).
// German prose quotes with „…" instead, deliberately left unprotected.
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

// isPackageDocName reports whether the match opens a package doc header, the one
// position Go doc convention puts a package's own name in.
func isPackageDocName(line string, matchStart int) bool {
	prefix := strings.Trim(line[:matchStart], "/* \t")
	return prefix == "Package" || prefix == "package"
}

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

// checkBackendComments reports transliterated stems on comment lines, skipping
// matches that name code (isReference) and each doc comment's own opening word
// (docHeaders). The protected word set comes from protectionSources alone; their
// own comments are never checked. A block comment's Text carries embedded "\n"s,
// so the reported line is the comment's start line plus the newlines before it.
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
