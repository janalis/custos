package shortechotagcanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// shortEchoTagCanBeUsed reports `<?php echo $x ?>` blocks that can use the
// short echo tag `<?= $x ?>`.
type shortEchoTagCanBeUsed struct{}

const shortEchoTagCanBeUsedMsg = "Use the short echo tag '<?= … ?>' here."

func (shortEchoTagCanBeUsed) ID() string { return "ShortEchoTagCanBeUsed" }
func (shortEchoTagCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KEcho, syntax.KExprStmt}
}

func (shortEchoTagCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	var args []syntax.Expr
	switch n := n.(type) {
	case *syntax.Echo:
		if n.Short || len(n.Exprs) == 0 { // E2
			return
		}
		args = n.Exprs
	case *syntax.ExprStmt:
		p, ok := n.Expr.(*syntax.Print) // D1 / E5
		if !ok || p.Expr == nil {
			return
		}
		args = []syntax.Expr{p.Expr}
	}
	f := ctx.File
	start := n.Span().Start
	kw, _ := astquery.TokenAfter(f, start) // the statement starts at echo/print
	// D2: a regular opening tag right before, only whitespace in between.
	open := shortEchoPrevTok(f, start)
	if open.Kind != syntax.TOpenTag {
		return
	}
	// D3: optional `;`, then `?>`, only whitespace in between.
	end := args[len(args)-1].Span().End
	next, ok := shortEchoNextTok(f, end)
	if ok && next.Kind == syntax.TSemicolon {
		end = next.End
		next, ok = shortEchoNextTok(f, end)
	}
	if !ok || next.Kind != syntax.TCloseTag {
		return
	}
	stmt := syntax.Span{Start: start, End: end}
	openSpan := syntax.Span{Start: open.Start, End: open.End}
	ctx.Report(syntax.Span{Start: kw.Start, End: kw.End}, shortEchoTagCanBeUsedMsg, diagnostic.Fix{
		Title: "Use the short echo tag",
		Edits: func() []diagnostic.TextEdit {
			texts := make([]string, len(args))
			for i, a := range args {
				texts[i] = ctx.Text(a)
			}
			// The open tag token may carry one trailing whitespace char; keep it.
			tagText := ctx.SpanText(openSpan)
			trail := tagText[len(strings.TrimRight(tagText, " \t\r\n")):]
			return []diagnostic.TextEdit{
				{Span: openSpan, NewText: "<?=" + trail},
				{Span: stmt, NewText: strings.Join(texts, ", ")},
			}
		},
	})
}

// shortEchoPrevTok returns the token before off, skipping whitespace only.
// (A statement is always preceded by at least its file's open tag.)
func shortEchoPrevTok(f *syntax.File, off uint32) syntax.Token {
	i := max(astquery.TokenIndex(f, off)-1, 0)
	for i > 0 && f.Tokens[i].Kind == syntax.TWhitespace {
		i--
	}
	return f.Tokens[i]
}

// shortEchoNextTok returns the token starting at or after off, skipping
// whitespace only.
func shortEchoNextTok(f *syntax.File, off uint32) (syntax.Token, bool) {
	for i := astquery.TokenIndex(f, off); i < len(f.Tokens); i++ {
		if t := f.Tokens[i]; t.Kind != syntax.TWhitespace {
			return t, t.Kind != syntax.TEOF
		}
	}
	return syntax.Token{}, false
}
