package substrusedasarrayaccess

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// subStrUsedAsArrayAccess reports substr($s, $i, 1) on a string and suggests
// string offset access.
type subStrUsedAsArrayAccess struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (subStrUsedAsArrayAccess) Semantic()                {}
func (subStrUsedAsArrayAccess) ID() string               { return "SubStrUsedAsArrayAccess" }
func (subStrUsedAsArrayAccess) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (subStrUsedAsArrayAccess) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "substr") // D1
	if call == nil {
		return
	}
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) != 3 { // D2
		return
	}
	if lit, ok := args[2].(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt || lit.Raw != "1" { // D3
		return
	}
	switch args[0].(type) { // D4
	case *syntax.Variable, *syntax.ArrayDimFetch, *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
	default:
		return
	}
	if !ctx.TypeOf(args[0]).Has("string") { // D5
		return
	}
	src, off := ctx.Text(args[0]), ctx.Text(args[1])
	// A type alone says nothing about whether a negative offset clamps.
	offset, known := semanticquery.NativeInt(ctx, args[1])
	if !known || offset < -1 || sharedOffset(args[1]) {
		return
	}
	var access string
	if offset == -1 { // R1: only -1 keeps negative-offset behavior
		v, stable := args[0].(*syntax.Variable)
		if !stable || v.NameExpr != nil {
			return
		}
		access = src + "[" + semanticquery.QualifiedBuiltin(ctx, "strlen", call.Span().Start) + "(" + src + ") - 1]"
	} else { // R2: a proven nonnegative integer
		access = src + "[" + off + "]"
	}
	span := call.Span()
	// R3: before PHP 8, out-of-range substr can return false. Nor fix
	// another scalar: an int's offset is null rather than a digit.
	if ctx.PHP < phpversion.PHP80 || !ctx.TypeOf(args[0]).OnlyOf("string", "null") {
		ctx.Report(span, "Use '"+access+"' (string offset access) instead.")
		return
	}
	repl := "(" + access + " ?? '')"
	if arg, ok := call.Parent().(*syntax.Arg); ok && !arg.Unpack { // F1: an argument needs no parentheses
		repl = access + " ?? ''"
	}
	ctx.Report(span, "Use '"+repl+"' (string offset access) instead.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}

// sharedOffset reports whether the offset is a variable whose assigned value
// may be stale: storage reachable from other code (a global, a static, a
// by-reference parameter, closure import or array item) or a `++`/`--`.
func sharedOffset(e syntax.Expr) bool {
	name, ok := astquery.PlainVariableName(syntax.UnwrapParens(e))
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(e)
	if scope == nil { // top-level variables are globals
		return true
	}
	is := func(x syntax.Node) bool {
		got, ok := astquery.PlainVariableName(x)
		return ok && got == name
	}
	shared := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Param:
			shared = shared || x.ByRef && x.Var != nil && is(x.Var)
		case *syntax.ClosureUse:
			shared = shared || x.ByRef && x.Var != nil && is(x.Var)
		case *syntax.ArrayItem:
			shared = shared || x.ByRef && x.Value != nil && is(syntax.UnwrapParens(x.Value))
		case *syntax.StaticVar:
			shared = shared || x.Var != nil && is(x.Var)
		case *syntax.IncDec:
			shared = shared || is(syntax.UnwrapParens(x.Var))
		case *syntax.Global:
			for _, v := range x.Vars {
				shared = shared || is(v)
			}
		}
		return !shared
	})
	return shared
}
