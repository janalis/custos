package compatibility

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// fopenBinaryUnsafeUsage reports fopen() modes lacking the binary flag, using
// the text flag, or with a misplaced 'b'.
type fopenBinaryUnsafeUsage struct{}

func init() { register(fopenBinaryUnsafeUsage{}) }

func (fopenBinaryUnsafeUsage) ID() string { return "FopenBinaryUnsafeUsage" }

func (fopenBinaryUnsafeUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (fopenBinaryUnsafeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "fopen") || util.ArgCount(call) < 2 { // D1, D2
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
	m, _, _ := util.QuotedStringRaw(lit)
	if m == "" { // D4
		return
	}
	var sev meta.Severity
	var msg string
	hasB := strings.IndexByte(m, 'b') >= 0
	switch { // D5
	case hasB && !strings.HasSuffix(m, "b") && !strings.HasSuffix(m, "b+"):
		sev, msg = meta.SeverityError, "Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+')."
	case hasB:
		return
	case !ctx.Bool("ENFORCE_BINARY_MODIFIER_USAGE"):
		return
	case strings.IndexByte(m, 't') >= 0:
		sev, msg = meta.SeverityWarning, "Use the 'b' flag instead of 't' for binary-safe file access."
	default:
		sev, msg = meta.SeverityWarning, "Add the 'b' flag to the mode for binary-safe file access."
	}
	litSpan := lit.Span()
	repl := "'" + fopenFixMode(m) + "'"
	ctx.ReportSeverity(arg.Value.Span(), sev, msg, analysis.Fix{
		Title: "Use mode " + repl,
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: litSpan, NewText: repl}} },
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
