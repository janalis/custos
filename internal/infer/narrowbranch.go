package infer

import (
	"cmp"
	"slices"

	"custos/internal/syntax"
	"custos/internal/types"
)

// Multi-way branches: elseif chains (narrow.go), match arms and switch
// cases. A branch is reached when its own condition holds and every
// condition tested before it failed; `match` compares strictly (so
// `match (true) { is_string($x) => … }` narrows like `if`), `switch`
// loosely.

// maxBranchScan caps the earlier branches (elseifs, match arms, switch
// cases) whose failed conditions narrow a read; beyond it they are
// ignored (sound: less narrowing), so a 10k-arm match stays linear.
const maxBranchScan = 256

// nodeIndex returns the position of n in list (source order), len(list)
// when absent.
func nodeIndex[T syntax.Node](list []T, n syntax.Node) int {
	start := n.Span().Start
	i, ok := slices.BinarySearchFunc(list, start, func(x T, s uint32) int { return cmp.Compare(x.Span().Start, s) })
	if ok && syntax.Node(list[i]) == n {
		return i
	}
	return len(list)
}

// branchCmp is the comparison `subject OP value` a match arm or switch
// case performs (a synthetic node: only condition narrowing reads it).
func branchCmp(subject, value syntax.Expr, op syntax.TokenKind) *syntax.Binary {
	return &syntax.Binary{Left: subject, Op: syntax.TokenRef{Kind: op}, Right: value}
}

// anyCmp is `subject OP v1 || subject OP v2 || …`.
func anyCmp(subject syntax.Expr, values []syntax.Expr, op syntax.TokenKind) syntax.Expr {
	var or syntax.Expr = branchCmp(subject, values[0], op)
	for _, v := range values[1:] {
		or = &syntax.Binary{Left: or, Op: syntax.TokenRef{Kind: syntax.TBooleanOr}, Right: branchCmp(subject, v, op)}
	}
	return or
}

// matchArm narrows the type t of use, inside arm (in its body or one of
// its conditions, child), by the arms tested before.
func (e *Env) matchArm(use syntax.Expr, scope syntax.Node, t types.Type, arm *syntax.MatchArm, child syntax.Node, key string, after uint32) types.Type {
	m := arm.Parent().(*syntax.Match) // the parser always gives a condition (BadExpr at worst)
	i := nodeIndex(m.Arms, arm)
	if i == len(m.Arms) {
		return t
	}
	fail := func(t types.Type, v syntax.Expr) types.Type {
		if v.Span().End < after {
			return t
		}
		return e.condAt(use, scope, t, branchCmp(m.Cond, v, syntax.TIsIdentical), key, false, v.Span().End)
	}
	failArms := func(t types.Type, arms []*syntax.MatchArm) types.Type {
		for _, a := range arms {
			if a != arm {
				for _, v := range a.Conds {
					t = fail(t, v)
				}
			}
		}
		return t
	}
	if child != syntax.Node(arm.Body) {
		// A condition of arm: the arms before and the conditions before
		// it in arm did not match.
		if i < maxBranchScan {
			t = failArms(t, m.Arms[:i])
		}
		for j, v := range arm.Conds {
			if syntax.Node(v) == child || j >= maxBranchScan {
				break
			}
			t = fail(t, v)
		}
		return t
	}
	if arm.Conds == nil {
		// default: no other arm matched.
		if len(m.Arms) > maxBranchScan {
			return t
		}
		return failArms(t, m.Arms)
	}
	if i < maxBranchScan {
		t = failArms(t, m.Arms[:i])
	}
	last := arm.Conds[len(arm.Conds)-1]
	if len(arm.Conds) > maxBranchScan || arm.Conds[0].Span().End < after {
		return t
	}
	return e.condAt(use, scope, t, anyCmp(m.Cond, arm.Conds, syntax.TIsIdentical), key, true, last.Span().End)
}

// switchCase narrows the type t of use, in the statements of case c, by
// the case's own value (and the empty cases falling into it) and the
// cases tested before; nothing when c may be entered by falling through
// a non-empty case.
func (e *Env) switchCase(use syntax.Expr, scope syntax.Node, t types.Type, c *syntax.Case, key string, after uint32) types.Type {
	sw := c.Parent().(*syntax.Switch)
	i := nodeIndex(sw.Cases, c)
	if i == len(sw.Cases) {
		return t
	}
	fail := func(t types.Type, cases []*syntax.Case) types.Type {
		for _, o := range cases {
			if o.Cond != nil && o.Cond.Span().End >= after {
				t = e.condAt(use, scope, t, branchCmp(sw.Cond, o.Cond, syntax.TIsEqual), key, false, o.Cond.Span().End)
			}
		}
		return t
	}
	if c.Cond == nil {
		// default: no case matched (and not entered by falling through).
		if len(sw.Cases) > maxBranchScan || (i > 0 && !caseExits(sw.Cases[i-1])) {
			return t
		}
		return fail(t, sw.Cases)
	}
	values := []syntax.Expr{c.Cond}
	j := i - 1
	for ; j >= 0 && len(sw.Cases[j].Stmts) == 0; j-- {
		if sw.Cases[j].Cond == nil || len(values) >= maxBranchScan {
			return t
		}
		values = append(values, sw.Cases[j].Cond)
	}
	if j >= 0 && !caseExits(sw.Cases[j]) {
		return t
	}
	if j+1 <= maxBranchScan {
		t = fail(t, sw.Cases[:j+1])
	}
	slices.Reverse(values)
	if values[0].Span().End < after {
		return t
	}
	return e.condAt(use, scope, t, anyCmp(sw.Cond, values, syntax.TIsEqual), key, true, c.Cond.Span().End)
}
