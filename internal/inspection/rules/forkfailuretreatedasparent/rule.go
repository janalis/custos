// Package forkfailuretreatedasparent implements the native ForkFailureTreatedAsParent inspection.
package forkfailuretreatedasparent

import (
	"math"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Handle fork failure separately."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForkFailureTreatedAsParent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// Only a file spelling it can reach pcntl_fork (see astquery.Mentions).
	if !astquery.Mentions(ctx.File, "pcntl_fork") {
		return
	}
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	branch := n.(*syntax.If)
	input := syntax.UnwrapParens(branch.Cond)
	if u, ok := input.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		input = u.Expr
	}
	fork, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, fork, "pcntl_fork") || failureExcluded(ctx, input, fork) {
		return
	}
	ctx.ReportNode(branch.Cond, message)
}

// failureExcluded reports whether the tests on the fork result guarding the
// condition have already ruled out the failed fork (-1).
func failureExcluded(ctx *analysis.Context, input syntax.Expr, fork *syntax.FuncCall) bool {
	if semanticquery.NativeLabelled(ctx, input) {
		return false
	}
	t := ctx.Types().Native().CallResultAsAt(fork, input, types.Ints(types.IntRange{Lo: -1, Hi: math.MaxInt64}))
	return !t.IsUnknown() && !t.HasInt(-1)
}
