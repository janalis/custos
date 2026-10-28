package fopenbinaryunsafeusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
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
	mode := fopenModeArgument(call)
	if mode == nil {
		return
	}
	lit := semanticquery.LocalStringLiteral(ctx.Types(), mode) // D3
	if lit == nil {
		return
	}
	m, quote, _ := astquery.QuotedStringRaw(lit)
	if m == "" { // D4
		return
	}
	// D4: the raw text is not the mode when it holds an escape, and a quote
	// inside a double-quoted literal cannot be single-quoted as is.
	if strings.IndexByte(m, '\\') >= 0 || (quote == '"' && strings.IndexByte(m, '\'') >= 0) {
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
	case streamWrapper(call) || paramDefault(lit): // E8, E9
		return
	case strings.IndexByte(m, 't') >= 0:
		sev, msg = diagnostic.SeverityWarning, "Use the 'b' flag instead of 't' for binary-safe file access."
	default:
		sev, msg = diagnostic.SeverityWarning, "Add the 'b' flag to the mode for binary-safe file access."
	}
	litSpan := lit.Span()
	repl := "'" + fopenFixMode(m) + "'"
	ctx.ReportSeverity(mode.Span(), sev, msg, diagnostic.Fix{
		Title: "Use mode " + repl,
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: litSpan, NewText: repl}} },
	})
}

// fopenModeArgument returns the expression bound to fopen()'s `mode`
// parameter (D2): the argument named `mode`, else the second positional one.
// nil when a spread or a placeholder makes the binding unknown.
func fopenModeArgument(call *syntax.FuncCall) syntax.Expr {
	for i, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Value == nil || arg.Unpack {
			return nil
		}
		if arg.Name != nil {
			if arg.Name.Value == "mode" {
				return arg.Value
			}
			continue
		}
		if i == 1 {
			return arg.Value
		}
	}
	return nil
}

// streamWrapper reports whether fopen call names, by a literal or a
// concatenation starting with one, a stream other than a plain file
// (`php://memory`, `php://fd/3`, a registered wrapper) or a Unix device
// (`/dev/null`): the binary flag only matters for files.
func streamWrapper(call *syntax.FuncCall) bool {
	e := syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 0, "filename"))
	for {
		b, ok := e.(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TDot {
			break
		}
		e = syntax.UnwrapParens(b.Left)
	}
	lit, ok := e.(*syntax.Literal)
	if !ok {
		return false
	}
	name, ok := syntax.LiteralStringValue(lit)
	if !ok {
		return false
	}
	if strings.HasPrefix(name, "/dev/") {
		return true
	}
	scheme, _, found := strings.Cut(name, "://")
	return found && scheme != "" && !strings.EqualFold(scheme, "file")
}

// paramDefault reports whether lit is the default of a parameter: callers
// choose the mode (E9).
func paramDefault(lit syntax.Expr) bool {
	_, ok := lit.Parent().(*syntax.Param)
	return ok
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
