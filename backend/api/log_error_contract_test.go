package api

// An error-level log line in an error branch is the operator's only record, so it must carry an "error" attribute.
// Without it the cause is lost, and nilerr cannot see it because the error is still handled.

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

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
				if !isErrorLogCall(call) {
					return true
				}
				checked++
				if !carriesErrorAttr(call) {
					t.Errorf("%s: Error log call in an error branch carries no \"error\" attribute — the cause is lost", fset.Position(call.Pos()))
				}
				return true
			})
			return true
		})
	})

	if checked == 0 {
		t.Error("scan found no error-level log chain in any error branch; the walk is broken")
	}
}

// The name is what separates an error branch from any other nil check.
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

// An error value is any name ending in err: err, lookupErr, writeErr.
func isErrorName(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), "err")
}

// isErrorLogCall matches slog.Error("…", …), log.Error("…", …), logger.Error("…", …) and
// logging.FromContext(ctx).Error("…", …): the receiver names every logger below backend/api.
func isErrorLogCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Error" || len(call.Args) == 0 {
		return false
	}
	switch receiver := selector.X.(type) {
	case *ast.Ident:
		return receiver.Name == "slog" || receiver.Name == "log" || receiver.Name == "logger"
	case *ast.CallExpr:
		fromContext, ok := receiver.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := fromContext.X.(*ast.Ident)
		return ok && pkg.Name == "logging" && fromContext.Sel.Name == "FromContext"
	default:
		return false
	}
}

// carriesErrorAttr reports whether a key argument after the message is the literal "error".
func carriesErrorAttr(call *ast.CallExpr) bool {
	for i := 1; i < len(call.Args); i += 2 {
		if key, ok := call.Args[i].(*ast.BasicLit); ok && key.Value == `"error"` {
			return true
		}
	}
	return false
}
