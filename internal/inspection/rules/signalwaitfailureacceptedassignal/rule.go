// Package signalwaitfailureacceptedassignal implements the native SignalWaitFailureAcceptedAsSignal inspection.
package signalwaitfailureacceptedassignal

import (
	"math"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Reject signal-wait failure before handling the result."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalWaitFailureAcceptedAsSignal" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// Only a file spelling the builtin can call it (see astquery.Mentions).
	if len(ctx.File.Errors) > 0 || !astquery.Mentions(ctx.File, "pcntl_sigwaitinfo") || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	b := n.(*syntax.If)
	e := syntax.UnwrapParens(b.Cond)
	var assigned syntax.Expr
	if a, ok := e.(*syntax.Assign); ok {
		assigned = a.Var
		e = a.Value
	}
	c, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "pcntl_sigwaitinfo") {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 {
		return
	}
	// An assigned variable stays the wait result up to its first rebinding,
	// alias or dynamic write in the branch; other targets need the first call.
	var limit uint32
	if v, ok := assigned.(*syntax.Variable); ok && v.NameExpr == nil {
		limit = stableUntil(ctx, b.Body, v)
	}
	for _, record := range calls {
		call, ok := record.Node.(*syntax.FuncCall)
		if !ok || !b.Body.Span().Contains(call.Span()) || call.Args == nil {
			continue
		}
		for _, node := range call.Args.Args {
			a, ok := node.(*syntax.Arg)
			if !ok || a.Unpack {
				continue
			}
			same := semanticquery.NativeLocalValue(ctx, a.Value) == c
			if !same && assigned != nil && astquery.Equivalent(ctx.File, assigned, a.Value) {
				same = a.Value.Span().Start < limit || (limit == 0 && directFirst(b.Body, call))
			}
			if same && !failureRejected(ctx, c, a.Value) {
				ctx.ReportNode(b.Cond, message)
				return
			}
		}
	}
}

func directFirst(body syntax.Stmt, call *syntax.FuncCall) bool {
	if block, ok := body.(*syntax.Block); ok {
		body = block.Stmts[0]
	}
	st, ok := body.(*syntax.ExprStmt)
	return ok && st.Expr == call
}

// stableUntil returns the offset in body from which v may no longer hold the
// value assigned by the condition.
func stableUntil(ctx *analysis.Context, body syntax.Stmt, v *syntax.Variable) uint32 {
	limit := uint32(math.MaxUint32)
	stop := func(n syntax.Node) {
		limit = min(limit, n.Span().Start)
	}
	for _, access := range flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(v), v.Name) {
		if !body.Span().Contains(access.Var.Span()) {
			continue
		}
		use, captured := access.By.(*syntax.ClosureUse)
		if access.Write || access.ElemWrite || (captured && use.ByRef) || aliased(access.Var) {
			stop(access.Var)
		}
	}
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			if x.NameExpr != nil {
				stop(x)
			}
		case *syntax.Eval, *syntax.Include:
			stop(n)
		case *syntax.FuncCall:
			if name := semanticquery.NativeBuiltinName(ctx, x); name == "extract" || name == "parse_str" {
				stop(x)
			}
		}
		return true
	})
	return limit
}

// aliased reports a read that may bind v by reference: a reference
// assignment or array item, or an argument of anything but a plain function
// call (which is itself a consuming call).
func aliased(v *syntax.Variable) bool {
	switch p := v.Parent().(type) {
	case *syntax.Assign:
		return p.ByRef
	case *syntax.ArrayItem:
		return p.ByRef
	case *syntax.Arg:
		_, plain := p.Parent().Parent().(*syntax.FuncCall)
		return !plain
	}
	return false
}

// failureRejected reports whether the tests on the wait result guarding
// value, a use of it, have ruled out the failure value -1.
func failureRejected(ctx *analysis.Context, c *syntax.FuncCall, value syntax.Expr) bool {
	if semanticquery.NativeLabelled(ctx, value) {
		return false
	}
	t := ctx.Types().Native().CallResultAsAt(c, value, types.Ints(types.IntRange{Lo: -1, Hi: math.MaxInt64}))
	return !t.IsUnknown() && !t.HasInt(-1)
}
