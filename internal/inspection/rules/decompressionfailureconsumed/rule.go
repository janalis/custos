// Package decompressionfailureconsumed implements the native DecompressionFailureConsumed inspection.
package decompressionfailureconsumed

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check decompression before consuming data."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DecompressionFailureConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "strlen", "strtoupper":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "string")
	origin, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok {
		return
	}
	switch semanticquery.NativeBuiltinName(ctx, origin) {
	case "gzdecode", "gzuncompress", "gzinflate", "inflate_add":
	default:
		return
	}
	encoder, known := semanticquery.NativeValue(ctx, semanticquery.CallArgument(origin.Args, 0, "data")).(*syntax.FuncCall)
	if known && semanticquery.CallArgument(encoder.Args, 2, "encoding") == nil {
		expected := map[string]string{"gzdecode": "gzencode", "gzuncompress": "gzcompress", "gzinflate": "gzdeflate"}[semanticquery.NativeBuiltinName(ctx, origin)]
		if expected != "" && semanticquery.NativeBuiltin(ctx, encoder, expected) {
			level := semanticquery.CallArgument(encoder.Args, 1, "level")
			if level == nil {
				return
			}
			if n, ok := semanticquery.NativeInt(ctx, level); ok && n >= -1 && n <= 9 {
				return
			}
		}
	}

	if !ctx.Flow().Excludes(input, "false") && !guarded(ctx, input) {
		ctx.ReportNode(c, message)
	}
}

// guarded proves that a strict false test of the same value dominates input:
// a branch, loop or ternary arm selected by the test, the right operand of a
// short-circuit operator, or the code after a terminating failure branch.
func guarded(ctx *analysis.Context, input syntax.Expr) bool {
	in := func(n syntax.Node) bool { return n != nil && n.Span().Contains(input.Span()) }
	for p := input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		var cond syntax.Expr
		truth := true
		switch x := p.(type) {
		case *syntax.Ternary:
			if !in(x.Cond) {
				cond, truth = x.Cond, !in(x.Else)
			}
		case *syntax.Binary:
			switch x.Op.Kind {
			case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
				if in(x.Right) {
					cond, truth = x.Left, x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd
				}
			}
		case *syntax.If:
			// A later arm of the chain runs only once every earlier test failed.
			if in(x.Cond) {
				break
			}
			cond, truth = x.Cond, in(x.Body)
			for _, arm := range x.ElseIfs {
				if truth || in(arm) {
					break
				}
				if excludesFalse(ctx, input, arm.Cond, false) {
					return true
				}
			}
		case *syntax.ElseIf:
			if in(x.Body) {
				cond = x.Cond
			}
		case *syntax.While:
			if in(x.Body) {
				cond = x.Cond
			}
		}
		if cond != nil && excludesFalse(ctx, input, cond, truth) {
			return true
		}
		if st, ok := p.(syntax.Stmt); ok && exitedEarlier(ctx, input, st) {
			return true
		}
	}
	return false
}

// exitedEarlier finds an earlier sibling `if` without alternatives whose body
// never falls through and whose test fails only when input is not false.
// Statements in between are harmless: sameValue rejects a rewritten value.
func exitedEarlier(ctx *analysis.Context, input syntax.Expr, st syntax.Stmt) bool {
	// Outside a statement list (unbraced body) the index is negative.
	list, index, _ := astquery.StmtList(ctx.File, st)
	for i := 0; i < index; i++ {
		guard, ok := list[i].(*syntax.If)
		if ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && excludesFalse(ctx, input, guard.Cond, false) {
			return true
		}
	}
	return false
}

// excludesFalse reports whether cond evaluating to truth implies a strict
// comparison established that input is not false. Truthiness proves nothing.
func excludesFalse(ctx *analysis.Context, input, cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && excludesFalse(ctx, input, x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			// A true conjunction evaluated both operands.
			return truth && (excludesFalse(ctx, input, x.Left, true) || excludesFalse(ctx, input, x.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (excludesFalse(ctx, input, x.Left, false) || excludesFalse(ctx, input, x.Right, false))
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			if truth != (x.Op.Kind == syntax.TIsNotIdentical) {
				return false
			}
			for _, pair := range [][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && sameValue(ctx, input, pair[0]) {
					return true
				}
			}
		}
	}
	return false
}

// sameValue ties a tested operand to the producer reaching input. The operand
// is the variable itself or the assignment that stored the reaching result;
// a reassignment or a repeated call is a different value.
func sameValue(ctx *analysis.Context, input, tested syntax.Expr) bool {
	reaching := ctx.Flow().Value(input)
	tested = syntax.UnwrapParens(tested)
	if a, ok := tested.(*syntax.Assign); ok {
		return !a.ByRef && a.Op.Kind == syntax.TEqual && astquery.Equivalent(ctx.File, input, a.Var) && reaching.Expr != nil && syntax.UnwrapParens(reaching.Expr) == syntax.UnwrapParens(a.Value)
	}
	if !astquery.Equivalent(ctx.File, input, tested) {
		return false
	}
	seen := ctx.Flow().Value(tested)
	return reaching.Expr != nil && reaching.Expr == seen.Expr && reaching.Invalidated == seen.Invalidated
}
