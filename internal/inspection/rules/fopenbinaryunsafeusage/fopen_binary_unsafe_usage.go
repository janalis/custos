package fopenbinaryunsafeusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// fopenBinaryUnsafeUsage reports fopen() modes lacking the binary flag, using
// the text flag, or with a misplaced 'b'.
type fopenBinaryUnsafeUsage struct{}

func (fopenBinaryUnsafeUsage) ID() string               { return "FopenBinaryUnsafeUsage" }
func (fopenBinaryUnsafeUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (fopenBinaryUnsafeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "fopen") || astquery.ArgCount(call) < 2 { // D1, D2
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
	m, _, _ := astquery.QuotedStringRaw(lit)
	if m == "" { // D4
		return
	}
	var sev diagnostic.Severity
	var msg string
	hasB := strings.IndexByte(m, 'b') >= 0
	switch { // D5
	case hasB && !strings.HasSuffix(m, "b") && !strings.HasSuffix(m, "b+"):
		sev, msg = diagnostic.SeverityError, "Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+')."
	case hasB:
		return
	case !ctx.Bool("ENFORCE_BINARY_MODIFIER_USAGE"):
		return
	case strings.IndexByte(m, 't') >= 0:
		sev, msg = diagnostic.SeverityWarning, "Use the 'b' flag instead of 't' for binary-safe file access."
	default:
		sev, msg = diagnostic.SeverityWarning, "Add the 'b' flag to the mode for binary-safe file access."
	}
	litSpan := lit.Span()
	repl := "'" + fopenFixMode(m) + "'"
	ctx.ReportSeverity(arg.Value.Span(), sev, msg, diagnostic.Fix{
		Title: "Use mode " + repl,
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: litSpan, NewText: repl}} },
	})
}

// fopenFixMode implements F1.
func fopenFixMode(m string) string {
	r := strings.ReplaceAll(m, "b", "")
	r = strings.ReplaceAll(r, "t", "b")
	if strings.IndexByte(r, 'b') < 0 {
		if strings.IndexByte(r, '+') >= 0 {
			r = strings.ReplaceAll(r, "+", "b+")
		} else {
			r += "b"
		}
	}
	return r
}
