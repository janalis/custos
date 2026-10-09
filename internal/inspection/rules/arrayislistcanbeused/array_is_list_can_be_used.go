package arrayislistcanbeused

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// arrayIsListCanBeUsed reports hand-written list checks built on
// array_values()/array_keys() that array_is_list() expresses directly.
type arrayIsListCanBeUsed struct{}

func (arrayIsListCanBeUsed) ID() string               { return "ArrayIsListCanBeUsed" }
func (arrayIsListCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (arrayIsListCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	// D1: array_is_list() exists from 8.1 (spec Divergences: upstream gates at 8.0).
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	call := n.(*syntax.FuncCall)
	_, _, ok := astquery.CallName(call)
	name := ctx.GlobalFunctionName(call) // any case, global function only
	if !ok || (name != "array_values" && name != "array_keys") {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 1 { // E2
		return
	}
	bin, ok := call.Parent().(*syntax.Binary) // D2
	if !ok {
		return
	}
	// D2: strict comparisons only; loose `==`/`!=` compare key/value pairs
	// loosely and are not equivalent to array_is_list() (spec Divergences).
	negated := false
	switch bin.Op.Kind {
	case syntax.TIsIdentical:
	case syntax.TIsNotIdentical:
		negated = true
	default:
		return
	}
	other := bin.Right
	if bin.Right == syntax.Expr(call) {
		other = bin.Left
	}
	arg := args[0]
	if name == "array_values" {
		if !astquery.EquivalentFoldNames(ctx.File, other, arg) { // D3
			return
		}
	} else if !isZeroToCountMinusOne(ctx, other, arg) { // D4
		return
	}
	repl := semanticquery.QualifiedBuiltinFor(ctx, "array_is_list", call) + "(" + ctx.Text(arg) + ")" // D6
	if negated {
		repl = "!" + repl
	}
	msg := repl
	if name == "array_keys" {
		// range(0, -1) is [0, -1]: the range form is false for [] while
		// array_is_list([]) is true.
		if negated {
			repl = ctx.Text(arg) + " === [] || " + repl
		} else {
			repl = ctx.Text(arg) + " !== [] && " + repl
		}
		msg = repl
		switch bin.Parent().(type) {
		case *syntax.Binary, *syntax.Unary, *syntax.Ternary, *syntax.Instanceof:
			repl = "(" + repl + ")"
		}
	}
	span := bin.Span()
	ctx.Report(span, "Replace with '"+msg+"'.", diagnostic.Fix{
		Title: "Replace with '" + msg + "'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// isZeroToCountMinusOne reports whether e is `range(0, count(arg) - 1)`.
func isZeroToCountMinusOne(ctx *analysis.Context, e, arg syntax.Expr) bool {
	rng, ok := e.(*syntax.FuncCall)
	if !ok || ctx.GlobalFunctionName(rng) != "range" {
		return false
	}
	rargs, ok := astquery.CallArgValues(rng)
	if !ok || len(rargs) != 2 || !isLiteralText(ctx, rargs[0], syntax.LitInt, "0") {
		return false
	}
	sub, ok := rargs[1].(*syntax.Binary)
	if !ok || sub.Op.Kind != syntax.TMinus {
		return false
	}
	if !isLiteralText(ctx, sub.Right, syntax.LitInt, "1") {
		return false
	}
	cnt, ok := sub.Left.(*syntax.FuncCall)
	if !ok || ctx.GlobalFunctionName(cnt) != "count" {
		return false
	}
	cargs, ok := astquery.CallArgValues(cnt)
	return ok && len(cargs) == 1 && astquery.EquivalentFoldNames(ctx.File, cargs[0], arg)
}

func isLiteralText(ctx *analysis.Context, e syntax.Expr, k syntax.LiteralKind, raw string) bool {
	lit, ok := e.(*syntax.Literal)
	return ok && lit.LitKind == k && lit.Raw == raw
}
