package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// stringCaseManipulation reports position searches whose haystack/needle is
// case-converted only to make the search case-insensitive.
type stringCaseManipulation struct{}

func init() { register(stringCaseManipulation{}) }

func (stringCaseManipulation) ID() string { return "StringCaseManipulation" }

func (stringCaseManipulation) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var caseInsensitiveSearch = map[string]string{
	"strpos":     "stripos",
	"mb_strpos":  "mb_stripos",
	"strrpos":    "strripos",
	"mb_strrpos": "mb_strripos",
}

func isCaseConversion(name string) bool {
	switch name {
	case "strtolower", "mb_strtolower", "strtoupper", "mb_strtoupper":
		return true
	}
	return false
}

func (stringCaseManipulation) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	variant, ok := caseInsensitiveSearch[ctx.GlobalFunctionName(call)] // D1
	if !ok {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 2 { // D2
		return
	}
	var parts [2]string
	found := false
	for i, a := range args { // D3
		parts[i] = ctx.Text(a)
		inner, ok := a.(*syntax.FuncCall)
		if !ok {
			continue
		}
		if !isCaseConversion(ctx.GlobalFunctionName(inner)) {
			continue
		}
		iargs, ok := util.CallArgValues(inner)
		if !ok || len(iargs) != 1 {
			continue
		}
		parts[i] = ctx.Text(iargs[0])
		found = true
	}
	if !found { // D4
		return
	}
	// A namespaced function of that name would capture a bare call.
	repl := util.QualifiedBuiltin(ctx, variant, call.Span().Start) + "(" + parts[0] + ", " + parts[1] + ")"
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' instead of changing the case.", analysis.Fix{
		Title: "Use '" + variant + "'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}
