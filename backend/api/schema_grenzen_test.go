//go:build unit

package api

// An exported *Schema variable under backend/domain is the bound of a persisted
// field: handlers and constructors validate every input against it before the
// value reaches a table or an event. A field schema without an upper bound lets
// the Kasse store a value that no Beleg and no DSFinV-K export can render — and
// the schema is the only place that bound exists, because the columns are TEXT
// and INT without a length or range of their own.
//
// This test pins both ends. The table names, per field schema, the value at each
// bound the schema must accept and the value just outside it must reject; a new
// exported schema without a row makes the test red. A schema that carries no
// length or value bound at all — an enum, a struct schema, an alias — belongs in
// schemaExceptions with the reason it needs no row.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/domain/user"
)

// domainSchemaDir is the tree the discovery walks, relative to this package.
const domainSchemaDir = "../domain"

// maxInt4 is the largest value a PostgreSQL int4 column holds, and therefore the
// largest ID any row can carry.
const maxInt4 = math.MaxInt32

// lengthCase pins a string schema to its two length bounds. shortest and longest
// are the lengths it must accept; one character less than shortest and one more
// than longest must be rejected. filler is the character the test value is built
// from — it has to satisfy the schema's format rule, if it has one. A shortest of
// 0 marks an optional field: the empty value is accepted and there is no shorter
// one to reject.
type lengthCase struct {
	schema   string
	shortest int
	longest  int
	filler   rune
	accepts  func(wert string) bool
}

// lengthCases holds one row per exported string field schema under
// backend/domain.
var lengthCases = []lengthCase{
	{schema: "produkt.NameSchema", shortest: 3, longest: 100, filler: 'a', accepts: acceptsString(produkt.NameSchema)},
	{schema: "tisch.TischNameSchema", shortest: 3, longest: 100, filler: 'a', accepts: acceptsString(tisch.TischNameSchema)},
	{schema: "user.NameSchema", shortest: 3, longest: 50, filler: 'a', accepts: acceptsString(user.NameSchema)},
	{schema: "user.OnetimePasswordSchema", shortest: 6, longest: 6, filler: '1', accepts: acceptsString(user.OnetimePasswordSchema)},
	{schema: "user.PasswordSchema", shortest: 6, longest: 72, filler: 'a', accepts: acceptsString(user.PasswordSchema)},
	{schema: "user.UsernameSchema", shortest: 3, longest: 20, filler: 'a', accepts: acceptsString(user.UsernameSchema)},
}

// valueCase pins a number schema to its two value bounds. smallest and largest
// are the values it must accept; one less than smallest and one more than largest
// must be rejected. An ID is assigned by the database and never built from an
// input, so it deliberately carries no upper bound: its row sets largest to
// maxInt4 and unbounded, which drops the rejection above.
type valueCase struct {
	schema    string
	smallest  int
	largest   int
	unbounded bool
	accepts   func(wert int) bool
}

// valueCases holds one row per exported number field schema under
// backend/domain.
var valueCases = []valueCase{
	{schema: "kasse.PositionEingabeSchema", smallest: 1, largest: 999, accepts: acceptsNumber(kasse.PositionEingabeSchema)},
	{schema: "produkt.IDSchema", smallest: 1, largest: maxInt4, unbounded: true, accepts: acceptsNumber(produkt.IDSchema)},
	{schema: "produkt.PreisCentsSchema", smallest: 1, largest: 99999, accepts: acceptsNumber(produkt.PreisCentsSchema)},
	{schema: "tisch.TischIDSchema", smallest: 1, largest: maxInt4, unbounded: true, accepts: acceptsNumber(tisch.TischIDSchema)},
	{schema: "user.IDSchema", smallest: 1, largest: maxInt4, unbounded: true, accepts: acceptsNumber(user.IDSchema)},
}

// schemaException records an exported schema the table needs no row for, with the
// reason it carries no length or value bound.
type schemaException struct {
	schema string
	reason string
}

// schemaExceptions holds the exported schemas that bound no field length or
// value.
var schemaExceptions = []schemaException{
	{schema: "produkt.KategorieSchema", reason: "enum: OneOf over the three Kategorie values"},
	{schema: "produkt.ProduktSchema", reason: "struct schema: composed of the field schemas that carry the bounds"},
	{schema: "produkt.RichtungSchema", reason: "enum: OneOf over the two Richtung values"},
	{schema: "produkt.StatusSchema", reason: "enum: OneOf over the three Status values"},
	{schema: "produkt.SteuersatzSchema", reason: "alias of steuer.SteuersatzSchema, which has its own entry"},
	{schema: "produkt.VarianteSchema", reason: "struct schema: composed of the field schemas that carry the bounds"},
	{schema: "steuer.SteuersatzSchema", reason: "enum: OneOf over the four Steuersatz values"},
	{schema: "tisch.TischSchema", reason: "struct schema: composed of the field schemas that carry the bounds"},
	{schema: "tisch.TischStatusSchema", reason: "enum: OneOf over the three Status values"},
	{schema: "user.RoleSchema", reason: "enum: OneOf over the three Role values"},
	{schema: "user.StatusSchema", reason: "enum: OneOf over the three Status values"},
	{schema: "user.UserSchema", reason: "struct schema: composed of the field schemas that carry the bounds"},
}

