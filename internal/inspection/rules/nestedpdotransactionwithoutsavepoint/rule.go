// Package nestedpdotransactionwithoutsavepoint implements the native NestedPdoTransactionWithoutSavepoint inspection.
package nestedpdotransactionwithoutsavepoint

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Finish the transaction before beginning another one."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NestedPdoTransactionWithoutSavepoint" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "beginTransaction") {
		return
	}
	if !startsStatement(c) {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) == 0 {
		return
	}
	statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	previous, ok := statement.Expr.(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, previous, "PDO", "beginTransaction") && semanticquery.NativeValue(ctx, c.Var) != nil && semanticquery.NativeValue(ctx, c.Var) == semanticquery.NativeValue(ctx, previous.Var) {
		ctx.ReportNode(c, message)
	}
}

// startsStatement reports whether the call is the first thing its statement
// evaluates, unconditionally: the statement expression, a returned value or
// an if condition, possibly under parentheses, unary operators, a plain
// variable assignment, or as the left operand of a binary operator or the
// condition of a ternary. Anywhere else an operand evaluated earlier in the
// same statement may finish the transaction or short-circuit the call away.
func startsStatement(c *syntax.MethodCall) bool {
	var e syntax.Node = c
	for {
		switch parent := e.Parent().(type) {
		case *syntax.Paren, *syntax.Unary:
		case *syntax.Binary:
			if parent.Left != e {
				return false
			}
		case *syntax.Ternary:
			if parent.Cond != e {
				return false
			}
		case *syntax.Assign:
			if _, plain := parent.Var.(*syntax.Variable); !plain || parent.Op.Kind != syntax.TEqual {
				return false
			}
		case *syntax.ExprStmt, *syntax.Return:
			return true
		case *syntax.If:
			return parent.Cond == e
		default:
			return false
		}
		e = e.Parent()
	}
}
