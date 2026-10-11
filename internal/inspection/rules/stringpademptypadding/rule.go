// Package stringpademptypadding implements the native StringPadEmptyPadding inspection.
package stringpademptypadding

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonempty padding string."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StringPadEmptyPadding" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "str_pad") {
		return
	}
	padArg := semanticquery.CallArgument(c.Args, 2, "pad_string")
	lengthArg := semanticquery.CallArgument(c.Args, 1, "length")
	inputArg := semanticquery.CallArgument(c.Args, 0, "string")
	pad, pk := semanticquery.NativeString(ctx, padArg)
	length, lk := semanticquery.NativeInt(ctx, lengthArg)
	input, ik := semanticquery.NativeString(ctx, inputArg)
	if !pk || pad != "" || !lk || !ik || length <= int64(len(input)) {
		return
	}
	for _, arg := range []syntax.Expr{padArg, lengthArg, inputArg} {
		if stale(ctx, arg, maxCopies) {
			return
		}
	}
	ctx.ReportNode(c, message)
}

// maxCopies bounds the variable-to-variable copies followed behind an argument.
const maxCopies = 8

// stale reports whether e is a variable whose assigned literal may no longer
// be its value where it is read: local flow facts keep that literal across
// an increment, a decrement, a write through a by-reference closure capture
// and extract(). Plain copies from another variable are followed.
func stale(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	use, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok || use.NameExpr != nil {
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
		if !observed(use, access.Var) {
			continue
		}
		switch by := access.By.(type) {
		case *syntax.IncDec:
			return true
		case *syntax.ClosureUse:
			if by.ByRef {
				return true
			}
		case *syntax.Assign:
			if access.Write && by.Var == syntax.Expr(access.Var) && stale(ctx, by.Value, depth-1) {
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
		if call, ok := n.(*syntax.FuncCall); ok && !found && observed(use, call) {
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

// observed reports whether other precedes use in source order, or a loop
// around use contains it so a later iteration reads its effect.
func observed(use, other syntax.Node) bool {
	if other.Span().Start < use.Span().Start {
		return true
	}
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