// TestSchemaGrenzen fails when an exported schema under backend/domain reaches no
// table row, and when a row's schema misses one of its bounds. Adding a field
// schema therefore forces a decision: pin its bounds in the table, or document in
// schemaExceptions why it bounds nothing.
func TestSchemaGrenzen(t *testing.T) {
	declared := declaredFieldSchemas(t)
	if len(declared) == 0 {
		t.Fatalf("scan of %s found no exported schema; the walk is broken", domainSchemaDir)
	}

	covered := map[string]bool{}
	for _, testCase := range lengthCases {
		covered[testCase.schema] = true
		checkLengthBounds(t, testCase)
	}
	for _, testCase := range valueCases {
		covered[testCase.schema] = true
		checkValueBounds(t, testCase)
	}

	for _, key := range sortedKeys(declared) {
		if covered[key] || exceptionReason(key) != "" {
			continue
		}
		t.Errorf("%s: schema %s has no row in lengthCases or valueCases — pin its bounds there, or add it to schemaExceptions with a reason", declared[key], key)
	}

	for _, key := range sortedKeys(covered) {
		if _, ok := declared[key]; !ok {
			t.Errorf("lengthCases or valueCases holds a row for %s, which no package under %s declares; drop the row", key, domainSchemaDir)
		}
		if reason := exceptionReason(key); reason != "" {
			t.Errorf("%s has both a row and a schemaExceptions entry (%s); keep the row and drop the entry", key, reason)
		}
	}

	for _, exception := range schemaExceptions {
		if _, ok := declared[exception.schema]; !ok {
			t.Errorf("schemaExceptions lists %s (%s), which no package under %s declares; drop the entry", exception.schema, exception.reason, domainSchemaDir)
		}
	}
}

// checkLengthBounds asserts that the schema accepts a value at each of its length
// bounds and rejects one character beyond them.
func checkLengthBounds(t *testing.T, testCase lengthCase) {
	t.Helper()

	if !testCase.accepts(fill(testCase.filler, testCase.shortest)) {
		t.Errorf("%s rejects length %d, its shortest accepted one", testCase.schema, testCase.shortest)
	}
	if !testCase.accepts(fill(testCase.filler, testCase.longest)) {
		t.Errorf("%s rejects length %d, its longest accepted one", testCase.schema, testCase.longest)
	}
	// Below the empty value there is nothing to reject, so only a field with a
	// minimum length has a low side to check.
	if testCase.shortest > 0 && testCase.accepts(fill(testCase.filler, testCase.shortest-1)) {
		t.Errorf("%s accepts length %d, one below its minimum %d", testCase.schema, testCase.shortest-1, testCase.shortest)
	}
	if testCase.accepts(fill(testCase.filler, testCase.longest+1)) {
		t.Errorf("%s accepts length %d, one above its maximum %d — a persisted field needs an upper bound", testCase.schema, testCase.longest+1, testCase.longest)
	}
}

// checkValueBounds asserts that the schema accepts each of its value bounds and
// rejects the value beyond them.
func checkValueBounds(t *testing.T, testCase valueCase) {
	t.Helper()

	if !testCase.accepts(testCase.smallest) {
		t.Errorf("%s rejects %d, its smallest accepted value", testCase.schema, testCase.smallest)
	}
	if !testCase.accepts(testCase.largest) {
		t.Errorf("%s rejects %d, its largest accepted value", testCase.schema, testCase.largest)
	}
	if testCase.accepts(testCase.smallest - 1) {
		t.Errorf("%s accepts %d, one below its minimum %d", testCase.schema, testCase.smallest-1, testCase.smallest)
	}
	if testCase.unbounded {
		return
	}
	if testCase.accepts(testCase.largest + 1) {
		t.Errorf("%s accepts %d, one above its maximum %d — a persisted field needs an upper bound", testCase.schema, testCase.largest+1, testCase.largest)
	}
}

// acceptsString builds the check of a string schema. zog validates through a
// pointer and its Trim transform writes back, so every check gets its own copy of
// the value.
func acceptsString(schema *z.StringSchema[string]) func(string) bool {
	return func(wert string) bool {
		kopie := wert
		return schema.Validate(&kopie) == nil
	}
}

// acceptsNumber builds the check of a number schema.
func acceptsNumber(schema *z.NumberSchema[int]) func(int) bool {
	return func(wert int) bool {
		kopie := wert
		return schema.Validate(&kopie) == nil
	}
}

// fill returns a string of n filler characters.
func fill(filler rune, n int) string {
	return strings.Repeat(string(filler), n)
}

// exceptionReason returns the reason schemaExceptions gives for the schema, or an
// empty string when it lists none.
func exceptionReason(schema string) string {
	for _, exception := range schemaExceptions {
		if exception.schema == schema {
			return exception.reason
		}
	}

	return ""
}

// declaredFieldSchemas returns every exported *Schema variable declared in a
// non-test file below backend/domain, keyed as "<package directory>.<identifier>"
// and valued by the position of its declaration.
func declaredFieldSchemas(t *testing.T) map[string]token.Position {
	t.Helper()

	declared := map[string]token.Position{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(domainSchemaDir, func(path string, entry fs.DirEntry, err error) error {
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

		pkgDir := strings.TrimPrefix(filepath.ToSlash(filepath.Dir(path)), domainSchemaDir+"/")
		for _, name := range exportedSchemaNames(file) {
			declared[pkgDir+"."+name.Name] = fset.Position(name.Pos())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", domainSchemaDir, err)
	}

	return declared
}

// exportedSchemaNames returns the name of every exported variable in file whose
// identifier ends in Schema.
func exportedSchemaNames(file *ast.File) []*ast.Ident {
	names := []*ast.Ident{}
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
				if isSchemaName(name.Name) {
					names = append(names, name)
				}
			}
		}
	}

	return names
}

// isSchemaName reports whether name is that of an exported schema variable:
// exported and suffixed Schema, with something before the suffix.
func isSchemaName(name string) bool {
	return token.IsExported(name) && strings.HasSuffix(name, "Schema") && len(name) > len("Schema")
}

// sortedKeys returns the keys of m in lexical order, so the test reports its
// findings in the same order on every run.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
