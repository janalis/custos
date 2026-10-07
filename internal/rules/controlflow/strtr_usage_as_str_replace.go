package controlflow

import (
	"strings"
	"unicode/utf8"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// strTrUsageAsStrReplace reports `strtr($s, $from, $to)` with a
// single-character `$from` and suggests `str_replace`.
type strTrUsageAsStrReplace struct{}

func init() { register(strTrUsageAsStrReplace{}) }

func (strTrUsageAsStrReplace) ID() string { return "StrTrUsageAsStrReplace" }

func (strTrUsageAsStrReplace) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strTrUsageAsStrReplace) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, name, ok := util.CallName(call)
	if !ok || !strings.EqualFold(name, "strtr") || !ctx.IsGlobalFunctionCall(call, "strtr") { // D1
		return
	}
	args, ok := util.CallArgValues(call)
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
	repl := util.QualifiedBuiltinFor(ctx, "str_replace", call) + "(" + ctx.Text(args[1]) + ", " + ctx.Text(args[2]) + ", " + ctx.Text(args[0]) + ")"
	span := call.Span()
	ctx.Report(span, "Use '"+repl+"' instead.", analysis.Fix{
		Title: "Use 'str_replace'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// strtrSingleLiteral resolves e to a single string literal: e itself, or the
// only string literal among its possible values (D3); nil otherwise.
func strtrSingleLiteral(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	if util.IsStringLiteral(e) {
		return e
	}
	var lit syntax.Expr
	for _, v := range util.PossibleValues(ctx.File, e) {
		if !util.IsStringLiteral(v) {
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
	single := false
	switch {
	case strings.HasPrefix(raw, "<<<"):
		nl := strings.IndexByte(raw, '\n')
		end := strings.LastIndexByte(raw, '\n')
		if nl < 0 || end <= nl {
			return false
		}
		single = strings.Contains(raw[:nl], "'")
		content = strings.TrimLeft(raw[nl+1:end], " \t")
	case len(raw) >= 2:
		single = raw[0] == '\''
		content = raw[1 : len(raw)-1]
	default:
		return false
	}
	switch utf8.RuneCountInString(content) {
	case 1:
		// strtr() works on bytes: a multi-byte character would be mapped
		// byte by byte, unlike str_replace().
		return len(content) == 1 && content != "\n" && content != "\r"
	case 2:
		if content[0] != '\\' {
			return false
		}
		allowed := `\"$rnt`
		if single {
			allowed = `\'`
		}
		return strings.IndexByte(allowed, content[1]) >= 0
	}
	return false
}
