//go:build unit

package api

// An exported Err* variable under backend/api is an application sentinel: the
// application layer returns it, and the HTTP layer turns it into a stable error
// code that frontend/src/lib/errorMessages.ts renders as a German message. A
// sentinel the HTTP layer never names falls into MapError's fallback and reaches
// the client as a bare 500 — the user then sees the generic server-error text for
// a failure the backend understood exactly.
//
// This test parses the tree instead of relying on review: it collects every
// exported Err* variable declared below backend/api and every sentinel the HTTP
// layer names, and fails on a sentinel that appears in neither. A sentinel counts
// as named when a non-test file in an api/**/http package references it inside a
// helper.MapError call or an errors.Is guard — the two shapes the error contract
// takes here. A sentinel the HTTP layer names nowhere belongs in
// mappingExceptions, with the reason it needs no code of its own.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// apiImportPrefix is the import path of this package. Only imports below it can
// carry an application sentinel, so only they need alias resolution.
const apiImportPrefix = "github.com/nicograef/jotti/backend/api/"

// errDatabaseName is the sentinel every application package aliases from
// db.ErrDatabase. It deliberately carries no error code: a database failure is
// nothing the client can act on, so MapError's fallback answers 500 and the
// frontend shows the generic server-error message with the log reference.
const errDatabaseName = "ErrDatabase"

// mappingException records a sentinel no HTTP file names, with the reason it
// needs no error code of its own. sentinel holds a sentinelKey value.
type mappingException struct {
	sentinel string
	reason   string
}

// mappingExceptions holds the sentinels the HTTP layer names nowhere.
var mappingExceptions = []mappingException{
	{
		sentinel: "auth/application.ErrTokenGeneration",
		reason:   "a server-side failure (JWT signing, or an unexpected error out of VerifyPassword); the client has nothing to correct",
	},
	{
		sentinel: "fiskal/dsfinvk.ErrKeineVorgaenge",
		reason:   "reaches the client as leere_kassensitzung through its alias export/application.ErrLeereKassensitzung",
	},
	{
		sentinel: "stammdaten/user/application.ErrInvalidUserData",
		reason:   "the handler validates the same three user schemas the domain re-checks, so an accepted body never trips the domain check",
	},
}

// TestErrorMappingContract fails when an application sentinel reaches no error
// code. Adding a sentinel therefore forces a decision: name it in the handler, or
// document in mappingExceptions why it needs no code.
func TestErrorMappingContract(t *testing.T) {
	declared := collectDeclaredSentinels(t)
	named := collectNamedSentinels(t)

	if len(declared) == 0 || len(named) == 0 {
		t.Fatalf("scan found %d declared and %d named sentinels; the walk is broken", len(declared), len(named))
	}

	for key, pos := range declared {
		if strings.HasSuffix(key, "."+errDatabaseName) {
			continue
		}
		if isExcepted(key) {
			continue
		}
		if !named[key] {
			t.Errorf("%s: sentinel %s reaches no error code — name it in a helper.MapError call or an errors.Is guard under api/**/http, or add it to mappingExceptions with a reason", pos, key)
		}
	}

	for _, exception := range mappingExceptions {
		if _, ok := declared[exception.sentinel]; !ok {
			t.Errorf("mappingExceptions lists %s (%s), which no package declares; drop the entry", exception.sentinel, exception.reason)
		}
		if named[exception.sentinel] {
			t.Errorf("mappingExceptions lists %s (%s), but the HTTP layer names it; drop the entry", exception.sentinel, exception.reason)
		}
	}
}

// isExcepted reports whether mappingExceptions covers the sentinel named by key.
func isExcepted(key string) bool {
	for _, exception := range mappingExceptions {
		if exception.sentinel == key {
			return true
		}
	}

	return false
}

