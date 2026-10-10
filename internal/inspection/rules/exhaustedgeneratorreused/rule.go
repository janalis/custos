// Package exhaustedgeneratorreused implements the native ExhaustedGeneratorReused inspection.
package exhaustedgeneratorreused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Create a fresh generator before traversing again."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExhaustedGeneratorReused" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach, syntax.KFuncCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	var input syntax.Expr
	switch x := n.(type) {
	case *syntax.Foreach:
		input = x.Expr
	case *syntax.FuncCall:
		if !semanticquery.NativeBuiltin(ctx, x, "iterator_to_array") {
			return
		}
		input = semanticquery.CallArgument(x.Args, 0, "iterator")
	}
	producer, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok {
		return
	}
	if !semanticquery.NativeGeneratorBody(semanticquery.NativeFunctionBody(ctx, producer)) {
		return
	}
	if _, known := ctx.Flow().StateBefore(input, input, "iterator_to_array"); known {
		ctx.ReportNode(n, message)
		return
	}
	f, ok := n.(*syntax.Foreach)
	if !ok {
		return
	}
	previous, ok := astquery.PrevStmt(ctx.File, f)
	if !ok {
		return
	}
	first, ok := previous.(*syntax.Foreach)
	if !ok || !astquery.Equivalent(ctx.File, first.Expr, input) {
		return
	}
	body, ok := first.Body.(*syntax.Block)
	if !ok || len(body.Stmts) != 0 {
		return
	}
	gbody := semanticquery.NativeFunctionBody(ctx, producer)
	for _, st := range gbody.Stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			return
		}
		y, ok := es.Expr.(*syntax.Yield)
		if !ok || flowquery.MayHaveSideEffects(y.Value) {
			return
		}
	}
	ctx.ReportNode(n, message)
}
