// Package realpathfailureusedaspath implements the native RealpathFailureUsedAsPath inspection.
package realpathfailureusedaspath

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Reject canonicalization failure before joining paths."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RealpathFailureUsedAsPath" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.ExpansionDPathCall(ctx, c) {
		return
	}
	b, ok := semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "filename")).(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TDot {
		return
	}
	for _, input := range []syntax.Expr{b.Left, b.Right} {
		if !semanticquery.ExpansionDUnaliased(ctx, input, c) {
			continue
		}
		source, known := semanticquery.NativeLocalBuiltinResult(ctx, input, "false", "realpath").(*syntax.FuncCall)
		if !known || !semanticquery.NativeBuiltin(ctx, source, "realpath") || ctx.Flow().Excludes(input, "false") {
			continue
		}
		// The join may sit in an earlier assignment; a check between it and
		// the consuming call still keeps the failure away from the path API.
		if !semanticquery.NativeFailureFlowOnly(ctx, input) && !ctx.Flow().LabelledBefore(input, c) &&
			(semanticquery.NativeNarrowedExcludes(ctx, input, types.Unknown, "false") || narrowedExcludesAt(ctx, input, c, "false")) {
			continue
		}
		ctx.ReportNode(c, message)
		return
	}
}

// narrowedExcludesAt reports whether the conditions guarding at, a
// later expression that read x still reaches with the same value (a call
// consuming what was computed from x), prove x does not hold atom there. Only
// a plain local that is not static is followed.
func narrowedExcludesAt(ctx *analysis.Context, x, at syntax.Expr, atom string) bool {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	if !ok || semanticquery.NativeStaticLocal(ctx, v) || semanticquery.NativeLabelled(ctx, at) {
		return false
	}
	env := ctx.Types().Native()
	if !env.ReachesUnchanged(v, at) {
		return false
	}
	t := env.TypeOfVarAt(v.Name, at)
	return !t.IsUnknown() && !semanticquery.MayHoldAtom(t, atom)
}
