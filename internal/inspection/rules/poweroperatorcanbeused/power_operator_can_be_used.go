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
	if _, ok := call.Parent().(*syntax.Binary); ok {
		r = "(" + r + ")"
	}
	ctx.ReportNode(call, "Use '"+r+"' (exponentiation operator) instead.", diagnostic.Fix{ // D4
		Title: "Use the ** operator",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: call.Span(), NewText: r}}
		},
	})
}

// powOperand renders a pow() argument as an operand of `**`. Binary and
// ternary expressions are parenthesised (upstream behaviour); assignments,
// and unary expressions in base position, are also wrapped so the result
// keeps its meaning (`pow(-2, 2)` -> `(-2) ** 2`).
func powOperand(ctx *analysis.Context, e syntax.Expr, base bool) string {
	t := ctx.Text(e)
	switch v := e.(type) {
	case *syntax.Binary, *syntax.Instanceof, *syntax.Ternary, *syntax.Assign:
		return "(" + t + ")"
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