// collectDeclaredSentinels returns every exported Err* variable declared in a
// non-test file below this package, keyed by sentinelKey and valued by the
// position of its declaration.
func collectDeclaredSentinels(t *testing.T) map[string]token.Position {
	t.Helper()

	declared := map[string]token.Position{}
	forEachSourceFile(t, func(fset *token.FileSet, file *ast.File, pkgDir string) {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if isSentinelName(name.Name) {
						declared[sentinelKey(pkgDir, name.Name)] = fset.Position(name.Pos())
					}
				}
			}
		}
	})

	return declared
}

// collectNamedSentinels returns the sentinels the HTTP layer names, as the set of
// sentinelKey values referenced inside a helper.MapError call or an errors.Is
// guard in a non-test file of an api/**/http package.
func collectNamedSentinels(t *testing.T) map[string]bool {
	t.Helper()

	named := map[string]bool{}
	forEachSourceFile(t, func(_ *token.FileSet, file *ast.File, pkgDir string) {
		if filepath.Base(pkgDir) != "http" {
			return
		}

		aliases := importAliases(file)
		helperAlias := aliases[apiImportPrefix+"helper"]
		errorsAlias := aliases["errors"]

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isQualifiedCall(call, helperAlias, "MapError") && !isQualifiedCall(call, errorsAlias, "Is") {
				return true
			}
			for _, arg := range call.Args {
				for _, key := range sentinelRefs(arg, aliases, pkgDir) {
					named[key] = true
				}
			}
			return true
		})
	})

	return named
}

// sentinelRefs returns the sentinelKey of every sentinel referenced in expr,
// resolving a qualified reference through the file's import aliases and an
// unqualified one against the referencing package itself.
func sentinelRefs(expr ast.Expr, aliases map[string]string, pkgDir string) []string {
	keys := []string{}
	ast.Inspect(expr, func(node ast.Node) bool {
		switch ref := node.(type) {
		case *ast.SelectorExpr:
			qualifier, ok := ref.X.(*ast.Ident)
			if !ok || !isSentinelName(ref.Sel.Name) {
				return true
			}
			for path, alias := range aliases {
				if alias == qualifier.Name && strings.HasPrefix(path, apiImportPrefix) {
					keys = append(keys, sentinelKey(strings.TrimPrefix(path, apiImportPrefix), ref.Sel.Name))
				}
			}
			return false
		case *ast.Ident:
			if isSentinelName(ref.Name) {
				keys = append(keys, sentinelKey(pkgDir, ref.Name))
			}
		}
		return true
	})

	return keys
}

// forEachSourceFile parses every non-test Go file below this package and calls
// visit with the file and its package directory, relative to backend/api and
// slash-separated.
func forEachSourceFile(t *testing.T, visit func(fset *token.FileSet, file *ast.File, pkgDir string)) {
	t.Helper()

	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return parseErr
		}

		visit(fset, file, filepath.ToSlash(filepath.Dir(path)))
		return nil
	})
	if err != nil {
		t.Fatalf("walk backend/api: %v", err)
	}
}

// importAliases maps each import path of file to the name it is bound to: the
// explicit alias, or the last path segment when there is none. Every package
// below backend/api is named after its directory, so the segment is the name.
func importAliases(file *ast.File) map[string]string {
	aliases := map[string]string{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		if spec.Name != nil {
			aliases[path] = spec.Name.Name
			continue
		}
		aliases[path] = path[strings.LastIndex(path, "/")+1:]
	}

	return aliases
}

// isQualifiedCall reports whether call is qualifier.name(…). An empty qualifier
// never matches: it means the file does not import that package at all.
func isQualifiedCall(call *ast.CallExpr, qualifier string, name string) bool {
	if qualifier == "" {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != name {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == qualifier
}

// isSentinelName reports whether name is that of an application sentinel:
// exported and prefixed Err, with something following the prefix.
func isSentinelName(name string) bool {
	return strings.HasPrefix(name, "Err") && len(name) > len("Err")
}

// sentinelKey names a sentinel by package directory and identifier, e.g.
// "stammdaten/tisch/application.ErrTischNotFound".
func sentinelKey(pkgDir string, name string) string {
	return pkgDir + "." + name
}
