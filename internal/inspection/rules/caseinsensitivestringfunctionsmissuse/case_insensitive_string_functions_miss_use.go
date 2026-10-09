package caseinsensitivestringfunctionsmissuse

import (
	"unicode"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// caseInsensitiveStringFunctionsMissUse reports stripos()/strripos()/
// stristr() whose needle has no letters, where the case-sensitive variant
// gives the same result.
type caseInsensitiveStringFunctionsMissUse struct{}

func (caseInsensitiveStringFunctionsMissUse) ID() string {
	return "CaseInsensitiveStringFunctionsMissUse"
}

func (caseInsensitiveStringFunctionsMissUse) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var caseInsensitiveCounterparts = map[string]string{"stripos": "strpos", "strripos": "strrpos", "stristr": "strstr"}

func (caseInsensitiveStringFunctionsMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	nameNode, _, ok := astquery.FuncNamePart(call) // D1
	if !ok {
		return
	}
	// D1: the call must reach the global function (any case).
	counterpart, ok := caseInsensitiveCounterparts[ctx.GlobalFunctionName(call)]
	if !ok {
		return
	}
	if c := astquery.ArgCount(call); c < 2 || c > 3 { // D2
		return
	}
	arg, ok := call.Args.Args[1].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	lit := flowquery.SingleStringLiteral(ctx.File, arg.Value) // D3
	if lit == nil {
		return
	}
	content, ok := astquery.StringLiteralValue(lit.(*syntax.Literal).Raw)
	if !ok || content == "" || !utf8.ValidString(content) || hasLetter(content) { // D4
		return
	}
	span := astquery.NamePartSpan(nameNode)
	newName := counterpart
	if q, _, _ := astquery.CallName(call); q == "" { // a namespaced or imported counterpart would capture a bare call
		newName = semanticquery.QualifiedBuiltin(ctx, counterpart, call.Span().Start)
	}
	ctx.ReportSeverity(call.Span(), diagnostic.SeverityInfo, "Needle has no letters; use '"+counterpart+"(...)' instead.", diagnostic.Fix{
		Title: "Use " + counterpart + "()",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: newName}} },
	})
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
