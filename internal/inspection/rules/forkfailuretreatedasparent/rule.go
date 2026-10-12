// Package forkfailuretreatedasparent implements the native ForkFailureTreatedAsParent inspection.
package forkfailuretreatedasparent

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle fork failure separately."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForkFailureTreatedAsParent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	branch := n.(*syntax.If)
	input := syntax.UnwrapParens(branch.Cond)
	if u, ok := input.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		input = u.Expr
	}
	fork, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, fork, "pcntl_fork") || failureExcluded(ctx, branch, fork) {
		return
	}
	ctx.ReportNode(branch.Cond, message)
}

// failureExcluded reports whether every path reaching branch has already
// ruled out the failed fork: an earlier statement of an enclosing list leaves
// on failure, or branch sits behind the else side of a failure test.
func failureExcluded(ctx *analysis.Context, branch *syntax.If, fork *syntax.FuncCall) bool {
	for n := syntax.Node(branch); n != nil && !syntax.IsFuncLike(n); n = n.Parent() {
		if chain, ok := n.Parent().(*syntax.If); ok && n != syntax.Node(chain.Body) {
			// An else or elseif arm runs only after every earlier condition failed.
			if testsFailure(ctx, chain.Cond, fork) {
				return true
			}
			for _, arm := range chain.ElseIfs {
				if arm.Span().Start < n.Span().Start && testsFailure(ctx, arm.Cond, fork) {
					return true
				}
			}
		}
		if stmt, ok := n.(syntax.Stmt); ok {
			list, at, _ := astquery.StmtList(ctx.File, stmt)
			for _, earlier := range list[:max(at, 0)] {
				if guard, ok := earlier.(*syntax.If); ok && syntax.Terminates(guard.Body) && testsFailure(ctx, guard.Cond, fork) {
					return true
				}
			}
		}
	}
	return false
}

// testsFailure reports whether cond is a comparison that holds whenever the
// result of fork is -1, with the fork result on either side.
func testsFailure(ctx *analysis.Context, cond syntax.Expr, fork *syntax.FuncCall) bool {
	b, ok := syntax.UnwrapParens(cond).(*syntax.Binary)
	if !ok {
		return false
	}
	mirrored := b.Op.Kind
	switch mirrored {
	case syntax.TGreater:
		mirrored = syntax.TLess
	case syntax.TIsGreaterOrEqual:
		mirrored = syntax.TIsSmallerOrEqual
	case syntax.TLess, syntax.TIsSmallerOrEqual:
		mirrored = syntax.TGreater
	}
	return coversFailure(ctx, b.Left, b.Op.Kind, b.Right, fork) || coversFailure(ctx, b.Right, mirrored, b.Left, fork)
}

// coversFailure checks `subject op bound` where subject is a local still
// holding the value returned by fork, or the assignment storing that call,
// and bound is a known integer.
func coversFailure(ctx *analysis.Context, subject syntax.Expr, op syntax.TokenKind, bound syntax.Expr, fork *syntax.FuncCall) bool {
	switch tested := syntax.UnwrapParens(subject).(type) {
	case *syntax.Variable:
		if semanticquery.NativeValue(ctx, tested) != syntax.Expr(fork) {
			return false
		}
	case *syntax.Assign:
		if tested.Op.Kind != syntax.TEqual || semanticquery.NativeUnsilenced(tested.Value) != syntax.Expr(fork) {
			return false
		}
	default:
		return false
	}
	limit, known := semanticquery.NativeInt(ctx, bound)
	if !known {
		return false
	}
	switch op {
	case syntax.TIsIdentical, syntax.TIsEqual:
		return limit == -1
	case syntax.TLess:
		return limit >= 0
	case syntax.TIsSmallerOrEqual:
		return limit >= -1
	}
	return false
}
