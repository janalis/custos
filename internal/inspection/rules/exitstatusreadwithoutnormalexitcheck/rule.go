// Package exitstatusreadwithoutnormalexitcheck implements the native ExitStatusReadWithoutNormalExitCheck inspection.
package exitstatusreadwithoutnormalexitcheck

import (
	"math"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Check normal termination before decoding the exit code."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExitStatusReadWithoutNormalExitCheck" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pcntl_wexitstatus") {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "status")
	wait, flags := semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_waitpid"}, 1, "status"), 2
	if wait == nil {
		wait, flags = semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_wait"}, 0, "status"), 1
	}
	if wait == nil {
		return
	}
	blocking := semanticquery.CallArgument(wait.Args, flags, "flags") == nil
	if semanticquery.NativeLabelled(ctx, c) || !waited(ctx, wait, c, blocking) || !exited(ctx, wait, input, c) {
		ctx.ReportNode(c, message)
	}
}

// waited reports whether the wait result is known to be a child pid where
// read runs: it fails with -1 and, unless it blocks (flags omitted), may
// also return 0 when no child changed state.
func waited(ctx *analysis.Context, wait, read *syntax.FuncCall, blocking bool) bool {
	seed := types.Ints(types.IntRange{Lo: -1, Hi: math.MaxInt64})
	if blocking {
		seed = types.Ints(types.IntRange{Lo: -1, Hi: -1}, types.IntRange{Lo: 1, Hi: math.MaxInt64})
	}
	return ctx.Types().Native().CallResultAsAt(wait, read, seed).IntsOnly(1, math.MaxInt64)
}

// exited reports whether a pcntl_wifexited() test of the status, made after
// the wait, is known to have returned true where read runs.
func exited(ctx *analysis.Context, wait *syntax.FuncCall, status syntax.Expr, read *syntax.FuncCall) bool {
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= read.Span().Start || n != syntax.EnclosingVariableScope(read) && syntax.IsVariableScope(n) {
			return false
		}
		if call, ok := n.(*syntax.FuncCall); ok && call.Span().Start > wait.Span().End && semanticquery.NativeBuiltin(ctx, call, "pcntl_wifexited") && semanticquery.ExpansionCSame(ctx, status, semanticquery.CallArgument(call.Args, 0, "status")) {
			found = ctx.Types().Native().CallResultAt(call, read).OnlyOf("true")
		}
		return !found
	}
	if scope := syntax.EnclosingVariableScope(read); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}
