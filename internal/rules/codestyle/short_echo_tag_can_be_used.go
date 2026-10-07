package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// shortEchoTagCanBeUsed reports `<?php echo $x ?>` blocks that can use the
// short echo tag `<?= $x ?>`.
type shortEchoTagCanBeUsed struct{}

func init() { register(shortEchoTagCanBeUsed{}) }

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
	default:
		return
	}
	f := ctx.File
	start := n.Span().Start
	kw, ok := util.TokenAfter(f, start)
	if !ok || (kw.Kind != syntax.TEcho && kw.Kind != syntax.TPrint) {
		return
	}
	// D2: a regular opening tag right before, only whitespace in between.
	open, ok := shortEchoPrevTok(f, start)
	if !ok || open.Kind != syntax.TOpenTag {
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
	ctx.Report(syntax.Span{Start: kw.Start, End: kw.End}, shortEchoTagCanBeUsedMsg, analysis.Fix{
		Title: "Use the short echo tag",
		Edits: func() []analysis.TextEdit {
			texts := make([]string, len(args))
			for i, a := range args {
				texts[i] = ctx.Text(a)
			}
			// The open tag token may carry one trailing whitespace char; keep it.
			tagText := ctx.SpanText(openSpan)
			trail := tagText[len(strings.TrimRight(tagText, " \t\r\n")):]
			return []analysis.TextEdit{
				{Span: openSpan, NewText: "<?=" + trail},
				{Span: stmt, NewText: strings.Join(texts, ", ")},
			}
		},
	})
}

// shortEchoPrevTok returns the token before off, skipping whitespace only.
func shortEchoPrevTok(f *syntax.File, off uint32) (syntax.Token, bool) {
	for i := util.TokenIndex(f, off) - 1; i >= 0; i-- {
		if t := f.Tokens[i]; t.Kind != syntax.TWhitespace {
			return t, true
		}
	}
	return syntax.Token{}, false
}

// shortEchoNextTok returns the token starting at or after off, skipping
// whitespace only.
func shortEchoNextTok(f *syntax.File, off uint32) (syntax.Token, bool) {
	for i := util.TokenIndex(f, off); i < len(f.Tokens); i++ {
		if t := f.Tokens[i]; t.Kind != syntax.TWhitespace {
			return t, t.Kind != syntax.TEOF
		}
	}
	return syntax.Token{}, false
}
