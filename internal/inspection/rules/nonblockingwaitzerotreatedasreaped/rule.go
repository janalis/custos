// Package nonblockingwaitzerotreatedasreaped implements the native NonblockingWaitZeroTreatedAsReaped inspection.
package nonblockingwaitzerotreatedasreaped

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Require a positive PID from nonblocking wait."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonblockingWaitZeroTreatedAsReaped" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsNotIdentical && b.Op.Kind != syntax.TIsNotEqual && b.Op.Kind != syntax.TIsGreaterOrEqual {
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || !((b.Op.Kind == syntax.TIsGreaterOrEqual && value == 0) || (b.Op.Kind != syntax.TIsGreaterOrEqual && value == -1)) {
		return
	}
	wait, ok := semanticquery.NativeValue(ctx, b.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, wait, "pcntl_waitpid") {
		return
	}
	flag := semanticquery.CallArgument(wait.Args, 2, "flags")
	if semanticquery.NamesGlobalConstant(ctx, flag, "WNOHANG") && !zeroHandled(ctx, b, wait) {
		ctx.ReportNode(b, message)
	}
}

// zeroHandled reports whether the same wait result is told apart from zero
// before the comparison, in its condition or in the statement it governs.
// Later tests cannot protect the guarded code and are ignored.
func zeroHandled(ctx *analysis.Context, b *syntax.Binary, wait *syntax.FuncCall) bool {
	var stmt syntax.Node = b
	for stmt.Parent() != nil {
		if _, ok := stmt.(syntax.Stmt); ok {
			break
		}
		stmt = stmt.Parent()
	}
	end := stmt.Span().End
	same := func(e syntax.Expr) bool {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		return ok && semanticquery.NativeValue(ctx, v) == syntax.Expr(wait)
	}
	found := false
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if found || n.Span().Start >= end {
			return false
		}
		switch x := n.(type) {
		case *syntax.Binary:
			found = comparesZero(ctx, x, same)
		case *syntax.Variable:
			found = truthTested(x) && same(x)
		}
		return !found
	})
	return found
}

var mirrored = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TLess: syntax.TGreater, syntax.TGreater: syntax.TLess,
	syntax.TIsSmallerOrEqual: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TIsSmallerOrEqual,
}

// comparesZero accepts comparisons whose outcome differs between zero and
// every positive PID, with the wait result on either side.
func comparesZero(ctx *analysis.Context, b *syntax.Binary, same func(syntax.Expr) bool) bool {
	op, other := b.Op.Kind, b.Right
	if !same(b.Left) {
		if !same(b.Right) {
			return false
		}
		other = b.Left
		if m, ok := mirrored[op]; ok {
			op = m
		}
	}
	value, ok := semanticquery.NativeInt(ctx, other)
	if !ok {
		return false
	}
	switch op {
	case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual, syntax.TGreater, syntax.TIsSmallerOrEqual:
		return value == 0
	case syntax.TIsGreaterOrEqual, syntax.TLess:
		return value == 1
	}
	return false
}

// truthTested recognizes a variable used as a boolean or as the subject of a
// switch or match: zero is falsy while -1 and every PID are truthy.
func truthTested(v *syntax.Variable) bool {
	parent, child := astquery.ParentSkipParens(v)
	switch p := parent.(type) {
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		return p.Op.Kind == syntax.TBooleanAnd || p.Op.Kind == syntax.TBooleanOr
	case *syntax.If:
		return p.Cond == child
	case *syntax.ElseIf:
		return p.Cond == child
	case *syntax.Ternary:
		return p.Cond == child
	case *syntax.Switch:
		return p.Cond == child
	case *syntax.Match:
		return p.Cond == child
	}
	return false
}
