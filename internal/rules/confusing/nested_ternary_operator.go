package confusing

import (
	"custos/internal/analysis"
	"custos/internal/syntax"
)

// nestedTernaryOperator reports ternaries used directly as operands of
// another ternary, except unparenthesised short-ternary chains.
type nestedTernaryOperator struct{}

func init() { register(nestedTernaryOperator{}) }

const nestedTernaryOperatorMsg = "Avoid nesting ternary operators; use if/else or extract a variable."

func (nestedTernaryOperator) ID() string { return "NestedTernaryOperator" }

func (nestedTernaryOperator) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (nestedTernaryOperator) Check(ctx *analysis.Context, n syntax.Node) {
	o := n.(*syntax.Ternary)
	short := o.Then == nil
	check := func(op syntax.Expr) {
		inner, ok := syntax.UnwrapParens(op).(*syntax.Ternary)
		if !ok || inner.Span().Len() == 0 {
			return
		}
		if short && inner.Then == nil && syntax.Expr(inner) == op { // E1
			return
		}
		ctx.ReportNode(inner, nestedTernaryOperatorMsg)
	}
	check(o.Cond) // D1
	if !short {
		check(o.Then) // D2
	}
	check(o.Else) // D3
}
