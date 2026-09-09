//go:build unit

package api

// Event-JSON is decoded through the event contract types only: a decode target
// names kasse.PositionEventData or a payload type such as
// kasse.BestellungAufgenommenV1Data, never kasse.Position. The domain type carries
// no json tags (AGENTS.md rule 10) and matches the stored keys by accident — Go's
// decoder compares field names case-insensitively. A renamed domain field or a
// field the projection adds for display would then silently decode to its zero
// value, and a wrong Arbeitsbon is the first place that shows.
//
// This test parses every non-test file below backend/api and fails when a struct
// with json tags carries a field of a domain/kasse type that is not an event
// contract type: those end in EventData (a nested part of a payload) or in V<n>Data
// (a whole event payload). The same boundary keeps rule 10: a domain struct is
// never serialized as an API response either.

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"testing"
)

// kassePackagePath is the domain package that holds the event contract types; the
// walk resolves it to the name each file binds it to.
const kassePackagePath = "github.com/nicograef/jotti/backend/domain/kasse"

// eventVertragTyp matches the two event contract families of domain/kasse: a
// payload part (PositionEventData) and a whole event payload
// (BestellungAufgenommenV1Data).
var eventVertragTyp = regexp.MustCompile(`(EventData|V[0-9]+Data)$`)

// TestDecodeTargetsUseEventDataTypes fails when a JSON struct below backend/api
// names a domain/kasse type that is not an event contract type.
func TestDecodeTargetsUseEventDataTypes(t *testing.T) {
	decodeTargets := 0
	forEachSourceFile(t, func(fset *token.FileSet, file *ast.File, _ string) {
		kasseName, imported := importAliases(file)[kassePackagePath]

		ast.Inspect(file, func(node ast.Node) bool {
			structType, ok := node.(*ast.StructType)
			if !ok || !hasJSONTag(structType) {
				return true
			}
			decodeTargets++
			if !imported {
				return true
			}

			for _, field := range structType.Fields.List {
				for _, named := range packageTypeNames(field.Type, kasseName) {
					if eventVertragTyp.MatchString(named) {
						continue
					}
					t.Errorf(
						"%s: json struct field of type %s.%s is no event contract type — decode event JSON into a type ending in EventData or V<n>Data and convert it",
						fset.Position(field.Pos()), kasseName, named,
					)
				}
			}
			return true
		})
	})

	if decodeTargets == 0 {
		t.Fatal("scan found no json-tagged struct below backend/api; the walk is broken")
	}
}

// hasJSONTag reports whether any field of structType carries a json tag, which is
// what makes the struct a JSON (de)serialization target.
func hasJSONTag(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if field.Tag != nil && strings.Contains(field.Tag.Value, "json:") {
			return true
		}
	}

	return false
}

// packageTypeNames returns the type names that expr selects from pkgName, at any
// depth: a slice element, a pointer target and a map value count like a plain
// field type.
func packageTypeNames(expr ast.Expr, pkgName string) []string {
	if pkgName == "" {
		return nil
	}

	var names []string
	ast.Inspect(expr, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkgName {
			names = append(names, selector.Sel.Name)
		}
		return true
	})

	return names
}
