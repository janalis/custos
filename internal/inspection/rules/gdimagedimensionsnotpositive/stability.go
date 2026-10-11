package gdimagedimensionsnotpositive

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// scopeFacts summarises the constructs of one scope that can change a local
// without a visible assignment to it. Nested function-like and class bodies
// run on their own schedule and are not entered.
type scopeFacts struct {
	// wild: a write whose target name is only known at run time
	// (`$$name++`, extract(), parse_str()).
	wild bool
	// opaque: code that can run arbitrary application logic, and so rewrite
	// a variable shared with another scope.
	opaque bool
}

func factsOf(ctx *analysis.Context, scope syntax.Node) scopeFacts {
	var facts scopeFacts
	visit := func(n syntax.Node) bool {
		if n == scope {
			return true
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike:
			return false
		case *syntax.Closure, *syntax.ArrowFunction:
			// May be invoked later, including by a builtin taking a callback.
			facts.opaque = true
			return false
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Eval:
			facts.opaque = true
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, x) {
			case "":
				facts.opaque = true
			case "extract", "parse_str":
				facts.wild = true
			}
		case *syntax.IncDec:
			facts.wild = facts.wild || dynamicVariable(x.Var)
		case *syntax.Assign:
			facts.wild = facts.wild || x.Op.Kind != syntax.TEqual && dynamicVariable(x.Var)
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return facts
}

func dynamicVariable(e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.NameExpr != nil
}

// stableDimension reports whether the integer resolved for the dimension e
// can be trusted: every local it reads, directly or through the values those
// locals were assigned, must be stable in the scope of the call.
func stableDimension(ctx *analysis.Context, e syntax.Expr) bool {
	names := variableNames(e)
	if len(names) == 0 {
		return true
	}
	scope := syntax.EnclosingFuncLike(e)
	facts := factsOf(ctx, scope)
	seen := map[string]bool{}
	for _, name := range names {
		if !stableLocal(ctx, scope, name, facts, seen) {
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

// stableLocal reports whether the value the flow engine resolved for $name
// can be trusted at any point of scope. The engine does not model `++`/`--`,
// catch bindings, by-reference closure imports, run-time named writes, or
// calls rewriting a variable shared with another scope; each of them removes
// the proof, also when it hits a variable $name was assigned from.
func stableLocal(ctx *analysis.Context, scope syntax.Node, name string, facts scopeFacts, seen map[string]bool) bool {
	if facts.wild {
		return false
	}
	if seen[name] {
		return true
	}
	seen[name] = true
	shared := scope == nil
	for _, p := range syntax.FuncLikeParams(scope) {
		shared = shared || p.ByRef && p.Var != nil && p.Var.Name == name
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			shared = shared || use.ByRef && use.Var != nil && use.Var.Name == name
		}
	}
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		switch by := access.By.(type) {
		case *syntax.IncDec, *syntax.Catch:
			return false
		case *syntax.ClosureUse:
			if by.ByRef {
				return false
			}
		case *syntax.Global, *syntax.StaticVar:
			shared = true
		case *syntax.Assign:
			for _, source := range variableNames(by.Value) {
				if !stableLocal(ctx, scope, source, facts, seen) {
					return false
				}
			}
		}
	}
	return !shared || !facts.opaque
}
