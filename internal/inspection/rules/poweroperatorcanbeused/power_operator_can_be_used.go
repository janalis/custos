package poweroperatorcanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// powerOperatorCanBeUsed reports pow($a, $b) calls that can use `**`.
type powerOperatorCanBeUsed struct{}

// Semantic marks the rule as needing the project index (a user function
// named pow may be declared in another file).
func (powerOperatorCanBeUsed) Semantic()  {}
func (powerOperatorCanBeUsed) ID() string { return "PowerOperatorCanBeUsed" }
func (powerOperatorCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (powerOperatorCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP56 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	if _, name, ok := astquery.CallName(call); !ok || !strings.EqualFold(name, "pow") { // D1 (any case)
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "pow") { // D1
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 2 { // D2, E2
		return
	}
	base, exp := args[0], args[1]
	r := powOperand(ctx, base, true) + " ** " + powOperand(ctx, exp, false) // D3
	if powNeedsOuterParens(call) {
		r = "(" + r + ")"
	}
	msg := "Use '" + r + "' (exponentiation operator) instead."
	// F2: a comment between the call's punctuation and its arguments has no
	// place in the operator form; keep the report, withhold the fix.
	f, cs, bs, es := ctx.File, call.Span(), base.Span(), exp.Span()
	if astquery.HasComment(f, syntax.Span{Start: cs.Start, End: bs.Start}) ||
		astquery.HasComment(f, syntax.Span{Start: bs.End, End: es.Start}) ||
		astquery.HasComment(f, syntax.Span{Start: es.End, End: cs.End}) {
		ctx.ReportNode(call, msg)
		return
	}
	ctx.ReportNode(call, msg, diagnostic.Fix{ // D4
		Title: "Use the ** operator",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: cs, NewText: r}}
		},
	})
}

// powNeedsOuterParens reports whether the whole `b ** e` must be wrapped
// where the call stood: as an operand of a binary operator (upstream
// behaviour), as the subject of `instanceof` or of a postfix access/call
// (`pow(2, 3)[0]`, `pow($g, 2)->x`), which would otherwise bind to the
// exponent alone, and as the operand of the `clone` keyword.
func powNeedsOuterParens(call *syntax.FuncCall) bool {
	switch p := call.Parent().(type) {
	case *syntax.Binary, *syntax.Clone:
		return true
	case *syntax.Instanceof, *syntax.ArrayDimFetch, *syntax.PropertyFetch, *syntax.MethodCall,
		*syntax.FuncCall, *syntax.StaticCall, *syntax.StaticPropertyFetch, *syntax.ClassConstFetch:
		// The call is the leftmost child (not an index, name or class operand).
		return p.Span().Start == call.Span().Start
	}
	return false
}

// powOperand renders a pow() argument as an operand of `**`. Binary and
// ternary expressions are parenthesised (upstream behaviour); assignments,
// the low-precedence keyword expressions (yield, print, include, throw,
// arrow functions), and unary expressions in base position, are also
// wrapped so the result keeps its meaning (`pow(-2, 2)` -> `(-2) ** 2`).
func powOperand(ctx *analysis.Context, e syntax.Expr, base bool) string {
	t := ctx.Text(e)
	if astquery.NeedsParensAsUnaryOperand(e) {
		return "(" + t + ")"
	}
	switch v := e.(type) {
	case *syntax.Unary:
		if base {
			return "(" + t + ")"
		}
	case *syntax.IncDec:
		if base && v.Prefix {
			return "(" + t + ")"
		}
	}
	return t
}
