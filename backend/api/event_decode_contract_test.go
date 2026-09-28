package api

// Event-JSON decodes only into event contract types (kasse.PositionEventData, payload types), never into domain types like kasse.Position.
// Domain types carry no json tags and Go matches keys case-insensitively, so a renamed field silently decodes to its zero value.

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"testing"
)

const kassePackagePath = "github.com/nicograef/jotti/backend/domain/kasse"

// The two event contract families: a payload part (PositionEventData) and a
// whole event payload (BestellungAufgenommenV1Data).
var eventVertragTyp = regexp.MustCompile(`(EventData|V[0-9]+Data)$`)

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
		t.Error("scan found no json-tagged struct below backend/api; the walk is broken")
	}
}

// A json tag is what makes a struct a JSON (de)serialization target.
func hasJSONTag(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if field.Tag != nil && strings.Contains(field.Tag.Value, "json:") {
			return true
		}
	}

	return false
}

// packageTypeNames finds the names expr selects from pkgName at any depth: a
// slice element, a pointer target and a map value count like a plain field type.
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
