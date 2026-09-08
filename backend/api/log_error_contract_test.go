//go:build unit

package api

// An error-level log line inside an error branch is the only record of what went
// wrong: the client gets a code, the operator gets the log. A chain without
// .Err(err) drops the cause — the message then names the operation but never the
// reason, and nilerr cannot see it because the error is still handled.
//
// This test parses every non-test file below backend/api and fails on a
// log.Error() chain in a branch guarded by "<err> != nil" that carries no .Err(.
// It recognizes a chain by its terminating .Msg/.Msgf call and reads the selector
// names down the receiver chain, so it holds for zerolog.Ctx(ctx).Error() as well
// as for a stored logger.

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

// TestLogErrorCarriesTheError fails when an error branch logs at error level
// without passing the error along.
func TestLogErrorCarriesTheError(t *testing.T) {
	checked := 0
	forEachSourceFile(t, func(fset *token.FileSet, file *ast.File, _ string) {
		ast.Inspect(file, func(node ast.Node) bool {
			ifStmt, ok := node.(*ast.IfStmt)
			if !ok || !isErrorBranch(ifStmt.Cond) {
				return true
			}
			ast.Inspect(ifStmt.Body, func(inner ast.Node) bool {
				call, ok := inner.(*ast.CallExpr)
				if !ok {
					return true
				}
				chain, isLog := logChain(call)
				if !isLog || !chain["Error"] {
					return true
				}
				checked++
				if !chain["Err"] {
					t.Errorf("%s: log.Error() chain in an error branch carries no .Err(err) — the cause is lost", fset.Position(call.Pos()))
				}
				return true
			})
			return true
		})
	})

	if checked == 0 {
		t.Fatal("scan found no error-level log chain in any error branch; the walk is broken")
	}
}

// isErrorBranch reports whether cond is "<something named …err> != nil". The name
// is what separates an error branch from any other nil check.
func isErrorBranch(cond ast.Expr) bool {
	binary, ok := cond.(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return false
	}
	if ident, ok := binary.Y.(*ast.Ident); !ok || ident.Name != "nil" {
		return false
	}

	switch operand := binary.X.(type) {
	case *ast.Ident:
		return isErrorName(operand.Name)
	case *ast.SelectorExpr:
		return isErrorName(operand.Sel.Name)
	default:
		return false
	}
}

// isErrorName reports whether name is that of an error value: err, or a
// qualified variant such as lookupErr or writeErr.
func isErrorName(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), "err")
}

// logChain returns the set of method names of the logging chain that call
// terminates, and whether call terminates one at all. A chain ends in .Msg or
// .Msgf; the names are collected down the receiver chain, so
// log.Error().Err(err).Msg("…") yields {Msg, Err, Error}.
func logChain(call *ast.CallExpr) (map[string]bool, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || (selector.Sel.Name != "Msg" && selector.Sel.Name != "Msgf") {
		return nil, false
	}

	names := map[string]bool{}
	for current := ast.Expr(call); ; {
		inner, ok := current.(*ast.CallExpr)
		if !ok {
			return names, true
		}
		step, ok := inner.Fun.(*ast.SelectorExpr)
		if !ok {
			return names, true
		}
		names[step.Sel.Name] = true
		current = step.X
	}
}
