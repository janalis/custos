package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// arrayUniqueCanBeUsed reports array_keys()/count() over array_count_values(),
// which array_unique() expresses directly since PHP 7.2.
type arrayUniqueCanBeUsed struct{}

func init() { register(arrayUniqueCanBeUsed{}) }

func (arrayUniqueCanBeUsed) ID() string { return "ArrayUniqueCanBeUsed" }

func (arrayUniqueCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (arrayUniqueCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP72 { // D1
		return
	}
	inner, ok := util.IsFuncNamedFold(n, "array_count_values") // D2
	if !ok || !ctx.IsGlobalFunctionCall(inner, "array_count_values") {
		return
	}
	args, ok := util.CallArgValues(inner)
	if !ok || len(args) != 1 {
		return
	}
	outer := util.ParentFuncCall(inner) // D3
	if outer == nil {
		return
	}
	// Divergence (spec): only when the inner call is the outer call's sole argument.
	if oargs, ok := util.CallArgValues(outer); !ok || len(oargs) != 1 {
		return
	}
	a := ctx.Text(args[0])
	// F1: each call keeps its leading `\` and gets one when a bare name
	// would not reach the global function.
	unique := util.QualifiedBuiltinFor(ctx, "array_unique", inner)
	var repl, fixed string
	switch ctx.GlobalFunctionName(outer) {
	case "array_keys":
		repl = "array_values(array_unique(" + a + "))"
		fixed = util.QualifiedBuiltinFor(ctx, "array_values", outer) + "(" + unique + "(" + a + "))"
	case "count":
		repl = "count(array_unique(" + a + "))"
		fixed = util.QualifiedBuiltinFor(ctx, "count", outer) + "(" + unique + "(" + a + "))"
	default:
		return
	}
	span := outer.Span()
	ctx.Report(span, "Use '"+repl+"' instead (array_unique() is fast since PHP 7.2).", analysis.Fix{
		Title: "Use array_unique()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: fixed}}
		},
	})
}
