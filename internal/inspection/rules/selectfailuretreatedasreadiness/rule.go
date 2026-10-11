// Package selectfailuretreatedasreadiness implements the native SelectFailureTreatedAsReadiness inspection.
package selectfailuretreatedasreadiness

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle selection failure before processing ready streams."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SelectFailureTreatedAsReadiness" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || value != 0 {
		return
	}
	call, ok := semanticquery.NativeValue(ctx, b.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "stream_select") || failureExcluded(ctx, b) {
		return
	}
	ctx.ReportNode(b, message)
}

// failureExcluded reports whether a strict false test on the same stored
// result already ruled failure out before the readiness comparison runs.
// Direct calls never qualify: a second call is a different result.
func failureExcluded(ctx *analysis.Context, b *syntax.Binary) bool {
	if _, ok := syntax.UnwrapParens(b.Left).(*syntax.Variable); !ok {
		return false
	}
	if semanticquery.NativeSentinelGuard(ctx, b.Left, "false") {
		return true
	}
	var child syntax.Node = b
	for p := b.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch parent := p.(type) {
		case *syntax.Binary:
			// The right operand of a short-circuit only runs for one outcome of the left.
			and := parent.Op.Kind == syntax.TBooleanAnd || parent.Op.Kind == syntax.TAnd
			or := parent.Op.Kind == syntax.TBooleanOr || parent.Op.Kind == syntax.TOr
			if (and || or) && parent.Right == child && excludesFalse(ctx, parent.Left, and, b.Left) {
				return true
			}
		case *syntax.ElseIf:
			// An elseif is reached only after the leading test and earlier arms failed.
			chain, _ := parent.Parent().(*syntax.If)
			if chain != nil && earlierArmExcludesFalse(ctx, chain, parent, b.Left) {
				return true
			}
		}
	}
	return false
}

// earlierArmExcludesFalse checks the leading test and the arms before arm.
func earlierArmExcludesFalse(ctx *analysis.Context, chain *syntax.If, arm *syntax.ElseIf, input syntax.Expr) bool {
	if excludesFalse(ctx, chain.Cond, false, input) {
		return true
	}
	for _, earlier := range chain.ElseIfs {
		if earlier == arm {
			break
		}
		if excludesFalse(ctx, earlier.Cond, false, input) {
			return true
		}
	}
	return false
}

// excludesFalse reports whether condition evaluating to truth proves that
// input is not false: a strict comparison with false, or a short-circuit
// combination whose every operand is known for that outcome.
func excludesFalse(ctx *analysis.Context, condition syntax.Expr, truth bool, input syntax.Expr) bool {
	b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
	if !ok {
		return false
	}
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return truth && (excludesFalse(ctx, b.Left, true, input) || excludesFalse(ctx, b.Right, true, input))
	case syntax.TBooleanOr, syntax.TOr:
		return !truth && (excludesFalse(ctx, b.Left, false, input) || excludesFalse(ctx, b.Right, false, input))
	case syntax.TIsNotIdentical, syntax.TIsIdentical:
		if truth != (b.Op.Kind == syntax.TIsNotIdentical) {
			return false
		}
		for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
			constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
			if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && sameResult(ctx, input, pair[0]) {
				return true
			}
		}
	}
	return false
}

// sameResult requires both reads to see the same unchanged stored value.
func sameResult(ctx *analysis.Context, a, b syntax.Expr) bool {
	if !astquery.Equivalent(ctx.File, a, b) {
		return false
	}
	av, bv := ctx.Flow().Value(a), ctx.Flow().Value(b)
	return av.Expr == bv.Expr && av.Invalidated == bv.Invalidated
}
