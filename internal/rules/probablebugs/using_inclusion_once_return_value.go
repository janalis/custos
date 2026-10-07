package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// usingInclusionOnceReturnValue reports include_once/require_once whose
// result is used as a value.
type usingInclusionOnceReturnValue struct{}

func init() { register(usingInclusionOnceReturnValue{}) }

const usingInclusionOnceReturnValueMsg = "Only the first include_once/require_once returns the file's value; later ones return true."

func (usingInclusionOnceReturnValue) ID() string { return "UsingInclusionOnceReturnValue" }

func (usingInclusionOnceReturnValue) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInclude}
}

func (usingInclusionOnceReturnValue) Check(ctx *analysis.Context, n syntax.Node) {
	inc := n.(*syntax.Include)
	kw := strings.ToLower(ctx.SpanText(inc.Keyword.Span))                                 // keywords are case-insensitive
	if !strings.HasSuffix(kw, "_once") || inc.Expr == nil || inc.Expr.Span().Len() == 0 { // D1
		return
	}
	if es, ok := inc.Parent().(*syntax.ExprStmt); ok && es.Expr == syntax.Expr(inc) { // D2
		return
	}
	span := syntax.Span{Start: inc.Keyword.Span.Start, End: inc.Expr.Span().End}
	repl := strings.TrimSuffix(kw, "_once") + " "
	edit := syntax.Span{Start: inc.Keyword.Span.Start, End: inc.Expr.Span().Start}
	ctx.Report(span, usingInclusionOnceReturnValueMsg, analysis.Fix{
		Title: "Use " + strings.TrimSpace(repl),
		Edits: func() []analysis.TextEdit {
			// Only the keyword (and what precedes the operand) is replaced,
			// so nested inclusions inside the operand can be fixed too.
			return []analysis.TextEdit{{Span: edit, NewText: repl}}
		},
	})
}
