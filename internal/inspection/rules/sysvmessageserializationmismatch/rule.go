// Package sysvmessageserializationmismatch implements the native SysvMessageSerializationMismatch inspection.
package sysvmessageserializationmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use matching serialization flags for the message."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SysvMessageSerializationMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "msg_receive") {
		return
	}
	send := semanticquery.ExpansionDMessage(ctx, c)
	if send == nil {
		return
	}
	a := semanticquery.CallArgument(send.Args, 3, "serialize")
	b := semanticquery.CallArgument(c.Args, 5, "unserialize")
	x, xk := true, true
	if a != nil {
		x, xk = semanticquery.NativeTruth(ctx, a)
	}
	y, yk := true, true
	if b != nil {
		y, yk = semanticquery.NativeTruth(ctx, b)
	}
	// Receiving serialized bytes without decoding is a valid raw transport.
	if xk && yk && !x && y {
		text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(send.Args, 2, "message"))
		// PHP serialized values have either N; or a type/length separator.
		// The full serialization grammar and unknown payloads remain unproven.
		if !known || text == "N;" || strings.Contains(text, ":") {
			return
		}
		for _, operand := range []syntax.Expr{
			a, b,
			semanticquery.CallArgument(send.Args, 1, "message_type"),
			semanticquery.CallArgument(send.Args, 2, "message"),
			semanticquery.CallArgument(c.Args, 1, "desired_message_type"),
		} {
			if variable, ok := syntax.UnwrapParens(operand).(*syntax.Variable); ok && steppedBefore(ctx, variable) {
				return
			}
		}
		ctx.ReportNode(n, message)
	}
}

// steppedBefore reports whether use may observe an increment, a decrement or
// a write through a by-reference closure capture. Local flow facts keep the
// assigned literal across those constructs, so the value is no longer proven.
// A later change still counts when a loop repeats the use.
func steppedBefore(ctx *analysis.Context, use *syntax.Variable) bool {
	for _, access := range flowquery.VarAccesses(ctx.File, syntax.EnclosingFuncLike(use), use.Name) {
		switch by := access.By.(type) {
		case *syntax.IncDec:
		case *syntax.ClosureUse:
			if !by.ByRef {
				continue
			}
		default:
			continue
		}
		if access.Var.Span().Start < use.Span().Start || sharesLoop(use, access.Var) {
			return true
		}
	}
	return false
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
