package nestedpositiveifstatements

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// nestedPositiveIfStatements reports an if that is the only statement of a
// parent if (mergeable with &&) or of an else block (becomes else if).
type nestedPositiveIfStatements struct{}

func (nestedPositiveIfStatements) ID() string               { return "NestedPositiveIfStatements" }
func (nestedPositiveIfStatements) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }

const nestedPositiveIfMsg = "Merge this if statement into its parent construct."

func (nestedPositiveIfStatements) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.If)
	block, ok := c.Parent().(*syntax.Block)
	if !ok || block.Alt || len(block.Stmts) != 1 || c.Cond == nil { // D3, D7, E2, E6
		return
	}
	kw, _ := astquery.TokenAfter(ctx.File, c.Span().Start) // the span starts at `if`
	kwSpan := syntax.Span{Start: kw.Start, End: kw.End}
	f := ctx.File
	switch p := block.Parent().(type) {
	case *syntax.If: // Case A (D1: a block child of an if is its body)
		if npiIsOr(p.Cond) || npiIsOr(c.Cond) { // D2, E3
			return
		}
		if len(p.ElseIfs) > 0 || len(c.ElseIfs) > 0 { // D4, E5
			return
		}
		if !npiElseCompatible(f, p.Else, c.Else) { // D5, E4
			return
		}
		if body, ok := c.Body.(*syntax.Block); ok && body.Alt { // E6: the body text has no endif
			return
		}
		if npiLosesComments(f, block, c, c.Cond.Span(), c.Body.Span()) || npiMovesDirective(f, block, c) { // F3
			ctx.Report(kwSpan, nestedPositiveIfMsg)
			return
		}
		ctx.Report(kwSpan, nestedPositiveIfMsg, diagnostic.Fix{
			Title: "Merge into the parent if",
			Edits: func() []diagnostic.TextEdit {
				pc := string(f.Src[p.Cond.Span().Start:p.Cond.Span().End])
				if npiWrapParent(p.Cond) {
					pc = "(" + pc + ")"
				}
				cc := string(f.Src[c.Cond.Span().Start:c.Cond.Span().End])
				if npiWrapChild(c.Cond) {
					cc = "(" + cc + ")"
				}
				body := npiWithComments(f, block, c, c.Body)
				return []diagnostic.TextEdit{
					{Span: p.Cond.Span(), NewText: pc + " && " + cc},
					{Span: block.Span(), NewText: body},
				}
			},
		})
	case *syntax.Else: // Case B (D6: the block is the else body)
		if npiLosesComments(f, block, c, c.Span()) || npiMovesDirective(f, block, c) { // F3
			ctx.Report(kwSpan, nestedPositiveIfMsg)
			return
		}
		ctx.Report(kwSpan, nestedPositiveIfMsg, diagnostic.Fix{
			Title: "Turn into else if",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: block.Span(), NewText: npiWithComments(f, block, c, c)}}
			},
		})
	}
}

func npiIsOr(e syntax.Expr) bool {
	b, ok := e.(*syntax.Binary)
	return ok && (b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr)
}

// npiWrapParent reports whether a parent condition must be parenthesised
// before `&& CC` is appended: its right edge is open and binds looser than
// `&&`, so it would absorb the appended operand.
func npiWrapParent(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Assign, *syntax.Ternary, *syntax.Instanceof, *syntax.Print, *syntax.Include,
		*syntax.Throw, *syntax.Yield, *syntax.YieldFrom, *syntax.ArrowFunction:
		return true
	case *syntax.Binary:
		return e.Op.Kind != syntax.TBooleanAnd || npiWrapParent(e.Right)
	case *syntax.Unary: // `!$row = load()`, `@$u = $v`, `(bool) $u = $v`
		return npiWrapParent(e.Expr)
	case *syntax.Clone:
		return e.Args == nil && npiWrapParent(e.Expr)
	}
	return false
}

// npiWrapChild reports whether a child condition binds looser than the `&&`
// it becomes the right operand of (`||`/`or` never reach the fix, `and`
// keeps its meaning unwrapped).
func npiWrapChild(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Ternary:
		return true
	case *syntax.Binary:
		return e.Op.Kind == syntax.TCoalesce || e.Op.Kind == syntax.TXor
	}
	return false
}

