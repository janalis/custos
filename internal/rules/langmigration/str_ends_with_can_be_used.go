package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// strEndsWithCanBeUsed reports `substr($h, -strlen($n)) === $n` checks that
// str_ends_with() expresses directly.
type strEndsWithCanBeUsed struct{}

func init() { register(strEndsWithCanBeUsed{}) }

// Semantic marks the rule as needing the project index (user functions with
// the matched names may be declared in other files).
func (strEndsWithCanBeUsed) Semantic() {}

func (strEndsWithCanBeUsed) ID() string { return "StrEndsWithCanBeUsed" }

func (strEndsWithCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strEndsWithCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP80 { // E1
		return
	}
	m, ok := matchStrCallCompare(ctx, n.(*syntax.FuncCall), 2, "substr", "mb_substr") // D1-D3
	if !ok {
		return
	}
	neg, ok := m.args[1].(*syntax.Unary) // D4
	if !ok || neg.Op.Kind != syntax.TMinus {
		return
	}
	lenCall, ok := neg.Expr.(*syntax.FuncCall)
	if !ok {
		return
	}
	if name := ctx.GlobalFunctionName(lenCall); name != "strlen" && name != "mb_strlen" {
		return
	}
	largs, ok := util.CallArgValues(lenCall)
	if !ok || len(largs) != 1 {
		return
	}
	if !util.EquivalentFoldNames(ctx.File, m.other, largs[0]) { // D5
		return
	}
	if raw, _, ok := util.QuotedStringRaw(syntax.UnwrapParens(m.other)); ok && raw == "" { // D7
		return // empty needle: the original is false, str_ends_with() is true
	}
	// custos: a needle that may be '' makes the original false and
	// str_ends_with() true, so the fix needs a known non-empty needle.
	reportStrCallReplacementFix(ctx, m, "str_ends_with", m.other, m.cmp.Op.Kind == syntax.TIsNotIdentical, sewNonEmpty(ctx, m.other)) // D6
}

// sewNonEmpty reports whether e is certainly a non-empty string: a
// non-empty literal, a concatenation with such a part, or a variable or
// constant whose every possible value is one.
func sewNonEmpty(ctx *analysis.Context, e syntax.Expr) bool {
	e = syntax.UnwrapParens(e)
	if v, ok := util.QuotedStringValue(e); ok {
		return v != ""
	}
	if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
		return sewNonEmpty(ctx, b.Left) || sewNonEmpty(ctx, b.Right)
	}
	switch e.(type) {
	case *syntax.Variable, *syntax.ClassConstFetch, *syntax.ConstFetch:
		vals, complete := util.PossibleValuesComplete(ctx.File, e)
		if !complete || len(vals) == 0 {
			return false
		}
		for _, v := range vals {
			if v == e || !sewNonEmpty(ctx, v) {
				return false
			}
		}
		return true
	}
	return false
}
