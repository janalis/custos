// Package stringrepeatnegativecount implements the native StringRepeatNegativeCount inspection.
package stringrepeatnegativecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonnegative repetition count."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StringRepeatNegativeCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "str_repeat") {
		return
	}
	times := semanticquery.CallArgument(c.Args, 1, "times")
	v, k := semanticquery.NativeInt(ctx, times)
	if !k || v >= 0 {
		return
	}
	if raisedBefore(ctx, times, maxCopies) {
		return
	}
	ctx.ReportNode(c, message)
}

// maxCopies bounds the variable-to-variable copies followed behind the count.
const maxCopies = 8

// raisedBefore reports whether e is a variable that may observe an increment,
// a write through a by-reference closure capture or an extract() call. Local
// flow facts keep the assigned literal across those constructs, so the count
// is no longer a proven negative. Plain copies from another variable are
// followed. A decrement keeps a negative count negative, and a nested arrow
// function only steps its own copy (its occurrences are plain reads here). A
// later change still counts when a loop repeats the use.
func raisedBefore(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	use, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return false
	}
	if depth == 0 {
		return true
	}
	scope := syntax.EnclosingFuncLike(use)
	if extractedBefore(ctx, scope, use) {
		return true
	}
	for _, access := range flowquery.VarAccesses(ctx.File, scope, use.Name) {
		if access.Var.Span().Start >= use.Span().Start && !sharesLoop(use, access.Var) {
			continue
		}
		switch by := access.By.(type) {
		case *syntax.IncDec:
			if by.Op.Kind != syntax.TDec {
				return true
			}
		case *syntax.ClosureUse:
			if by.ByRef {
				return true
			}
		case *syntax.Assign:
			if access.Write && by.Var == syntax.Expr(access.Var) && raisedBefore(ctx, by.Value, depth-1) {
				return true
			}
		}
	}
	return false
}

// extractedBefore reports whether an extract() call of scope may rebind
// variables before use is read.
func extractedBefore(ctx *analysis.Context, scope syntax.Node, use *syntax.Variable) bool {
	found := false
	visit := func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok && !found && (call.Span().Start < use.Span().Start || sharesLoop(use, call)) {
			found = semanticquery.NativeBuiltin(ctx, call, "extract")
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// sharesLoop reports whether a loop around use also contains other.
func sharesLoop(use, other syntax.Node) bool {
	for p := use.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if span := p.Span(); span.Start <= other.Span().Start && other.Span().End <= span.End {
				return true
			}
		}
	}
	return false
}