// npiElseCompatible implements D5.
func npiElseCompatible(f *syntax.File, pe, ce *syntax.Else) bool {
	if ce == nil {
		return pe == nil
	}
	if pe == nil {
		return false
	}
	pb, ok1 := pe.Body.(*syntax.Block)
	cb, ok2 := ce.Body.(*syntax.Block)
	if !ok1 || !ok2 || pb.Alt || cb.Alt || len(pb.Stmts) != len(cb.Stmts) {
		return false
	}
	return npiSameTokens(f, pb.Span(), cb.Span())
}

// npiSameTokens compares the significant token texts of two spans.
func npiSameTokens(f *syntax.File, a, b syntax.Span) bool {
	i, j := astquery.TokenIndex(f, a.Start), astquery.TokenIndex(f, b.Start)
	next := func(k int, s syntax.Span) (int, bool) {
		for ; k < len(f.Tokens) && f.Tokens[k].End <= s.End; k++ {
			switch f.Tokens[k].Kind {
			case syntax.TWhitespace, syntax.TComment, syntax.TDocComment:
				continue
			}
			return k, true
		}
		return k, false
	}
	for {
		var okA, okB bool
		i, okA = next(i, a)
		j, okB = next(j, b)
		if !okA || !okB {
			return okA == okB
		}
		ta, tb := f.Tokens[i], f.Tokens[j]
		if string(f.Src[ta.Start:ta.End]) != string(f.Src[tb.Start:tb.End]) {
			return false
		}
		i++
		j++
	}
}

// npiWithComments returns the text of n where comments lying in block before
// c are inserted right after the opening brace of c's body (when braced).
// npiDirectives mark comments that act on the line after them.
var npiDirectives = []string{"@phpstan-ignore", "@psalm-suppress", "phpcs:ignore", "@noinspection", "@custos-ignore"}

// npiMovesDirective reports whether a comment the fix moves from before the
// inner if c into its body is a directive for the line after it: moved, it
// would act on another statement.
func npiMovesDirective(f *syntax.File, block *syntax.Block, c *syntax.If) bool {
	for k := astquery.TokenIndex(f, block.Span().Start); k < len(f.Tokens) && f.Tokens[k].End <= c.Span().Start; k++ {
		t := f.Tokens[k]
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment {
			continue
		}
		for _, d := range npiDirectives {
			if strings.Contains(string(f.Src[t.Start:t.End]), d) {
				return true
			}
		}
	}
	return false
}

// npiLosesComments reports whether block holds a comment the fix would
// drop: one neither before the inner if c with a braced body (moved into
// it) nor inside a kept span.
func npiLosesComments(f *syntax.File, block *syntax.Block, c *syntax.If, kept ...syntax.Span) bool {
	body, braced := c.Body.(*syntax.Block)
	braced = braced && !body.Alt && body.Span().Len() > 0 && f.Src[body.Span().Start] == '{'
	bs := block.Span()
	for k := astquery.TokenIndex(f, bs.Start); k < len(f.Tokens) && f.Tokens[k].Start < bs.End; k++ {
		t := f.Tokens[k]
		if t.Kind != syntax.TComment && t.Kind != syntax.TDocComment || braced && t.End <= c.Span().Start {
			continue
		}
		in := false
		for _, s := range kept {
			in = in || t.Start >= s.Start && t.End <= s.End
		}
		if !in {
			return true
		}
	}
	return false
}

func npiWithComments(f *syntax.File, block *syntax.Block, c *syntax.If, n syntax.Node) string {
	ns := n.Span()
	text := string(f.Src[ns.Start:ns.End])
	body, ok := c.Body.(*syntax.Block)
	if !ok || body.Alt || body.Span().Len() == 0 || f.Src[body.Span().Start] != '{' {
		return text
	}
	var comments []string
	for k := astquery.TokenIndex(f, block.Span().Start); k < len(f.Tokens); k++ {
		t := f.Tokens[k]
		if t.End > c.Span().Start {
			break
		}
		if t.Kind == syntax.TComment || t.Kind == syntax.TDocComment {
			comments = append(comments, string(f.Src[t.Start:t.End]))
		}
	}
	if len(comments) == 0 {
		return text
	}
	at := body.Span().Start + 1 - ns.Start
	indent := astquery.IndentBefore(f.Src, c.Span().Start)
	var b strings.Builder
	b.WriteString(text[:at])
	for _, cm := range comments {
		b.WriteString("\n" + indent + cm)
	}
	rest := text[at:]
	if t := strings.TrimLeft(rest, " \t"); !strings.HasPrefix(t, "\n") && !strings.HasPrefix(t, "\r") {
		b.WriteString("\n" + indent)
	}
	b.WriteString(rest)
	return b.String()
}
