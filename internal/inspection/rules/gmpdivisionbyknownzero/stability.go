package gmpdivisionbyknownzero

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// stableDivisor reports whether the value resolved for the divisor e can be
// trusted: every local it reads, directly or through the values those locals
// were assigned, must only change through assignments the flow engine models.
func stableDivisor(ctx *analysis.Context, e syntax.Expr) bool {
	names := variableNames(e)
	if len(names) == 0 {
		return true
	}
	scope := syntax.EnclosingFuncLike(e)
	if importsVariables(ctx, scope) {
		return false
	}
	seen := map[string]bool{}
	for _, name := range names {
		if !stableLocal(ctx, scope, name, seen) {
			return false
		}
	}
	return true
}

// variableNames lists the plain variables read by e.
func variableNames(e syntax.Expr) []string {
	var names []string
	syntax.Inspect(e, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
			names = append(names, v.Name)
		}
		return true
	})
	return names
}

// importsVariables reports whether scope calls extract() or parse_str(),
// which write locals whose names are only known at run time. Nested
// function-like and class bodies are separate scopes and are not entered.
func importsVariables(ctx *analysis.Context, scope syntax.Node) bool {
	found := false
	visit := func(n syntax.Node) bool {
		if n == scope {
			return true
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, x) {
			case "extract", "parse_str":
				found = true
			}
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// stableLocal reports whether the value the flow engine resolved for $name
// can be trusted anywhere in scope. The engine does not model `++`/`--`,
// catch bindings or by-reference closure imports; each of them removes the
// proof, also when it hits a variable $name was assigned from.
func stableLocal(ctx *analysis.Context, scope syntax.Node, name string, seen map[string]bool) bool {
	if seen[name] {
		return true
	}
	seen[name] = true
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		switch by := access.By.(type) {
		case *syntax.IncDec, *syntax.Catch:
			return false
		case *syntax.ClosureUse:
			if by.ByRef {
				return false
			}
		case *syntax.Assign:
			for _, source := range variableNames(by.Value) {
				if !stableLocal(ctx, scope, source, seen) {
					return false
				}
			}
		}
	}
	return true
}
