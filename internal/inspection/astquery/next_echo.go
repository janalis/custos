package astquery

import "custos/internal/php/syntax"

// NextEcho returns the expression echoed by the statement right after s
// when that statement echoes exactly one, without parentheses; nil
// otherwise.
func NextEcho(f *syntax.File, s syntax.Stmt) syntax.Expr {
	next, ok := NextStmt(f, s)
	if !ok {
		return nil
	}
	echo, ok := next.(*syntax.Echo)
	if !ok || len(echo.Exprs) != 1 {
		return nil
	}
	return syntax.UnwrapParens(echo.Exprs[0])
}

// CallsExtract reports whether scope (the file when nil) calls extract(),
// which can bind any local variable.
func CallsExtract(f *syntax.File, scope syntax.Node) bool {
	found := false
	visit := func(x syntax.Node) bool {
		if _, ok := IsFuncNamedFold(x, "extract"); ok {
			found = true
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}
