// Package hashfilefailureusedasdigest implements HashFileFailureUsedAsDigest.
package hashfilefailureusedasdigest

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject file hashing failure before using the digest."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HashFileFailureUsedAsDigest" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, input := range inputs(ctx, n) {
		if guarded(ctx, input) {
			return
		}
	}
	semanticquery.CheckHashFileFailureUsedAsDigest(ctx, n, message)
}

// inputs lists the operands the shared check treats as a consumed digest.
func inputs(ctx *analysis.Context, n syntax.Node) []syntax.Expr {
	var out []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		if semanticquery.NativeBuiltin(ctx, c, "file_put_contents") {
			out = append(out, semanticquery.CallArgument(c.Args, 1, "data"))
		}
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsNotIdentical {
			return nil
		}
		if _, ok := semanticquery.NativeString(ctx, c.Right); ok {
			out = append(out, c.Left)
		}
		if _, ok := semanticquery.NativeString(ctx, c.Left); ok {
			out = append(out, c.Right)
		}
	}
	return out
}

// guarded proves that false cannot reach input: either the flow facts already
// exclude it, or a truthiness, short-circuit, ternary or early-exit guard on
// the same reaching value dominates the use.
func guarded(ctx *analysis.Context, input syntax.Expr) bool {
	if ctx.Flow().Excludes(input, "false") {
		return true
	}
	if _, ok := syntax.UnwrapParens(input).(*syntax.Variable); !ok {
		return false
	}
	for p := input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			if x.Right != nil && x.Right.Span().Contains(input.Span()) && (isAnd(x) && rejects(ctx, input, x.Left, true) || isOr(x) && rejects(ctx, input, x.Left, false)) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(input.Span()) && rejects(ctx, input, x.Cond, true) {
				return true
			}
			if x.Else != nil && x.Else.Span().Contains(input.Span()) && rejects(ctx, input, x.Cond, false) {
				return true
			}
		case *syntax.If:
			if x.Body != nil && x.Body.Span().Contains(input.Span()) && rejects(ctx, input, x.Cond, true) {
				return true
			}
			if x.Else != nil && x.Else.Span().Contains(input.Span()) && rejects(ctx, input, x.Cond, false) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && exitedEarlier(ctx, input, st) {
			return true
		}
	}
	return false
}

// exitedEarlier finds an earlier sibling `if` without alternatives whose body
// always leaves the path and whose condition holds whenever input is false.
func exitedEarlier(ctx *analysis.Context, input syntax.Expr, st syntax.Stmt) bool {
	list, i, _ := astquery.StmtList(ctx.File, st)
	for _, previous := range list[:max(i, 0)] {
		guard, ok := previous.(*syntax.If)
		if ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && rejects(ctx, input, guard.Cond, false) {
			return true
		}
	}
	return false
}

// rejects reports whether cond evaluating to truth implies input is not false.
func rejects(ctx *analysis.Context, input, cond syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Variable:
		return truth && sameValue(ctx, input, c)
	case *syntax.Empty:
		return !truth && sameValue(ctx, input, c.Expr)
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && rejects(ctx, input, c.Expr, !truth)
	case *syntax.Binary:
		switch {
		case isAnd(c) == truth && (isAnd(c) || isOr(c)):
			// `a && b` held, or `a || b` failed: both operands are decided.
			return rejects(ctx, input, c.Left, truth) || rejects(ctx, input, c.Right, truth)
		case isAnd(c) || isOr(c):
			return rejects(ctx, input, c.Left, truth) && rejects(ctx, input, c.Right, truth)
		case c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual:
			return truth && comparesFalse(ctx, input, c)
		case c.Op.Kind == syntax.TIsIdentical || c.Op.Kind == syntax.TIsEqual:
			return !truth && comparesFalse(ctx, input, c)
		}
	}
	return false
}

func isAnd(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func isOr(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}

func comparesFalse(ctx *analysis.Context, input syntax.Expr, b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && sameValue(ctx, input, pair[0]) {
			return true
		}
	}
	return false
}

// sameValue ties a guard operand to the value reaching input: a reassignment
// or invalidation between the two creates a different value.
func sameValue(ctx *analysis.Context, input, other syntax.Expr) bool {
	if other == nil || !astquery.Equivalent(ctx.File, syntax.UnwrapParens(input), syntax.UnwrapParens(other)) {
		return false
	}
	a, b := ctx.Flow().Value(input), ctx.Flow().Value(other)
	return a.Expr == b.Expr && a.Invalidated == b.Invalidated
}
