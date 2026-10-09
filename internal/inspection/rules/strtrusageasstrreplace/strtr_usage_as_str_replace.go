package strtrusageasstrreplace

import (
	"strings"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// strTrUsageAsStrReplace reports `strtr($s, $from, $to)` with a
// single-character `$from` and suggests `str_replace`.
type strTrUsageAsStrReplace struct{}

func (strTrUsageAsStrReplace) ID() string { return "StrTrUsageAsStrReplace" }
func (strTrUsageAsStrReplace) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strTrUsageAsStrReplace) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, name, ok := astquery.CallName(call)
	if !ok || !strings.EqualFold(name, "strtr") || !ctx.IsGlobalFunctionCall(call, "strtr") { // D1
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 3 { // D2
		return
	}
	from := strtrSingleLiteral(ctx, args[1])            // D3
	if from == nil || !singleCharFrom(ctx.Text(from)) { // D4
		return
	}
	to := strtrSingleLiteral(ctx, args[2]) // D5
	if to == nil || !singleCharFrom(ctx.Text(to)) {
		return
	}
	repl := semanticquery.QualifiedBuiltinFor(ctx, "str_replace", call) + "(" + ctx.Text(args[1]) + ", " + ctx.Text(args[2]) + ", " + ctx.Text(args[0]) + ")"
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' instead.", diagnostic.Fix{
		Title: "Use 'str_replace'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// strtrSingleLiteral resolves e to a single string literal: e itself, or the
// only string literal among its possible values (D3); nil otherwise.
func strtrSingleLiteral(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if astquery.IsStringLiteral(e) {
		return e
	}
	var lit syntax.Expr
	for _, v := range flowquery.PossibleValues(ctx.File, e) {
		if !astquery.IsStringLiteral(v) {
			continue
		}
		if lit != nil {
			return nil // several candidates
		}
		lit = v
	}
	return lit
}

// singleCharFrom implements D4 on the raw source of a string literal.
func singleCharFrom(raw string) bool {
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	var content string
	allowed := `\"$rnt` // escapes standing for one character
	switch {
	case strings.HasPrefix(raw, "<<<"):
		nl := strings.IndexByte(raw, '\n')
		end := strings.LastIndexByte(raw, '\n')
		if nl < 0 || end <= nl {
			return false
		}
		allowed = `\$rnt` // heredoc: \" is two characters
		if strings.Contains(raw[:nl], "'") {
			allowed = "" // nowdoc: no escapes at all
		}
		content = strings.TrimLeft(raw[nl+1:end], " \t")
	case len(raw) >= 2:
		if raw[0] == '\'' {
			allowed = `\'`
		}
		content = raw[1 : len(raw)-1]
	default: // unterminated literal at the end of a broken file
		return false
	}
	switch utf8.RuneCountInString(content) {
	case 1:
		// strtr() works on bytes: a multi-byte character would be mapped
		// byte by byte, unlike str_replace().
		return len(content) == 1 && content != "\n" && content != "\r"
	case 2:
		return content[0] == '\\' && strings.IndexByte(allowed, content[1]) >= 0
	}
	return false
}
