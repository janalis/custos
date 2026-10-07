package performance

import (
	"unicode"
	"unicode/utf8"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// caseInsensitiveStringFunctionsMissUse reports stripos()/strripos()/
// stristr() whose needle has no letters, where the case-sensitive variant
// gives the same result.
type caseInsensitiveStringFunctionsMissUse struct{}

func init() { register(caseInsensitiveStringFunctionsMissUse{}) }

func (caseInsensitiveStringFunctionsMissUse) ID() string {
	return "CaseInsensitiveStringFunctionsMissUse"
}

func (caseInsensitiveStringFunctionsMissUse) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var caseInsensitiveCounterparts = map[string]string{"stripos": "strpos", "strripos": "strrpos", "stristr": "strstr"}

func (caseInsensitiveStringFunctionsMissUse) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	nameNode, _, ok := util.FuncNamePart(call) // D1
	if !ok {
		return
	}
	// D1: the call must reach the global function (any case).
	counterpart, ok := caseInsensitiveCounterparts[ctx.GlobalFunctionName(call)]
	if !ok {
		return
	}
	if c := util.ArgCount(call); c < 2 || c > 3 { // D2
		return
	}
	arg, ok := call.Args.Args[1].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	lit := util.SingleStringLiteral(ctx.File, arg.Value) // D3
	if lit == nil {
		return
	}
	content, ok := util.StringLiteralValue(lit.(*syntax.Literal).Raw)
	if !ok || content == "" || !utf8.ValidString(content) || hasLetter(content) { // D4
		return
	}
	span := util.NamePartSpan(nameNode)
	newName := counterpart
	if q, _, _ := util.CallName(call); q == "" { // a namespaced or imported counterpart would capture a bare call
		newName = util.QualifiedBuiltin(ctx, counterpart, call.Span().Start)
	}
	ctx.ReportSeverity(call.Span(), meta.SeverityInfo, "Needle has no letters; use '"+counterpart+"(...)' instead.", analysis.Fix{
		Title: "Use " + counterpart + "()",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: newName}} },
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
