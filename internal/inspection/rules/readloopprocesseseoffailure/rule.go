// Package readloopprocesseseoffailure implements the native ReadLoopProcessesEofFailure inspection.
package readloopprocesseseoffailure

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Loop on successful reads before consuming line data."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadLoopProcessesEofFailure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KWhile} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.While)
	if ctx.PHP < phpversion.PHP70 || !strict(ctx) {
		return
	}
	condition, ok := syntax.UnwrapParens(loop.Cond).(*syntax.Unary)
	if !ok || condition.Op.Kind != syntax.TExclaim {
		return
	}
	eof, ok := syntax.UnwrapParens(condition.Expr).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, eof, "feof") {
		return
	}
	handle := semanticquery.CallArgument(eof.Args, 0, "stream")
	if _, ok := handle.(*syntax.Variable); !ok {
		return
	}
	block, ok := loop.Body.(*syntax.Block)
	if !ok {
		return
	}
	for i, st := range block.Stmts {
		value := consumer(ctx, st)
		if value != nil && sameRead(ctx, value, handle) {
			ctx.ReportNode(loop.Cond, message)
			return
		}
		statement, ok := st.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		assignment, ok := statement.Expr.(*syntax.Assign)
		if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual || i+1 >= len(block.Stmts) {
			continue
		}
		target, ok := assignment.Var.(*syntax.Variable)
		if !ok {
			continue
		}
		if v, ok := consumer(ctx, block.Stmts[i+1]).(*syntax.Variable); ok && v.Name == target.Name && sameRead(ctx, assignment.Value, handle) {
			ctx.ReportNode(loop.Cond, message)
			return
		}
	}
}

func strict(ctx *analysis.Context) bool {
	for _, st := range ctx.File.Stmts {
		if decl, ok := st.(*syntax.Declare); ok {
			for _, item := range decl.Items {
				if item.Key.Value == "strict_types" {
					n, known := semanticquery.NativeInt(ctx, item.Value)
					return known && n == 1
				}
			}
		}
	}
	return false
}

func consumer(ctx *analysis.Context, st syntax.Stmt) syntax.Expr {
	var expression syntax.Expr
	switch s := st.(type) {
	case *syntax.ExprStmt:
		expression = s.Expr
	case *syntax.Echo:
		if len(s.Exprs) == 1 {
			expression = s.Exprs[0]
		}
	}
	if a, ok := expression.(*syntax.Assign); ok {
		expression = a.Value
	}
	call, ok := expression.(*syntax.FuncCall)
	if !ok {
		return nil
	}
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "strlen" && name != "trim" && name != "strtoupper" {
		return nil
	}
	return semanticquery.CallArgument(call.Args, 0, "string")
}

func sameRead(ctx *analysis.Context, e, handle syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "fgets") {
		return false
	}
	v := semanticquery.CallArgument(c.Args, 0, "stream")
	hv, rv := ctx.Flow().Value(handle), ctx.Flow().Value(v)
	return hv.Complete && rv.Complete && hv.Identity != 0 && hv.Identity == rv.Identity && hv.Invalidated == rv.Invalidated
}
