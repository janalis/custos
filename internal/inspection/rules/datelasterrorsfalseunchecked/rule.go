// Package datelasterrorsfalseunchecked implements the DateLastErrorsFalseUnchecked inspection.
package datelasterrorsfalseunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateLastErrorsFalseUnchecked" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }

const message = "Handle the false result before reading parsing diagnostics."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP82 {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	origin := semanticquery.NativeValue(ctx, fetch.Var)
	if ctx.Flow().Excludes(fetch.Var, "false") {
		return
	}
	if semanticquery.NativeDateStatic(ctx, origin, "getLastErrors") == nil {
		return
	}
	if silentProbe(fetch) || (guard{ctx, fetch.Var}).dominates(fetch) {
		return
	}
	ctx.ReportNode(fetch, message)
}
func (rule) Semantic() {}

// silentProbe recognises the contexts where PHP evaluates an offset of false
// without a diagnostic: isset, empty and the tested side of `??`.
func silentProbe(fetch *syntax.ArrayDimFetch) bool {
	for child, parent := syntax.Node(fetch), fetch.Parent(); ; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != child {
				return false
			}
		case *syntax.Paren:
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			return p.Op.Kind == syntax.TCoalesce && p.Left == child
		default:
			return false
		}
	}
}

// guard proves, from the conditions that must have held to reach a use, that
// the checked value is not false there.
type guard struct {
	ctx    *analysis.Context
	target syntax.Expr
}

// dominates walks outward from the use to an earlier sibling statement that
// leaves the path on false. The shared flow analysis, asked before this
// guard, answers every guard of a plain variable; the walk is left with a
// value held in a property.
func (g guard) dominates(use syntax.Node) bool {
	for parent := use.Parent(); parent != nil && !syntax.IsVariableScope(parent); parent = parent.Parent() {
		if st, ok := parent.(syntax.Stmt); ok && g.exitedEarlier(st) {
			return true
		}
	}
	return false
}

// exitedEarlier finds a preceding sibling `if` without alternatives whose
// body never falls through and whose condition holds whenever the value is false.
func (g guard) exitedEarlier(st syntax.Stmt) bool {
	list, index, ok := astquery.StmtList(g.ctx.File, st)
	if !ok {
		return false
	}
	for _, previous := range list[:index] {
		exit, ok := previous.(*syntax.If)
		if ok && exit.Else == nil && len(exit.ElseIfs) == 0 && syntax.Terminates(exit.Body) && g.implies(exit.Cond, false) {
			return true
		}
	}
	return false
}

// implies reports whether the condition evaluating to truth proves the value
// is not false. Unknown shapes prove nothing.
func (g guard) implies(condition syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(condition).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.implies(x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (g.implies(x.Left, true) || g.implies(x.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (g.implies(x.Left, false) || g.implies(x.Right, false))
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.comparesFalse(x)
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.comparesFalse(x)
		}
		return false
	case *syntax.Empty:
		return !truth && g.same(x.Expr)
	case *syntax.Isset:
		if !truth {
			return false
		}
		for _, probed := range x.Vars {
			if offset, ok := syntax.UnwrapParens(probed).(*syntax.ArrayDimFetch); ok && g.same(offset.Var) {
				return true
			}
		}
		return false
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, x, "is_array") && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
	default:
		return truth && g.same(x)
	}
}

func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if value, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok && !value && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same ties a tested expression to the value being read: identical text and
// the same reaching producer, so a reassignment between test and use breaks it.
// A tested plain assignment stands for the value it stores.
func (g guard) same(tested syntax.Expr) bool {
	tested = syntax.UnwrapParens(tested)
	if stored, ok := tested.(*syntax.Assign); ok && stored.Op.Kind == syntax.TEqual && !stored.ByRef {
		return astquery.Equivalent(g.ctx.File, stored.Var, g.target) && g.ctx.Flow().Value(g.target).Expr == syntax.UnwrapParens(stored.Value)
	}
	if tested == nil || !astquery.Equivalent(g.ctx.File, tested, g.target) {
		return false
	}
	a, b := g.ctx.Flow().Value(tested), g.ctx.Flow().Value(g.target)
	return a.Expr == b.Expr && a.Invalidated == b.Invalidated
}
