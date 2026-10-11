package arrayuniquecanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// arrayUniqueCanBeUsed reports array_keys()/count() over array_count_values(),
// which array_unique() expresses directly since PHP 7.2.
type arrayUniqueCanBeUsed struct{}

func (arrayUniqueCanBeUsed) ID() string { return "ArrayUniqueCanBeUsed" }
func (arrayUniqueCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (arrayUniqueCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP72 { // D1
		return
	}
	inner, ok := astquery.IsFuncNamedFold(n, "array_count_values") // D2
	if !ok || !ctx.IsGlobalFunctionCall(inner, "array_count_values") {
		return
	}
	args, ok := astquery.CallArgValues(inner)
	if !ok || len(args) != 1 {
		return
	}
	outer := astquery.ParentFuncCall(inner) // D3
	if outer == nil {
		return
	}
	// Divergence (spec): only when the inner call is the outer call's sole argument.
	if oargs, ok := astquery.CallArgValues(outer); !ok || len(oargs) != 1 {
		return
	}
	a := ctx.Text(args[0])
	// F1: each call keeps its leading `\` and gets one when a bare name
	// would not reach the global function.
	unique := semanticquery.QualifiedBuiltinFor(ctx, "array_unique", inner)
	var repl, fixed string
	switch ctx.GlobalFunctionName(outer) {
	case "array_keys":
		repl = "array_values(array_unique(" + a + "))"
		fixed = semanticquery.QualifiedBuiltinFor(ctx, "array_values", outer) + "(" + unique + "(" + a + "))"
	case "count":
		repl = "count(array_unique(" + a + "))"
		fixed = semanticquery.QualifiedBuiltinFor(ctx, "count", outer) + "(" + unique + "(" + a + "))"
	default:
		return
	}
	if !aucSafe(ctx, args[0], ctx.GlobalFunctionName(outer) == "array_keys") || aucAliased(ctx, inner, args[0]) {
		return
	}
	span := outer.Span()
	ctx.Report(span, "Use '"+repl+"' instead (array_unique() is fast since PHP 7.2).", diagnostic.Fix{
		Title: "Use array_unique()",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: fixed}}
		},
	})
}

// The key form also needs to preserve PHP's numeric-string key coercion.
func aucSafe(ctx *analysis.Context, arg syntax.Expr, keys bool) bool {
	t := ctx.Types().Native().TypeOf(arg)
	if t.IsArrayLike() && !t.Has("array") && (t.Elem().OnlyOf("int") || (!keys && t.Elem().OnlyOf("int", "string"))) {
		return true
	}
	arr := semanticquery.NativeArray(ctx, arg)
	if arr == nil {
		return false
	}
	for _, item := range arr.Items {
		if _, ok := semanticquery.NativeInt(ctx, item.Value); ok {
			continue
		}
		key, ok := semanticquery.NativeArrayKey(ctx, item.Value)
		if !ok || (keys && !strings.HasPrefix(key, "s:")) {
			return false
		}
	}
	return true
}

// aucAliased reports whether a variable of the argument may be written
// through a reference somewhere in its scope: neither the inferred element
// types nor the resolved literal account for such writes, so the proof of
// the element domain does not hold.
func aucAliased(ctx *analysis.Context, inner *syntax.FuncCall, arg syntax.Expr) bool {
	aliased := false
	visit := func(name string) func(syntax.Node) bool {
		return func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Param:
				aliased = aliased || (x.ByRef && x.Var.Name == name)
			case *syntax.ClosureUse:
				aliased = aliased || (x.ByRef && x.Var.Name == name)
			case *syntax.Assign:
				aliased = aliased || (x.ByRef && astquery.MentionsVariable(x, name))
			case *syntax.ArrayItem:
				aliased = aliased || (x.ByRef && astquery.MentionsVariable(x.Value, name))
			case *syntax.Foreach:
				// A by-value loop over the array is harmless; the array as a
				// loop target, or iterated by reference, is not.
				aliased = aliased || astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name) ||
					(x.ByRef && astquery.MentionsVariable(x.Expr, name))
			case *syntax.Include, *syntax.Eval:
				aliased = true // shares the local variables
			case *syntax.FuncCall:
				_, extract := astquery.IsFuncNamedFold(x, "extract")
				aliased = aliased || extract || (x != inner && aucCallMayWrite(ctx, x, x.Args, name))
			case *syntax.MethodCall:
				aliased = aliased || aucCallMayWrite(ctx, x, x.Args, name)
			case *syntax.StaticCall:
				aliased = aliased || aucCallMayWrite(ctx, x, x.Args, name)
			case *syntax.New:
				aliased = aliased || aucCallMayWrite(ctx, x, x.Args, name)
			}
			return !aliased
		}
	}
	scope := syntax.EnclosingVariableScope(inner)
	syntax.Inspect(arg, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && !aliased {
			if scope == nil {
				syntax.InspectFile(ctx.File, visit(v.Name))
			} else {
				syntax.Inspect(scope, visit(v.Name))
			}
		}
		return !aliased
	})
	return aliased
}

// aucCallMayWrite reports whether a call receives $name, or one of its
// elements, in a position that is not proven to be passed by value.
func aucCallMayWrite(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList, name string) bool {
	if list == nil {
		return false
	}
	passed := false
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			continue
		}
		base := syntax.UnwrapParens(arg.Value)
		for {
			dim, ok := base.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			base = syntax.UnwrapParens(dim.Var)
		}
		if v, ok := base.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
			passed = true
		}
	}
	if !passed {
		return false
	}
	callee, ok := semanticquery.ResolveCallee(ctx, call)
	return !ok || semanticquery.ArgBindsByRef(list, callee.Params, true)
}
