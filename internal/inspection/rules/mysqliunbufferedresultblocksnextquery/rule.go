// Package mysqliunbufferedresultblocksnextquery implements the native MysqliUnbufferedResultBlocksNextQuery inspection.
package mysqliunbufferedresultblocksnextquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Consume or free the unbuffered result before another query."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliUnbufferedResultBlocksNextQuery" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli", "query") {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) == 0 {
		return
	}
	st, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := st.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return
	}
	if _, ok := assignment.Var.(*syntax.Variable); !ok {
		return
	}
	producer, ok := assignment.Value.(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "mysqli", "query") {
		return
	}
	if semanticquery.NativeValue(ctx, c.Var) == nil || semanticquery.NativeValue(ctx, c.Var) != semanticquery.NativeValue(ctx, producer.Var) {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(producer.Args, 0, "query"))
	if !known {
		return
	}
	parts := strings.Fields(strings.ToUpper(sql))
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "SELECT", "SHOW", "DESCRIBE", "EXPLAIN":
	default:
		return
	}
	mode, ok := semanticquery.CallArgument(producer.Args, 1, "result_mode").(*syntax.ConstFetch)
	if ok && semanticquery.GlobalConstName(ctx, mode) == "MYSQLI_USE_RESULT" {
		ctx.ReportNode(c, message)
	}
}
